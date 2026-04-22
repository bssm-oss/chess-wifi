package tui

import "github.com/bssm-oss/chess-wifi/internal/game"

func (m *model) squareFromMouse(x, y int) (string, bool) {
	if x < m.boardBounds.x || y < m.boardBounds.y {
		return "", false
	}
	relX := x - m.boardBounds.x
	relY := y - m.boardBounds.y
	if relX >= m.boardBounds.w || relY >= m.boardBounds.h {
		return "", false
	}
	visibleFile := relX / cellWidth
	visibleRank := 7 - (relY / cellHeight)
	file := visibleFile
	rank := visibleRank
	if m.peerSession != nil && m.peerSession.Role() == game.Black {
		file = 7 - visibleFile
		rank = 7 - visibleRank
	}
	sq, err := game.ParseSquareName(file, rank)
	if err != nil {
		return "", false
	}
	return sq, true
}

func (m *model) isLegalTarget(square string) bool {
	for _, move := range m.legalMoves {
		if move.Target == square {
			return true
		}
	}
	return false
}

func (m *model) promotionFromMouse(x, y int) int {
	baseY := boardY + cellHeight*8 + 6
	if y < baseY || y > baseY+2 {
		return -1
	}
	startX := boardX + 2
	for i := range m.promotion.Options {
		left := startX + i*5
		right := left + 4
		if x >= left && x <= right {
			return i
		}
	}
	return -1
}
