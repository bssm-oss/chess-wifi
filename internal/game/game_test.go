package game

import "testing"

func TestMatchApplyMoveUCI(t *testing.T) {
	m := New("Host", "Guest")
	snap, err := m.ApplyMoveUCI("e2e4")
	if err != nil {
		t.Fatalf("ApplyMoveUCI returned error: %v", err)
	}
	if snap.Turn != Black {
		t.Fatalf("expected black turn after white move, got %s", snap.Turn)
	}
	if snap.LastMoveUCI != "e2e4" {
		t.Fatalf("expected last move e2e4, got %q", snap.LastMoveUCI)
	}
	if len(snap.MoveHistory) != 1 {
		t.Fatalf("expected move history length 1, got %d", len(snap.MoveHistory))
	}
}

func TestLegalMovesForSquarePromotionOptions(t *testing.T) {
	fen := "8/P7/8/8/8/8/8/k6K w - - 0 1"
	moves, err := LegalMovesForSquare(fen, White, "a7")
	if err != nil {
		t.Fatalf("LegalMovesForSquare returned error: %v", err)
	}
	if len(moves) != 4 {
		t.Fatalf("expected 4 promotion moves, got %d", len(moves))
	}
}

func TestPieceAt(t *testing.T) {
	piece, side, ok, err := PieceAt("8/8/8/8/8/8/4P3/4K3 w - - 0 1", "e2")
	if err != nil {
		t.Fatalf("PieceAt returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected piece on e2")
	}
	if piece != 'P' || side != White {
		t.Fatalf("expected white pawn, got piece=%q side=%s", string(piece), side)
	}
}
