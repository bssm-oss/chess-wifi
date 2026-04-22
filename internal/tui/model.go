package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/session"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen string

const (
	screenMenu    screen = "menu"
	screenHost    screen = "host_form"
	screenJoin    screen = "join_form"
	screenWaiting screen = "waiting"
	screenMatch   screen = "match"
	screenError   screen = "error"
)

const (
	cellWidth  = 5
	cellHeight = 2
	boardX     = 2
	boardY     = 8
)

type hostAcceptedMsg struct{ session *session.PeerSession }
type hostErrorMsg struct{ err error }
type joinResultMsg struct {
	session *session.PeerSession
	err     error
}
type sessionEventMsg struct{ event session.Event }

type promotionChoice struct {
	Target  string
	Options []game.MoveOption
	Index   int
}

type rect struct {
	x int
	y int
	w int
	h int
}

type model struct {
	screen       screen
	menuIndex    int
	hostInputs   []textinput.Model
	joinInputs   []textinput.Model
	focusIndex   int
	listener     *session.HostListener
	peerSession  *session.PeerSession
	snapshot     game.Snapshot
	message      string
	width        int
	height       int
	cursorFile   int
	cursorRank   int
	selected     string
	legalMoves   []game.MoveOption
	promotion    *promotionChoice
	boardBounds  rect
	joining      bool
	waitingSince time.Time
	quitting     bool
}

func newModel() *model {
	hostName := textinput.New()
	hostName.Placeholder = "플레이어 이름"
	hostName.SetValue("Host")
	hostName.Focus()

	hostPort := textinput.New()
	hostPort.Placeholder = "포트"
	hostPort.SetValue(strconv.Itoa(session.DefaultPort))

	joinName := textinput.New()
	joinName.Placeholder = "플레이어 이름"
	joinName.SetValue("Guest")
	joinName.Focus()

	joinAddr := textinput.New()
	joinAddr.Placeholder = "호스트 주소 (예: 192.168.0.10:8787)"

	return &model{
		screen:      screenMenu,
		hostInputs:  []textinput.Model{hostName, hostPort},
		joinInputs:  []textinput.Model{joinName, joinAddr},
		cursorFile:  4,
		cursorRank:  1,
		boardBounds: rect{x: boardX, y: boardY, w: cellWidth * 8, h: cellHeight * 8},
		message:     "같은 Wi-Fi에서 직접 연결되는 체스를 준비하세요.",
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case hostAcceptedMsg:
		m.listener = nil
		m.peerSession = msg.session
		m.snapshot = msg.session.Snapshot()
		m.screen = screenMatch
		m.message = fmt.Sprintf("%s connected. White moves first.", msg.session.Peer().Name)
		return m, waitForSessionEvent(msg.session)
	case hostErrorMsg:
		m.screen = screenError
		m.message = msg.err.Error()
		m.listener = nil
		return m, nil
	case joinResultMsg:
		m.joining = false
		if msg.err != nil {
			m.screen = screenError
			m.message = msg.err.Error()
			return m, nil
		}
		m.peerSession = msg.session
		m.snapshot = msg.session.Snapshot()
		m.screen = screenMatch
		m.message = fmt.Sprintf("Connected to %s.", msg.session.Peer().Name)
		return m, waitForSessionEvent(msg.session)
	case sessionEventMsg:
		switch msg.event.Type {
		case session.EventSnapshot:
			m.snapshot = msg.event.Snapshot
			m.message = msg.event.Snapshot.Message
			m.clearSelection()
		case session.EventError:
			if msg.event.Snapshot.FEN != "" {
				m.snapshot = msg.event.Snapshot
			}
			m.message = msg.event.Message
		case session.EventClosed:
			m.snapshot = msg.event.Snapshot
			m.screen = screenError
			m.message = msg.event.Message
			return m, nil
		}
		if m.peerSession != nil {
			return m, waitForSessionEvent(m.peerSession)
		}
	}
	return m, nil
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.promotion != nil {
		return m.handlePromotionKey(msg)
	}

	switch m.screen {
	case screenMenu:
		switch msg.String() {
		case "up", "k":
			if m.menuIndex > 0 {
				m.menuIndex--
			}
		case "down", "j":
			if m.menuIndex < 1 {
				m.menuIndex++
			}
		case "enter", " ":
			if m.menuIndex == 0 {
				m.focusHostField(0)
				m.screen = screenHost
			} else {
				m.focusJoinField(0)
				m.screen = screenJoin
			}
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case screenHost:
		if msg.String() == "esc" {
			m.screen = screenMenu
			return m, nil
		}
		if msg.String() == "tab" || msg.String() == "shift+tab" || msg.String() == "up" || msg.String() == "down" {
			m.cycleHostFocus(msg.String())
			return m, nil
		}
		if msg.String() == "enter" && m.focusIndex == len(m.hostInputs)-1 {
			return m.startHosting()
		}
		var cmd tea.Cmd
		m.hostInputs[m.focusIndex], cmd = m.hostInputs[m.focusIndex].Update(msg)
		return m, cmd
	case screenJoin:
		if msg.String() == "esc" {
			m.screen = screenMenu
			return m, nil
		}
		if msg.String() == "tab" || msg.String() == "shift+tab" || msg.String() == "up" || msg.String() == "down" {
			m.cycleJoinFocus(msg.String())
			return m, nil
		}
		if msg.String() == "enter" && m.focusIndex == len(m.joinInputs)-1 {
			return m.startJoin()
		}
		var cmd tea.Cmd
		m.joinInputs[m.focusIndex], cmd = m.joinInputs[m.focusIndex].Update(msg)
		return m, cmd
	case screenWaiting:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			if m.listener != nil {
				_ = m.listener.Close()
				m.listener = nil
			}
			m.screen = screenMenu
			m.message = "Hosting cancelled."
		}
		return m, nil
	case screenMatch:
		return m.handleMatchKey(msg)
	case screenError:
		switch msg.String() {
		case "enter", "esc":
			m.screen = screenMenu
			m.message = "새 매치를 시작할 수 있습니다."
			m.peerSession = nil
			m.listener = nil
			m.clearSelection()
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	}
	return m, nil
}

func (m *model) handlePromotionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.promotion.Index > 0 {
			m.promotion.Index--
		}
	case "right", "l":
		if m.promotion.Index < len(m.promotion.Options)-1 {
			m.promotion.Index++
		}
	case "enter", " ":
		move := m.promotion.Options[m.promotion.Index]
		m.promotion = nil
		m.message = "Submitting promotion..."
		return m, m.submitMove(move.UCI)
	case "esc":
		m.promotion = nil
	}
	return m, nil
}

