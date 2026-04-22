package tui

import (
	"strings"
	"testing"

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
	m := newModel()
	m.viewSide = game.Black
	m.snapshot = game.Snapshot{FEN: "8/8/8/8/8/8/8/8 w - - 0 1"}
	board := m.renderBoard()
	if !strings.Contains(board, "1") || !strings.Contains(board, "8") {
		t.Fatal("expected rank labels to render for black perspective")
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
