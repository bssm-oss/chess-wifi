package tui

import "github.com/bssm-oss/chess-wifi/internal/game"

func (m *model) updateLayoutBounds() {
	originX, originY := boardCellOrigin()
	m.boardBounds = rect{x: originX, y: originY, w: cellWidth * 8, h: cellHeight * 8}
}

func (m *model) squareFromMouse(x, y int) (string, bool) {
	m.updateLayoutBounds()
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
	if m.side() == game.Black {
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
	startX, startY := promotionOrigin()
	if y < startY || y > startY+1 {
		return -1
	}
	for i := range m.promotion.Options {
		left := startX + i*6
		right := left + 4
		if x >= left && x <= right {
			return i
		}
	}
	return -1
}

func boardCellOrigin() (int, int) {
	x := framePaddingX + panelBorderSize + panelPaddingX + rankLabelWidth
	y := framePaddingY + headerHeight + panelBorderSize + panelPaddingY
	return x, y
}

func promotionOrigin() (int, int) {
	x, y := boardCellOrigin()
	return x + 1, y + boardRows + 6
}
