package tui

import (
	"testing"

	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/session"
)

func TestSquareFromMouseWhitePerspective(t *testing.T) {
	m := newModel()
	m.peerSession = &session.PeerSession{}
	m.snapshot = game.Snapshot{FEN: "8/8/8/8/8/8/8/8 w - - 0 1"}
	sq, ok := m.squareFromMouse(boardX+1, boardY+cellHeight*7+1)
	if !ok {
		t.Fatal("expected mouse hit to map to square")
	}
	if sq != "a1" {
		t.Fatalf("expected a1, got %s", sq)
	}
}