func (m *model) handleMatchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.peerSession != nil {
			_ = m.peerSession.Close()
		}
		m.quitting = true
		return m, tea.Quit
	case "r":
		if m.peerSession != nil {
			m.message = "Resigning match..."
			return m, func() tea.Msg {
				return sessionEventMsg{event: session.Event{Type: session.EventError, Message: resignErr(m.peerSession.Resign())}}
			}
		}
	case "esc":
		m.clearSelection()
	case "left", "h":
		if m.cursorFile > 0 {
			m.cursorFile--
		}
	case "right", "l":
		if m.cursorFile < 7 {
			m.cursorFile++
		}
	case "up", "k":
		if m.cursorRank < 7 {
			m.cursorRank++
		}
	case "down", "j":
		if m.cursorRank > 0 {
			m.cursorRank--
		}
	case "enter", " ":
		sq, err := game.ParseSquareName(m.cursorFile, m.cursorRank)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		return m.activateSquare(sq)
	}
	return m, nil
}

func (m *model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenMatch {
		return m, nil
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if m.promotion != nil {
		if choice := m.promotionFromMouse(msg.X, msg.Y); choice >= 0 {
			move := m.promotion.Options[choice]
			m.promotion = nil
			return m, m.submitMove(move.UCI)
		}
		return m, nil
	}
	sq, ok := m.squareFromMouse(msg.X, msg.Y)
	if !ok {
		return m, nil
	}
	return m.activateSquare(sq)
}

func (m *model) activateSquare(square string) (tea.Model, tea.Cmd) {
	if m.peerSession == nil || m.snapshot.Status != "active" {
		return m, nil
	}
	if m.snapshot.Turn != m.peerSession.Role() {
		m.message = "상대 턴입니다."
		return m, nil
	}
	if m.selected == "" {
		piece, side, ok, err := game.PieceAt(m.snapshot.FEN, square)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		if !ok || side != m.peerSession.Role() || piece == 0 {
			return m, nil
		}
		moves, err := game.LegalMovesForSquare(m.snapshot.FEN, m.peerSession.Role(), square)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		if len(moves) == 0 {
			return m, nil
		}
		m.selected = square
		m.legalMoves = moves
		m.message = fmt.Sprintf("Selected %s", strings.ToUpper(square))
		return m, nil
	}
	if m.selected == square {
		m.clearSelection()
		return m, nil
	}
	var matches []game.MoveOption
	for _, move := range m.legalMoves {
		if move.Target == square {
			matches = append(matches, move)
		}
	}
	if len(matches) == 0 {
		m.clearSelection()
		return m.activateSquare(square)
	}
	if len(matches) == 1 {
		m.clearSelection()
		m.message = "Submitting move..."
		return m, m.submitMove(matches[0].UCI)
	}
	m.promotion = &promotionChoice{Target: square, Options: matches}
	m.message = "Choose promotion piece"
	return m, nil
}

func (m *model) startHosting() (tea.Model, tea.Cmd) {
	port, err := strconv.Atoi(strings.TrimSpace(m.hostInputs[1].Value()))
	if err != nil {
		m.message = "포트는 숫자여야 합니다."
		return m, nil
	}
	listener, err := session.StartHost(strings.TrimSpace(m.hostInputs[0].Value()), port)
	if err != nil {
		m.message = err.Error()
		m.screen = screenError
		return m, nil
	}
	m.listener = listener
	m.screen = screenWaiting
	m.waitingSince = time.Now()
	m.message = "같은 Wi-Fi의 상대에게 주소를 공유하세요."
	return m, waitForHostAccepted(listener)
}

func (m *model) startJoin() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.joinInputs[0].Value())
	addr := strings.TrimSpace(m.joinInputs[1].Value())
	if name == "" || addr == "" {
		m.message = "이름과 주소를 모두 입력하세요."
		return m, nil
	}
	m.joining = true
	m.message = "Connecting..."
	return m, joinCmd(name, addr)
}

