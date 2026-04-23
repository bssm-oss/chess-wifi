package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/bssm-oss/chess-wifi/internal/discovery"
	"github.com/bssm-oss/chess-wifi/internal/game"
	"github.com/bssm-oss/chess-wifi/internal/session"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
	cellWidth         = 5
	cellHeight        = 2
	compactCellWidth  = 3
	compactCellHeight = 1
	framePaddingX     = 2
	framePaddingY     = 1
	headerHeight      = 3
	panelBorderSize   = 1
	panelPaddingX     = 2
	panelPaddingY     = 1
	rankLabelWidth    = 2
	boardRows         = 16
	boardFilesY       = 16
	discoveryEvery    = 2 * time.Second
)

type hostAcceptedMsg struct {
	listener *session.HostListener
	session  *session.PeerSession
}
type hostErrorMsg struct {
	listener *session.HostListener
	err      error
}
type joinResultMsg struct {
	session *session.PeerSession
	err     error
}
type sessionEventMsg struct{ event session.Event }
type discoveryTickMsg struct{}
type discoveryResultMsg struct {
	matches []discovery.Match
	err     error
}
type clipboardResultMsg struct {
	value string
	err   error
	osc   string
}

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
	screen               screen
	menuIndex            int
	hostInputs           []textinput.Model
	joinInputs           []textinput.Model
	focusIndex           int
	listener             *session.HostListener
	peerSession          *session.PeerSession
	snapshot             game.Snapshot
	message              string
	width                int
	height               int
	viewSide             game.Side
	cursorFile           int
	cursorRank           int
	selected             string
	legalMoves           []game.MoveOption
	promotion            *promotionChoice
	boardBounds          rect
	joining              bool
	discoveries          []discovery.Match
	discoveryErr         string
	discoveryRun         bool
	menuBounds           []rect
	discoveryBounds      []rect
	hostInputBounds      []rect
	hostStartBounds      rect
	hostBackBounds       rect
	joinInputBounds      []rect
	joinConnectBounds    rect
	joinBackBounds       rect
	waitingAddressBounds []rect
	waitingCopyBounds    []rect
	waitingCancelBounds  rect
	matchResignBounds    rect
	matchQuitBounds      rect
	clipboardOSC         string
	waitingSince         time.Time
	quitting             bool
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
		screen:       screenMenu,
		hostInputs:   []textinput.Model{hostName, hostPort},
		joinInputs:   []textinput.Model{joinName, joinAddr},
		viewSide:     game.White,
		cursorFile:   4,
		cursorRank:   1,
		boardBounds:  rect{x: 0, y: 0, w: cellWidth * 8, h: cellHeight * 8},
		message:      "같은 Wi-Fi에서 직접 연결되는 체스를 준비하세요.",
		discoveryRun: true,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(scanDiscoveryCmd(), discoveryTickCmd())
}

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
		if m.listener == nil || m.listener != msg.listener || m.screen != screenWaiting {
			return m, nil
		}
		m.listener = nil
		m.peerSession = msg.session
		m.viewSide = msg.session.Role()
		m.snapshot = msg.session.Snapshot()
		m.screen = screenMatch
		m.message = fmt.Sprintf("%s connected. White moves first.", msg.session.Peer().Name)
		return m, waitForSessionEvent(msg.session)
	case hostErrorMsg:
		if m.listener == nil || m.listener != msg.listener || m.screen != screenWaiting {
			return m, nil
		}
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
		m.viewSide = msg.session.Role()
		m.snapshot = msg.session.Snapshot()
		m.screen = screenMatch
		m.message = fmt.Sprintf("Connected to %s.", msg.session.Peer().Name)
		return m, waitForSessionEvent(msg.session)
	case discoveryTickMsg:
		if m.canScanDiscovery() && !m.discoveryRun {
			m.discoveryRun = true
			return m, tea.Batch(scanDiscoveryCmd(), discoveryTickCmd())
		}
		return m, discoveryTickCmd()
	case discoveryResultMsg:
		m.discoveryRun = false
		if msg.err != nil {
			m.discoveryErr = msg.err.Error()
		} else {
			m.discoveryErr = ""
			m.discoveries = msg.matches
		}
		if m.screen == screenMenu && m.menuIndex >= m.menuChoiceCount() {
			m.menuIndex = m.menuChoiceCount() - 1
		}
		return m, nil
	case clipboardResultMsg:
		m.clipboardOSC = msg.osc
		if msg.err != nil {
			m.message = fmt.Sprintf("터미널 복사 요청됨: %s (시스템 클립보드: %v)", msg.value, msg.err)
		} else {
			m.message = fmt.Sprintf("복사됨: %s", msg.value)
		}
		return m, nil
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
			if m.menuIndex < m.menuChoiceCount()-1 {
				m.menuIndex++
			}
		case "enter", " ":
			if m.menuIndex == 0 {
				m.focusHostField(0)
				m.screen = screenHost
			} else if m.menuIndex == 1 {
				m.focusJoinField(0)
				m.screen = screenJoin
			} else {
				return m.startDiscoveredJoin(m.discoveries[m.menuIndex-2])
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
		case "c":
			if m.listener != nil && len(m.listener.Addresses) > 0 {
				return m, copyCmd(m.listener.Addresses[0])
			}
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
			m.viewSide = game.White
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

func (m *model) menuChoiceCount() int {
	return 2 + len(m.discoveries)
}

func (m *model) canScanDiscovery() bool {
	return m.listener == nil && m.peerSession == nil && !m.joining && (m.screen == screenMenu || m.screen == screenJoin || m.screen == screenError)
}

func (m *model) compactLayout() bool {
	return (m.width > 0 && m.width < 100) || (m.height > 0 && m.height < 30)
}

func (m *model) cellWidth() int {
	if m.compactLayout() {
		return compactCellWidth
	}
	return cellWidth
}

func (m *model) cellHeight() int {
	if m.compactLayout() {
		return compactCellHeight
	}
	return cellHeight
}

func (m *model) side() game.Side {
	if m.viewSide == "" {
		return game.White
	}
	return m.viewSide
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
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	switch m.screen {
	case screenMenu:
		return m.handleMenuMouse(msg)
	case screenHost:
		return m.handleHostMouse(msg)
	case screenJoin:
		return m.handleJoinMouse(msg)
	case screenWaiting:
		return m.handleWaitingMouse(msg)
	case screenMatch:
		return m.handleMatchMouse(msg)
	case screenError:
		m.screen = screenMenu
		m.message = "새 매치를 시작할 수 있습니다."
		m.peerSession = nil
		m.viewSide = game.White
		m.listener = nil
		m.clearSelection()
		return m, nil
	default:
		return m, nil
	}
}

func (m *model) handleMenuMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	for i, bounds := range m.menuBounds {
		if bounds.contains(msg.X, msg.Y) {
			m.menuIndex = i
			if i == 0 {
				m.focusHostField(0)
				m.screen = screenHost
				return m, nil
			}
			if i == 1 {
				m.focusJoinField(0)
				m.screen = screenJoin
				return m, nil
			}
			return m.startDiscoveredJoin(m.discoveries[i-2])
		}
	}
	return m.handleDiscoveryMouse(msg)
}

