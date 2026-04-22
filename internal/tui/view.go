package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenMenu:
		return m.renderFrame("chess-wifi match", m.renderMenu())
	case screenHost:
		return m.renderFrame("호스트 설정", m.renderHostForm())
	case screenJoin:
		return m.renderFrame("참가 설정", m.renderJoinForm())
	case screenWaiting:
		return m.renderFrame("상대 연결 대기 중", m.renderWaiting())
	case screenMatch:
		return m.renderFrame("LAN Match", m.renderMatch())
	case screenError:
		return m.renderFrame("상태 알림", m.renderError())
	default:
		return ""
	}
}

func (m *model) renderFrame(title, body string) string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		subtleStyle.Render(m.message),
	)
	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", body))
}

func (m *model) renderMenu() string {
	choices := []string{"Host a match", "Join a match"}
	var lines []string
	for i, choice := range choices {
		prefix := "  "
		style := menuItemStyle
		if i == m.menuIndex {
			prefix = "▸ "
			style = menuActiveStyle
		}
		lines = append(lines, style.Render(prefix+choice))
	}
	help := subtleStyle.Render("↑/↓ 선택 · Enter 실행 · q 종료")
	return lipgloss.JoinVertical(lipgloss.Left,
		panelStyle.Render(strings.Join(lines, "\n")),
		"",
		help,
	)
}

func (m *model) renderHostForm() string {
	body := []string{
		labelStyle.Render("이름"),
		inputStyle.Render(m.hostInputs[0].View()),
		"",
		labelStyle.Render("포트"),
		inputStyle.Render(m.hostInputs[1].View()),
		"",
		subtleStyle.Render("Tab으로 이동 · 마지막 필드에서 Enter로 시작 · Esc 뒤로"),
	}
	return panelStyle.Render(strings.Join(body, "\n"))
}

func (m *model) renderJoinForm() string {
	body := []string{
		labelStyle.Render("이름"),
		inputStyle.Render(m.joinInputs[0].View()),
		"",
		labelStyle.Render("호스트 주소"),
		inputStyle.Render(m.joinInputs[1].View()),
		"",
		subtleStyle.Render("예: 192.168.0.12:8787 · Tab 이동 · Enter 연결"),
	}
	if m.joining {
		body = append(body, "", accentStyle.Render("연결 중..."))
	}
	return panelStyle.Render(strings.Join(body, "\n"))
}

func (m *model) renderWaiting() string {
	addresses := []string{subtleStyle.Render("호스트 주소를 아직 확인하지 못했습니다.")}
	if m.listener != nil && len(m.listener.Addresses) > 0 {
		addresses = nil
		for _, addr := range m.listener.Addresses {
			addresses = append(addresses, accentStyle.Render(addr))
		}
	}
	waited := subtleStyle.Render(fmt.Sprintf("대기 시간: %s", timeSince(m.waitingSince)))
	body := append([]string{
		labelStyle.Render("같은 Wi-Fi의 상대에게 아래 주소를 공유하세요."),
		"",
	}, addresses...)
	body = append(body,
		"",
		waited,
		subtleStyle.Render("상대가 연결되면 자동으로 보드로 전환됩니다. Esc 취소"),
	)
	return panelStyle.Render(strings.Join(body, "\n"))
}

func (m *model) renderError() string {
	return panelStyle.Render(strings.Join([]string{
		dangerStyle.Render(m.message),
		"",
		subtleStyle.Render("Enter 또는 Esc로 메뉴로 돌아갑니다. q로 종료합니다."),
	}, "\n"))
}

func (m *model) renderMatch() string {
	board := m.renderBoard()
	sidebar := m.renderSidebar()
	content := lipgloss.JoinHorizontal(lipgloss.Top, board, "  ", sidebar)
	if m.promotion != nil {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", m.renderPromotion())
	}
	return content
}

func (m *model) renderBoard() string {
	perspective := game.White
	if m.peerSession != nil {
		perspective = m.peerSession.Role()
	}
	var lines []string
	for vrank := 7; vrank >= 0; vrank-- {
		worldRank := vrank
		if perspective == game.Black {
			worldRank = 7 - vrank
		}
		label := lipgloss.NewStyle().Width(2).Foreground(colorMuted).Render(fmt.Sprintf("%d", vrank+1))
		var top []string
		var bottom []string
		for vfile := 0; vfile < 8; vfile++ {
			worldFile := vfile
			if perspective == game.Black {
				worldFile = 7 - vfile
			}
			sq, _ := game.ParseSquareName(worldFile, worldRank)
			cellTop, cellBottom := m.renderSquare(sq, worldFile, worldRank)
			top = append(top, cellTop)
			bottom = append(bottom, cellBottom)
		}
		lines = append(lines, label+strings.Join(top, ""))
		lines = append(lines, lipgloss.NewStyle().Width(2).Render(" ")+strings.Join(bottom, ""))
	}
	var files []string
	for vfile := 0; vfile < 8; vfile++ {
		fileRune := rune('a' + vfile)
		if perspective == game.Black {
			fileRune = rune('a' + (7 - vfile))
		}
		files = append(files, lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center).Foreground(colorMuted).Render(string(fileRune)))
	}
	lines = append(lines, lipgloss.NewStyle().Width(2).Render(" ")+strings.Join(files, ""))
	boardBlock := strings.Join(lines, "\n")
	return panelStyle.Render(boardBlock)
}