func (m *model) submitMove(uci string) tea.Cmd {
	return func() tea.Msg {
		if m.peerSession == nil {
			return sessionEventMsg{event: session.Event{Type: session.EventError, Message: "no session"}}
		}
		if err := m.peerSession.SubmitMove(uci); err != nil {
			return sessionEventMsg{event: session.Event{Type: session.EventError, Snapshot: m.snapshot, Message: err.Error()}}
		}
		return nil
	}
}

func waitForHostAccepted(listener *session.HostListener) tea.Cmd {
	return func() tea.Msg {
		select {
		case peer, ok := <-listener.Accepted():
			if ok && peer != nil {
				return hostAcceptedMsg{session: peer}
			}
		case err, ok := <-listener.Errors():
			if ok && err != nil {
				return hostErrorMsg{err: err}
			}
		}
		return hostErrorMsg{err: fmt.Errorf("host listener closed")}
	}
}

func joinCmd(name, address string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		peer, err := session.Join(ctx, address, name)
		return joinResultMsg{session: peer, err: err}
	}
}

func waitForSessionEvent(peer *session.PeerSession) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-peer.Events()
		if !ok {
			return sessionEventMsg{event: session.Event{Type: session.EventClosed, Snapshot: peer.Snapshot(), Message: "connection closed"}}
		}
		return sessionEventMsg{event: event}
	}
}

func (m *model) clearSelection() {
	m.selected = ""
	m.legalMoves = nil
	m.promotion = nil
}

func (m *model) focusHostField(idx int) {
	m.focusIndex = idx
	for i := range m.hostInputs {
		if i == idx {
			m.hostInputs[i].Focus()
		} else {
			m.hostInputs[i].Blur()
		}
	}
}

func (m *model) focusJoinField(idx int) {
	m.focusIndex = idx
	for i := range m.joinInputs {
		if i == idx {
			m.joinInputs[i].Focus()
		} else {
			m.joinInputs[i].Blur()
		}
	}
}

func (m *model) cycleHostFocus(key string) {
	if key == "shift+tab" || key == "up" {
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.hostInputs) - 1
		}
	} else {
		m.focusIndex = (m.focusIndex + 1) % len(m.hostInputs)
	}
	m.focusHostField(m.focusIndex)
}

func (m *model) cycleJoinFocus(key string) {
	if key == "shift+tab" || key == "up" {
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.joinInputs) - 1
		}
	} else {
		m.focusIndex = (m.focusIndex + 1) % len(m.joinInputs)
	}
	m.focusJoinField(m.focusIndex)
}

func resignErr(err error) string {
	if err != nil {
		return err.Error()
	}
	return "Resignation sent."
}
