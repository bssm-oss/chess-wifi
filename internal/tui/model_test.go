package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/bssm-oss/chess-wifi/internal/discovery"
	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHostErrorIgnoredAfterCancel(t *testing.T) {
	m := newModel()
	listener, err := session.StartHost("Host", 0)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	m.listener = listener
	m.screen = screenWaiting

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(*model)
	if m2.screen != screenMenu {
		t.Fatalf("expected menu after cancel, got %s", m2.screen)
	}

	updated, _ = m2.Update(hostErrorMsg{listener: listener, err: errString("host listener closed")})
	m2 = updated.(*model)
	if m2.screen != screenMenu {
		t.Fatalf("expected stale host error to be ignored, got %s", m2.screen)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestMenuShowsDiscoveredMatches(t *testing.T) {
	m := newModel()
	m.discoveryRun = false
	m.discoveries = []discovery.Match{
		{PlayerName: "Host", Address: "127.0.0.1:8787", LastSeen: time.Now()},
	}
	m.menuIndex = 2

	view := m.renderMenu()
	if !strings.Contains(view, "열려있는 LAN 매치") {
		t.Fatal("expected discovery section in menu")
	}
	if !strings.Contains(view, "Join Host · 127.0.0.1:8787") {
		t.Fatalf("expected discovered match in menu choices, got %q", view)
	}
}

func TestEnterOnDiscoveredMatchStartsJoin(t *testing.T) {
	hostListener, err := session.StartHost("Host", 9104)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = hostListener.Close() })

	acceptedCh := make(chan *session.PeerSession, 1)
	go func() {
		acceptedCh <- <-hostListener.Accepted()
	}()

	m := newModel()
	m.discoveryRun = false
	m.discoveries = []discovery.Match{
		{PlayerName: "Host", Address: "127.0.0.1:9104", LastSeen: time.Now()},
	}
	m.menuIndex = 2

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(*model)
	if !m2.joining {
		t.Fatal("expected model to enter joining state")
	}
	if m2.joinInputs[1].Value() != "127.0.0.1:9104" {
		t.Fatalf("expected discovered address to be copied, got %q", m2.joinInputs[1].Value())
	}
	if cmd == nil {
		t.Fatal("expected join command")
	}

	msg := cmd()
	result, ok := msg.(joinResultMsg)
	if !ok {
		t.Fatalf("expected joinResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("join command returned error: %v", result.err)
	}
	t.Cleanup(func() { _ = result.session.Close() })

	select {
	case host := <-acceptedCh:
		t.Cleanup(func() { _ = host.Close() })
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for host acceptance")
	}
}

func TestMouseClickOnMenuStartsDiscoveredJoin(t *testing.T) {
	hostListener, err := session.StartHost("Host", 9105)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = hostListener.Close() })

	acceptedCh := make(chan *session.PeerSession, 1)
	go func() {
		acceptedCh <- <-hostListener.Accepted()
	}()

	m := newModel()
	m.discoveryRun = false
	m.discoveries = []discovery.Match{
		{PlayerName: "Host", Address: "127.0.0.1:9105", LastSeen: time.Now()},
	}
	_ = m.renderMenu()

	bounds := m.menuBounds[2]
	updated, cmd := m.handleMouse(tea.MouseMsg{X: bounds.x + 1, Y: bounds.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m2 := updated.(*model)
	if !m2.joining {
		t.Fatal("expected mouse click to start joining")
	}
	if cmd == nil {
		t.Fatal("expected join command")
	}
	msg := cmd()
	result, ok := msg.(joinResultMsg)
	if !ok {
		t.Fatalf("expected joinResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("join command returned error: %v", result.err)
	}
	t.Cleanup(func() { _ = result.session.Close() })

	select {
	case host := <-acceptedCh:
		t.Cleanup(func() { _ = host.Close() })
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for host acceptance")
	}
}

func TestWaitingAddressClickReturnsCopyCommand(t *testing.T) {
	m := newModel()
	m.screen = screenWaiting
	m.listener = &session.HostListener{Addresses: []string{"127.0.0.1:8787"}}
	_ = m.renderWaiting()

	bounds := m.waitingCopyBounds[0]
	_, cmd := m.handleMouse(tea.MouseMsg{X: bounds.x + 1, Y: bounds.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if cmd == nil {
		t.Fatal("expected copy command")
	}
}

func TestCompactMatchViewFitsDefaultTerminalHeight(t *testing.T) {
	m := newModel()
	m.screen = screenMatch
	m.width = 80
	m.height = 24
	m.viewSide = game.White
	m.peerSession = &session.PeerSession{}
	m.snapshot = game.Snapshot{
		FEN:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Turn:        game.White,
		Status:      "active",
		MoveHistory: []string{"e2e4", "e7e5"},
	}
	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 24 {
		t.Fatalf("expected compact match view to fit 24 lines, got %d\n%s", len(lines), view)
	}
	if !strings.Contains(view, "[ Resign ]") || !strings.Contains(view, "[ Quit ]") {
		t.Fatalf("expected mouse action buttons in compact view, got %q", view)
	}
}