func (m *model) handleHostMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	for i, bounds := range m.hostInputBounds {
		if bounds.contains(msg.X, msg.Y) {
			m.focusHostField(i)
			return m, nil
		}
	}
	if m.hostStartBounds.contains(msg.X, msg.Y) {
		return m.startHosting()
	}
	if m.hostBackBounds.contains(msg.X, msg.Y) {
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m *model) handleJoinMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	for i, bounds := range m.joinInputBounds {
		if bounds.contains(msg.X, msg.Y) {
			m.focusJoinField(i)
			return m, nil
		}
	}
	if m.joinConnectBounds.contains(msg.X, msg.Y) {
		return m.startJoin()
	}
	if m.joinBackBounds.contains(msg.X, msg.Y) {
		m.screen = screenMenu
		return m, nil
	}
	return m.handleDiscoveryMouse(msg)
}

func (m *model) handleDiscoveryMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	for i, bounds := range m.discoveryBounds {
		if bounds.contains(msg.X, msg.Y) && i < len(m.discoveries) {
			return m.startDiscoveredJoin(m.discoveries[i])
		}
	}
	return m, nil
}

func (m *model) handleWaitingMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.listener == nil {
		return m, nil
	}
	for i, bounds := range m.waitingAddressBounds {
		if bounds.contains(msg.X, msg.Y) && i < len(m.listener.Addresses) {
			return m, copyCmd(m.listener.Addresses[i])
		}
	}
	for i, bounds := range m.waitingCopyBounds {
		if bounds.contains(msg.X, msg.Y) && i < len(m.listener.Addresses) {
			return m, copyCmd(m.listener.Addresses[i])
		}
	}
	if m.waitingCancelBounds.contains(msg.X, msg.Y) {
		_ = m.listener.Close()
		m.listener = nil
		m.screen = screenMenu
		m.message = "Hosting cancelled."
	}
	return m, nil
}

func (m *model) handleMatchMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.matchResignBounds.contains(msg.X, msg.Y) && m.peerSession != nil {
		m.message = "Resigning match..."
		return m, func() tea.Msg {
			return sessionEventMsg{event: session.Event{Type: session.EventError, Message: resignErr(m.peerSession.Resign())}}
		}
	}
	if m.matchQuitBounds.contains(msg.X, msg.Y) {
		if m.peerSession != nil {
			_ = m.peerSession.Close()
		}
		m.quitting = true
		return m, tea.Quit
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
	if m.snapshot.Turn != m.side() {
		m.message = "상대 턴입니다."
		return m, nil
	}
	if m.selected == "" {
		piece, side, ok, err := game.PieceAt(m.snapshot.FEN, square)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		if !ok || side != m.side() || piece == 0 {
			return m, nil
		}
		moves, err := game.LegalMovesForSquare(m.snapshot.FEN, m.side(), square)
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

func (m *model) startDiscoveredJoin(match discovery.Match) (tea.Model, tea.Cmd) {
	m.joinInputs[1].SetValue(match.Address)
	name := strings.TrimSpace(m.joinInputs[0].Value())
	if name == "" {
		name = "Guest"
	}
	m.joining = true
	m.message = fmt.Sprintf("Connecting to %s at %s...", match.PlayerName, match.Address)
	return m, joinCmd(name, match.Address)
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
				return hostAcceptedMsg{listener: listener, session: peer}
			}
		case err, ok := <-listener.Errors():
			if ok && err != nil {
				return hostErrorMsg{listener: listener, err: err}
			}
		}
		return hostErrorMsg{listener: listener, err: fmt.Errorf("host listener closed")}
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

func discoveryTickCmd() tea.Cmd {
	return tea.Tick(discoveryEvery, func(time.Time) tea.Msg {
		return discoveryTickMsg{}
	})
}

func scanDiscoveryCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		defer cancel()
		matches, err := discovery.ScanWithOptions(ctx, discovery.ScanOptions{Timeout: 1200 * time.Millisecond})
		return discoveryResultMsg{matches: matches, err: err}
	}
}

func copyCmd(value string) tea.Cmd {
	return func() tea.Msg {
		return clipboardResultMsg{value: value, err: clipboard.WriteAll(value), osc: ansi.SetSystemClipboard(value)}
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

func (r rect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}
