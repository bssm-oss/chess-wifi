package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bssm-oss/chess-wifi/internal/discovery"
	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/lan"
	"github.com/bssm-oss/chess-wifi/internal/netproto"
)

const (
	DefaultPort      = 8787
	HeartbeatEvery   = 5 * time.Second
	ReadTimeout      = 15 * time.Second
	WriteTimeout     = 10 * time.Second
	defaultGuestName = "Guest"
)

type HostListener struct {
	listener     net.Listener
	Addresses    []string
	accepted     chan *PeerSession
	errs         chan error
	closeOnce    sync.Once
	stopAnnounce func()
}

func StartHost(name string, port int) (*HostListener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listen on %d: %w", port, err)
	}
	addresses, err := lan.PrivateIPv4(port)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	h := &HostListener{
		listener:  listener,
		Addresses: addresses,
		accepted:  make(chan *PeerSession, 1),
		errs:      make(chan error, 1),
	}
	if stopAnnounce, err := discovery.StartAnnouncer(context.Background(), name, port); err == nil {
		h.stopAnnounce = stopAnnounce
	}
	go h.acceptLoop(name)
	return h, nil
}

func (h *HostListener) Accepted() <-chan *PeerSession { return h.accepted }
func (h *HostListener) Errors() <-chan error          { return h.errs }

func (h *HostListener) Close() error {
	var err error
	h.closeOnce.Do(func() {
		if h.stopAnnounce != nil {
			h.stopAnnounce()
		}
		err = h.listener.Close()
	})
	return err
}

func (h *HostListener) acceptLoop(name string) {
	defer close(h.accepted)
	defer close(h.errs)
	defer h.Close()
	conn, err := h.listener.Accept()
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			h.errs <- err
		}
		return
	}
	peer, err := newHostSession(conn, name)
	if err != nil {
		h.errs <- err
		_ = conn.Close()
		return
	}
	h.accepted <- peer
}

func Join(ctx context.Context, address, name string) (*PeerSession, error) {
	dialer := net.Dialer{Timeout: WriteTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial host: %w", err)
	}
	peer, err := newClientSession(conn, name)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return peer, nil
}

type EventType string

const (
	EventSnapshot EventType = "snapshot"
	EventError    EventType = "error"
	EventClosed   EventType = "closed"
)

type Event struct {
	Type     EventType
	Snapshot game.Snapshot
	Message  string
}

type PeerSession struct {
	mu         sync.RWMutex
	conn       net.Conn
	codec      *netproto.Codec
	self       game.Player
	peer       game.Player
	role       game.Side
	mode       string
	match      *game.Match
	state      game.Snapshot
	events     chan Event
	closed     chan struct{}
	closeOnce  sync.Once
	eventsOnce sync.Once
	workers    sync.WaitGroup
	heartbeats *time.Ticker
}

func (p *PeerSession) Self() game.Player    { return p.self }
func (p *PeerSession) Peer() game.Player    { return p.peer }
func (p *PeerSession) Role() game.Side      { return p.role }
func (p *PeerSession) Events() <-chan Event { return p.events }

func (p *PeerSession) Snapshot() game.Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *PeerSession) SubmitMove(uci string) error {
	if p.role != p.state.Turn || p.state.Status != "active" {
		return errors.New("it is not your turn")
	}
	if p.mode == "host" {
		return p.applyHostMove(uci)
	}
	return p.writeMessage(netproto.TypeMoveIntent, netproto.MoveIntent{ExpectedVersion: p.Snapshot().Version, MoveUCI: uci})
}

func (p *PeerSession) Resign() error {
	if p.mode == "host" {
		snap := p.match.Resign(p.role)
		p.setSnapshot(snap)
		p.emit(Event{Type: EventSnapshot, Snapshot: snap})
		return p.writeMessage(netproto.TypeSnapshot, netproto.Snapshot{State: snap})
	}
	return p.writeMessage(netproto.TypeAction, netproto.ActionIntent{ExpectedVersion: p.Snapshot().Version, Action: "resign"})
}

