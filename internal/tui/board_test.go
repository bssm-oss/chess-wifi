package tui

import (
	"testing"
	"time"

	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSquareFromMouseWhitePerspective(t *testing.T) {
	m := newModel()
	m.updateLayoutBounds()
	m.snapshot = game.Snapshot{FEN: "8/8/8/8/8/8/8/8 w - - 0 1"}
	sq, ok := m.squareFromMouse(m.boardBounds.x+1, m.boardBounds.y+cellHeight*7+1)
	if !ok {
		t.Fatal("expected mouse hit to map to square")
	}
	if sq != "a1" {
		t.Fatalf("expected a1, got %s", sq)
	}
}

func TestSquareFromMouseBlackPerspective(t *testing.T) {
	m := newModel()
	m.updateLayoutBounds()
	m.viewSide = game.Black
	m.snapshot = game.Snapshot{FEN: "8/8/8/8/8/8/8/8 w - - 0 1"}
	sq, ok := m.squareFromMouse(m.boardBounds.x+1, m.boardBounds.y+1)
	if !ok {
		t.Fatal("expected mouse hit to map to square")
	}
	if sq != "h1" {
		t.Fatalf("expected h1, got %s", sq)
	}
}

func TestRenderBoardShowsBlackPerspectiveLabels(t *testing.T) {
	got := []int{}
	for vrank := 7; vrank >= 0; vrank-- {
		got = append(got, rankLabelForView(vrank, game.Black))
	}
	want := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected black perspective labels: got %v want %v", got, want)
		}
	}
}

func TestHandleMouseSelectsPieceFromRenderedBoard(t *testing.T) {
	m := newModel()
	m.screen = screenMatch
	m.viewSide = game.White
	m.peerSession = &session.PeerSession{}
	m.snapshot = game.Snapshot{
		FEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Turn:   game.White,
		Status: "active",
	}
	m.updateLayoutBounds()
	x := m.boardBounds.x + cellWidth*4 + 2
	y := m.boardBounds.y + cellHeight*6 + 1
	updated, _ := m.handleMouse(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m2 := updated.(*model)
	if m2.selected != "e2" {
		t.Fatalf("expected e2 to be selected, got %q", m2.selected)
	}
	if len(m2.legalMoves) == 0 {
		t.Fatal("expected legal moves for selected pawn")
	}
}

func TestMouseClickMoveSyncsAcrossPeerSessions(t *testing.T) {
	hostListener, err := session.StartHost("Host", 9102)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = hostListener.Close() })

	acceptedCh := make(chan *session.PeerSession, 1)
	go func() {
		peer := <-hostListener.Accepted()
		acceptedCh <- peer
	}()

	guest, err := session.Join(t.Context(), "127.0.0.1:9102", "Guest")
	if err != nil {
		t.Fatalf("Join returned error: %v", err)
	}
	t.Cleanup(func() { _ = guest.Close() })

	var host *session.PeerSession
	select {
	case host = <-acceptedCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for host session")
	}
	t.Cleanup(func() { _ = host.Close() })

	m := newModel()
	m.screen = screenMatch
	m.viewSide = game.White
	m.peerSession = host
	m.snapshot = host.Snapshot()
	m.updateLayoutBounds()

	updated, cmd := m.handleMouse(tea.MouseMsg{X: m.boardBounds.x + cellWidth*4 + 2, Y: m.boardBounds.y + cellHeight*6 + 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = updated.(*model)
	if cmd != nil {
		t.Fatal("expected no command on source square selection")
	}
	if m.selected != "e2" {
		t.Fatalf("expected e2 selected, got %q", m.selected)
	}

	updated, cmd = m.handleMouse(tea.MouseMsg{X: m.boardBounds.x + cellWidth*4 + 2, Y: m.boardBounds.y + cellHeight*4 + 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = updated.(*model)
	if cmd == nil {
		t.Fatal("expected command on destination square click")
	}
	if msg := cmd(); msg != nil {
		if eventMsg, ok := msg.(sessionEventMsg); ok {
			t.Fatalf("unexpected session event message: %+v", eventMsg)
		}
	}

	select {
	case event := <-guest.Events():
		if event.Type != session.EventSnapshot {
			t.Fatalf("expected snapshot event, got %s", event.Type)
		}
		if event.Snapshot.LastMoveUCI != "e2e4" {
			t.Fatalf("expected move e2e4 to sync, got %q", event.Snapshot.LastMoveUCI)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for guest snapshot after mouse move")
	}
}
