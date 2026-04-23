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
	choices := []string{"Host a match", "Join by address"}
	for _, match := range m.discoveries {
		choices = append(choices, fmt.Sprintf("Join %s · %s", match.PlayerName, match.Address))
	}
	m.menuBounds = make([]rect, len(choices))
	menuX, menuY := panelContentOrigin()
	var lines []string
	for i, choice := range choices {
		prefix := "  "
		style := menuItemStyle
		if i == m.menuIndex {
			prefix = "▸ "
			style = menuActiveStyle
		}
		m.menuBounds[i] = rect{x: menuX, y: menuY + i, w: max(24, len(choice)+4), h: 1}
		lines = append(lines, style.Render(prefix+choice))
	}
	help := subtleStyle.Render("마우스 클릭 / ↑↓ 선택 · Enter 실행 · q 종료")
	return lipgloss.JoinVertical(lipgloss.Left,
		panelStyle.Render(strings.Join(lines, "\n")),
		"",
		m.renderDiscoverySummary(),
		"",
		help,
	)
}

func (m *model) renderHostForm() string {
	x, y := panelContentOrigin()
	m.hostInputBounds = []rect{
		{x: x, y: y + 2, w: 24, h: 3},
		{x: x, y: y + 7, w: 18, h: 3},
	}
	m.hostStartBounds = rect{x: x, y: y + 12, w: 17, h: 1}
	m.hostBackBounds = rect{x: x + 19, y: y + 12, w: 9, h: 1}
	body := []string{
		labelStyle.Render("이름"),
		inputStyle.Render(m.hostInputs[0].View()),
		"",
		labelStyle.Render("포트"),
		inputStyle.Render(m.hostInputs[1].View()),
		"",
		buttonStyle.Render("[ Start hosting ]") + " " + buttonStyle.Render("[ Back ]"),
		subtleStyle.Render("필드/버튼 클릭 가능 · Tab 이동 · Enter 시작 · Esc 뒤로"),
	}
	return panelStyle.Render(strings.Join(body, "\n"))
}

func (m *model) renderJoinForm() string {
	x, y := panelContentOrigin()
	m.joinInputBounds = []rect{
		{x: x, y: y + 2, w: 24, h: 3},
		{x: x, y: y + 7, w: 34, h: 3},
	}
	m.joinConnectBounds = rect{x: x, y: y + 12, w: 11, h: 1}
	m.joinBackBounds = rect{x: x + 13, y: y + 12, w: 9, h: 1}
	body := []string{
		labelStyle.Render("이름"),
		inputStyle.Render(m.joinInputs[0].View()),
		"",
		labelStyle.Render("호스트 주소"),
		inputStyle.Render(m.joinInputs[1].View()),
		"",
		buttonStyle.Render("[ Connect ]") + " " + buttonStyle.Render("[ Back ]"),
		subtleStyle.Render("방 목록/필드/버튼 클릭 가능 · 예: 192.168.0.12:8787"),
	}
	if m.joining {
		body = append(body, "", accentStyle.Render("연결 중..."))
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		panelStyle.Render(strings.Join(body, "\n")),
		"",
		m.renderDiscoverySummary(),
	)
}

func (m *model) renderDiscoverySummary() string {
	lines := []string{labelStyle.Render("열려있는 LAN 매치")}
	x, y := panelContentOrigin()
	switch m.screen {
	case screenMenu:
		y += len(m.menuBounds) + 5
	case screenJoin:
		y += 17
	default:
		y += 5
	}
	m.discoveryBounds = nil
	if len(m.discoveries) > 0 {
		for i, match := range m.discoveries {
			line := fmt.Sprintf("%d. %s · %s · %s 전", i+1, match.PlayerName, match.Address, timeSince(match.LastSeen))
			m.discoveryBounds = append(m.discoveryBounds, rect{x: x, y: y + 1 + i, w: max(44, len(line)+4), h: 1})
			if m.screen == screenMenu && m.menuIndex == i+2 {
				lines = append(lines, menuActiveStyle.Render("▸ "+line))
			} else {
				lines = append(lines, infoStyle.Render("  "+line))
			}
		}
		if m.discoveryRun {
			lines = append(lines, subtleStyle.Render("검색 중..."))
		}
		lines = append(lines, subtleStyle.Render("목록에서 Enter를 누르면 바로 연결합니다."))
	} else if m.discoveryRun {
		lines = append(lines, subtleStyle.Render("검색 중..."))
	} else if m.discoveryErr != "" {
		lines = append(lines, dangerStyle.Render(m.discoveryErr))
	} else {
		lines = append(lines, subtleStyle.Render("아직 발견된 매치가 없습니다. Host가 대기 중이면 자동으로 나타납니다."))
	}
	return panelStyle.Width(64).Render(strings.Join(lines, "\n"))
}