func (p *PeerSession) Close() error {
	var err error
	p.closeOnce.Do(func() {
		close(p.closed)
		if p.heartbeats != nil {
			p.heartbeats.Stop()
		}
		err = p.conn.Close()
		p.eventsOnce.Do(func() {
			go func() {
				p.workers.Wait()
				close(p.events)
			}()
		})
	})
	return err
}

func newHostSession(conn net.Conn, hostName string) (*PeerSession, error) {
	codec := netproto.NewCodec(conn, conn)
	if err := conn.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
		return nil, err
	}
	env, err := codec.Read()
	if err != nil {
		return nil, err
	}
	if env.Type != netproto.TypeHello {
		return nil, fmt.Errorf("expected hello, got %s", env.Type)
	}
	hello, err := netproto.DecodePayload[netproto.Hello](env)
	if err != nil {
		return nil, err
	}
	if hello.ProtocolVersion != netproto.ProtocolVersion {
		return nil, fmt.Errorf("protocol mismatch: %s", hello.ProtocolVersion)
	}
	guestName := hello.PlayerName
	if guestName == "" {
		guestName = defaultGuestName
	}
	match := game.New(hostName, guestName)
	self := game.Player{Name: hostName, Side: game.White}
	peer := game.Player{Name: guestName, Side: game.Black}
	if err := writeEnvelope(conn, codec, netproto.TypeWelcome, netproto.Welcome{ProtocolVersion: netproto.ProtocolVersion, Self: self, Peer: peer}); err != nil {
		return nil, err
	}
	state := match.Snapshot()
	if err := writeEnvelope(conn, codec, netproto.TypeSnapshot, netproto.Snapshot{State: state}); err != nil {
		return nil, err
	}
	p := &PeerSession{
		conn:       conn,
		codec:      codec,
		self:       self,
		peer:       peer,
		role:       game.White,
		mode:       "host",
		match:      match,
		state:      state,
		events:     make(chan Event, 8),
		closed:     make(chan struct{}),
		heartbeats: time.NewTicker(HeartbeatEvery),
	}
	p.workers.Add(2)
	go p.readLoop()
	go p.pingLoop()
	return p, nil
}

func newClientSession(conn net.Conn, name string) (*PeerSession, error) {
	codec := netproto.NewCodec(conn, conn)
	if err := writeEnvelope(conn, codec, netproto.TypeHello, netproto.Hello{ProtocolVersion: netproto.ProtocolVersion, PlayerName: name}); err != nil {
		return nil, err
	}
	if err := conn.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
		return nil, err
	}
	env, err := codec.Read()
	if err != nil {
		return nil, err
	}
	if env.Type != netproto.TypeWelcome {
		return nil, fmt.Errorf("expected welcome, got %s", env.Type)
	}
	welcome, err := netproto.DecodePayload[netproto.Welcome](env)
	if err != nil {
		return nil, err
	}
	if welcome.ProtocolVersion != netproto.ProtocolVersion {
		return nil, fmt.Errorf("protocol mismatch: %s", welcome.ProtocolVersion)
	}
	env, err = codec.Read()
	if err != nil {
		return nil, err
	}
	if env.Type != netproto.TypeSnapshot {
		return nil, fmt.Errorf("expected snapshot, got %s", env.Type)
	}
	snapMsg, err := netproto.DecodePayload[netproto.Snapshot](env)
	if err != nil {
		return nil, err
	}
	p := &PeerSession{
		conn:       conn,
		codec:      codec,
		self:       welcome.Peer,
		peer:       welcome.Self,
		role:       welcome.Peer.Side,
		mode:       "client",
		state:      snapMsg.State,
		events:     make(chan Event, 8),
		closed:     make(chan struct{}),
		heartbeats: time.NewTicker(HeartbeatEvery),
	}
	p.workers.Add(2)
	go p.readLoop()
	go p.pingLoop()
	return p, nil
}