func (m *model) renderSquare(square string, file, rank int) (string, string) {
	pieceRune, _, ok, _ := game.PieceAt(m.snapshot.FEN, square)
	piece := " "
	if ok {
		piece = unicodePiece(pieceRune)
	}
	base := darkSquareStyle
	if (file+rank)%2 == 0 {
		base = lightSquareStyle
	}
	if square == m.selected {
		base = selectedSquareStyle
	} else if m.isLegalTarget(square) {
		base = legalSquareStyle
	} else if square == m.snapshot.LastMoveUCI[:min(len(m.snapshot.LastMoveUCI), 2)] || (len(m.snapshot.LastMoveUCI) >= 4 && square == m.snapshot.LastMoveUCI[2:4]) {
		base = lastMoveStyle
	}
	if cur, _ := game.ParseSquareName(m.cursorFile, m.cursorRank); cur == square {
		base = base.BorderForeground(colorAccent)
	}
	top := base.Copy().Width(cellWidth).Height(1).Render(" ")
	bottom := base.Copy().Width(cellWidth).Height(1).Align(lipgloss.Center).Render(piece)
	return top, bottom
}

func (m *model) renderSidebar() string {
	role := "spectator"
	peer := "—"
	if m.peerSession != nil {
		role = string(m.peerSession.Role())
		peer = fmt.Sprintf("%s (%s)", m.peerSession.Peer().Name, m.peerSession.Peer().Side)
	}
	self := "—"
	if m.peerSession != nil {
		self = fmt.Sprintf("%s (%s)", m.peerSession.Self().Name, m.peerSession.Self().Side)
	}
	lastMoves := "Opening position"
	if len(m.snapshot.MoveHistory) > 0 {
		start := len(m.snapshot.MoveHistory) - 6
		if start < 0 {
			start = 0
		}
		lastMoves = strings.Join(m.snapshot.MoveHistory[start:], "  ")
	}
	statusColor := accentStyle
	if m.snapshot.Status != "active" {
		statusColor = successStyle
	}
	sections := []string{
		labelStyle.Render("You") + "\n" + infoStyle.Render(self),
		labelStyle.Render("Peer") + "\n" + infoStyle.Render(peer),
		labelStyle.Render("Turn") + "\n" + statusColor.Render(string(m.snapshot.Turn)),
		labelStyle.Render("State") + "\n" + infoStyle.Render(m.snapshot.Status),
		labelStyle.Render("Result") + "\n" + infoStyle.Render(orDash(m.snapshot.Result)),
		labelStyle.Render("Recent moves") + "\n" + subtleStyle.Width(28).Render(lastMoves),
		labelStyle.Render("Controls") + "\n" + subtleStyle.Render("Mouse click to move\nArrow / hjkl cursor\nEnter/Space select\nr resign · q quit"),
		labelStyle.Render("Message") + "\n" + infoStyle.Width(28).Render(m.message),
		labelStyle.Render("Perspective") + "\n" + infoStyle.Render(role),
	}
	return panelStyle.Width(34).Render(strings.Join(sections, "\n\n"))
}

func (m *model) renderPromotion() string {
	var buttons []string
	for i, option := range m.promotion.Options {
		style := promotionButtonStyle
		if i == m.promotion.Index {
			style = promotionButtonActiveStyle
		}
		buttons = append(buttons, style.Render(strings.ToUpper(option.Promotion)))
	}
	content := []string{
		labelStyle.Render("프로모션 선택"),
		subtleStyle.Render("←/→ 또는 클릭 후 Enter"),
		lipgloss.JoinHorizontal(lipgloss.Left, buttons...),
	}
	return panelStyle.Render(strings.Join(content, "\n"))
}

func unicodePiece(piece rune) string {
	mapping := map[rune]string{
		'K': "♔", 'Q': "♕", 'R': "♖", 'B': "♗", 'N': "♘", 'P': "♙",
		'k': "♚", 'q': "♛", 'r': "♜", 'b': "♝", 'n': "♞", 'p': "♟",
	}
	if symbol, ok := mapping[piece]; ok {
		return symbol
	}
	return string(piece)
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func timeSince(ts time.Time) string {
	if ts.IsZero() {
		return "0s"
	}
	return time.Since(ts).Round(time.Second).String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