func (m *model) renderWaiting() string {
	addresses := []string{subtleStyle.Render("호스트 주소를 아직 확인하지 못했습니다.")}
	m.waitingAddressBounds = nil
	m.waitingCopyBounds = nil
	m.waitingCancelBounds = rect{}
	x, y := panelContentOrigin()
	if m.listener != nil && len(m.listener.Addresses) > 0 {
		addresses = nil
		for i, addr := range m.listener.Addresses {
			rowY := y + 2 + i
			m.waitingAddressBounds = append(m.waitingAddressBounds, rect{x: x, y: rowY, w: len(addr), h: 1})
			m.waitingCopyBounds = append(m.waitingCopyBounds, rect{x: x + len(addr) + 2, y: rowY, w: 10, h: 1})
			addresses = append(addresses, accentStyle.Render(addr)+" "+buttonStyle.Render("[ Copy ]"))
		}
	}
	waited := subtleStyle.Render(fmt.Sprintf("대기 시간: %s", timeSince(m.waitingSince)))
	m.waitingCancelBounds = rect{x: x, y: y + 4 + len(addresses), w: 10, h: 1}
	body := append([]string{
		labelStyle.Render("같은 Wi-Fi의 상대에게 아래 주소를 공유하세요."),
		"",
	}, addresses...)
	body = append(body,
		"",
		waited,
		buttonStyle.Render("[ Cancel ]"),
		subtleStyle.Render("주소나 Copy 클릭으로 복사 · c 첫 주소 복사 · Esc 취소"),
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
	gap := "  "
	if m.compactLayout() {
		gap = " "
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, board, gap, sidebar)
	if m.promotion != nil {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", m.renderPromotion())
	}
	return content
}

func (m *model) renderBoard() string {
	m.updateLayoutBounds()
	perspective := m.side()
	var lines []string
	cw := m.cellWidth()
	for vrank := 7; vrank >= 0; vrank-- {
		worldRank := vrank
		if perspective == game.Black {
			worldRank = 7 - vrank
		}
		label := lipgloss.NewStyle().Width(2).Foreground(colorMuted).Render(fmt.Sprintf("%d", rankLabelForView(vrank, perspective)))
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
		if m.compactLayout() {
			lines = append(lines, label+strings.Join(bottom, ""))
		} else {
			lines = append(lines, label+strings.Join(top, ""))
			lines = append(lines, lipgloss.NewStyle().Width(2).Render(" ")+strings.Join(bottom, ""))
		}
	}
	var files []string
	for vfile := 0; vfile < 8; vfile++ {
		fileRune := rune('a' + vfile)
		if perspective == game.Black {
			fileRune = rune('a' + (7 - vfile))
		}
		files = append(files, lipgloss.NewStyle().Width(cw).Align(lipgloss.Center).Foreground(colorMuted).Render(string(fileRune)))
	}
	lines = append(lines, lipgloss.NewStyle().Width(2).Render(" ")+strings.Join(files, ""))
	boardBlock := strings.Join(lines, "\n")
	return panelStyle.Render(boardBlock)
}

func rankLabelForView(visibleRank int, perspective game.Side) int {
	if perspective == game.Black {
		return 8 - visibleRank
	}
	return visibleRank + 1
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
	top := base.Copy().Width(m.cellWidth()).Height(1).Render(" ")
	bottom := base.Copy().Width(m.cellWidth()).Height(1).Align(lipgloss.Center).Render(piece)
	return top, bottom
}

func (m *model) renderSidebar() string {
	if m.compactLayout() {
		return m.renderCompactSidebar()
	}
	role := string(m.side())
	peer := "—"
	if m.peerSession != nil {
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
		labelStyle.Render("Controls") + "\n" + subtleStyle.Render("Mouse click to move\nArrow / hjkl cursor\nEnter/Space select") + "\n" + buttonStyle.Render("[ Resign ]") + " " + buttonStyle.Render("[ Quit ]"),
		labelStyle.Render("Message") + "\n" + infoStyle.Width(28).Render(m.message),
		labelStyle.Render("Perspective") + "\n" + infoStyle.Render(role),
	}
	m.updateMatchButtonBounds(false)
	return panelStyle.Width(34).Render(strings.Join(sections, "\n\n"))
}

func (m *model) renderCompactSidebar() string {
	role := string(m.side())
	self := "—"
	peer := "—"
	if m.peerSession != nil {
		self = fmt.Sprintf("%s (%s)", m.peerSession.Self().Name, m.peerSession.Self().Side)
		peer = fmt.Sprintf("%s (%s)", m.peerSession.Peer().Name, m.peerSession.Peer().Side)
	}
	lastMoves := "Opening"
	if len(m.snapshot.MoveHistory) > 0 {
		start := len(m.snapshot.MoveHistory) - 4
		if start < 0 {
			start = 0
		}
		lastMoves = strings.Join(m.snapshot.MoveHistory[start:], " ")
	}
	lines := []string{
		labelStyle.Render("You ") + infoStyle.Render(self),
		labelStyle.Render("Peer ") + infoStyle.Render(peer),
		labelStyle.Render("Turn ") + accentStyle.Render(string(m.snapshot.Turn)) + "  " + labelStyle.Render("State ") + infoStyle.Render(m.snapshot.Status),
		labelStyle.Render("Moves ") + subtleStyle.Width(28).Render(lastMoves),
		labelStyle.Render("Msg ") + infoStyle.Width(28).Render(m.message),
		labelStyle.Render("View ") + infoStyle.Render(role),
		buttonStyle.Render("[ Resign ]") + " " + buttonStyle.Render("[ Quit ]"),
	}
	m.updateMatchButtonBounds(true)
	return panelStyle.Width(34).Render(strings.Join(lines, "\n"))
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

func panelContentOrigin() (int, int) {
	return framePaddingX + panelBorderSize + panelPaddingX, framePaddingY + headerHeight + panelBorderSize + panelPaddingY
}

func (m *model) updateMatchButtonBounds(compact bool) {
	boardPanelWidth := panelBorderSize*2 + panelPaddingX*2 + rankLabelWidth + m.cellWidth()*8
	gap := 2
	if compact {
		gap = 1
	}
	x := framePaddingX + boardPanelWidth + gap + panelBorderSize + panelPaddingX
	y := framePaddingY + headerHeight + panelBorderSize + panelPaddingY
	if compact {
		y += 6
	} else {
		y += 18
	}
	m.matchResignBounds = rect{x: x, y: y, w: 10, h: 1}
	m.matchQuitBounds = rect{x: x + 11, y: y, w: 8, h: 1}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