func (p *PeerSession) applyHostMove(uci string) error {
	snap, err := p.match.ApplyMoveUCI(uci)
	if err != nil {
		return err
	}
	p.setSnapshot(snap)
	p.emit(Event{Type: EventSnapshot, Snapshot: snap})
	return p.writeMessage(netproto.TypeSnapshot, netproto.Snapshot{State: snap})
}

func (p *PeerSession) readLoop() {
	defer p.workers.Done()
	for {
		select {
		case <-p.closed:
			return
		default:
		}
		if err := p.conn.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
			p.emitClose(err)
			return
		}
		env, err := p.codec.Read()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				p.emitClose(errors.New("connection closed"))
				return
			}
			p.emitClose(err)
			return
		}
		switch env.Type {
		case netproto.TypeSnapshot:
			snap, err := netproto.DecodePayload[netproto.Snapshot](env)
			if err != nil {
				p.emit(Event{Type: EventError, Message: err.Error()})
				continue
			}
			p.setSnapshot(snap.State)
			p.emit(Event{Type: EventSnapshot, Snapshot: snap.State})
		case netproto.TypeError:
			msg, err := netproto.DecodePayload[netproto.Error](env)
			if err != nil {
				p.emit(Event{Type: EventError, Message: err.Error()})
				continue
			}
			if msg.State.FEN != "" {
				p.setSnapshot(msg.State)
			}
			p.emit(Event{Type: EventError, Message: msg.Message, Snapshot: msg.State})
		case netproto.TypeMoveIntent:
			if p.mode != "host" {
				continue
			}
			move, err := netproto.DecodePayload[netproto.MoveIntent](env)
			if err != nil {
				_ = p.writeMessage(netproto.TypeError, netproto.Error{Message: err.Error(), State: p.Snapshot()})
				continue
			}
			if move.ExpectedVersion != p.Snapshot().Version {
				_ = p.writeMessage(netproto.TypeError, netproto.Error{Message: "stale board state", State: p.Snapshot()})
				continue
			}
			if err := p.applyHostMove(move.MoveUCI); err != nil {
				_ = p.writeMessage(netproto.TypeError, netproto.Error{Message: err.Error(), State: p.Snapshot()})
			}
		case netproto.TypeAction:
			action, err := netproto.DecodePayload[netproto.ActionIntent](env)
			if err != nil {
				_ = p.writeMessage(netproto.TypeError, netproto.Error{Message: err.Error(), State: p.Snapshot()})
				continue
			}
			if p.mode == "host" && action.Action == "resign" {
				snap := p.match.Resign(game.Black)
				p.setSnapshot(snap)
				p.emit(Event{Type: EventSnapshot, Snapshot: snap})
				_ = p.writeMessage(netproto.TypeSnapshot, netproto.Snapshot{State: snap})
			}
		case netproto.TypePing:
			continue
		}
	}
}

func (p *PeerSession) pingLoop() {
	defer p.workers.Done()
	for {
		select {
		case <-p.closed:
			return
		case <-p.heartbeats.C:
			if err := p.writeMessage(netproto.TypePing, netproto.Ping{}); err != nil {
				p.emitClose(err)
				return
			}
		}
	}
}

func (p *PeerSession) setSnapshot(snapshot game.Snapshot) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = snapshot
}

func (p *PeerSession) emit(event Event) {
	select {
	case <-p.closed:
		if event.Type != EventClosed {
			return
		}
	default:
	}
	select {
	case p.events <- event:
	default:
	}
}

func (p *PeerSession) emitClose(err error) {
	msg := "connection interrupted"
	if err != nil {
		msg = err.Error()
	}
	p.emit(Event{Type: EventClosed, Snapshot: p.Snapshot(), Message: msg})
	_ = p.Close()
}

func (p *PeerSession) writeMessage(msgType string, payload any) error {
	return writeEnvelope(p.conn, p.codec, msgType, payload)
}

func writeEnvelope(conn net.Conn, codec *netproto.Codec, msgType string, payload any) error {
	if err := conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	return codec.Write(netproto.Envelope{Type: msgType, Payload: payload})
}
