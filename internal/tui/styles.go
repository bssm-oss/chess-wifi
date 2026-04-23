package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorBackground = lipgloss.Color("#201d1d")
	colorSurface    = lipgloss.Color("#302c2c")
	colorText       = lipgloss.Color("#fdfcfc")
	colorMuted      = lipgloss.Color("#9a9898")
	colorAccent     = lipgloss.Color("#007aff")
	colorSuccess    = lipgloss.Color("#30d158")
	colorDanger     = lipgloss.Color("#ff3b30")
	colorLightSq    = lipgloss.Color("#f1eeee")
	colorDarkSq     = lipgloss.Color("#646262")
	colorGold       = lipgloss.Color("#ff9f0a")
)

var (
	appStyle = lipgloss.NewStyle().
			Background(colorBackground).
			Foreground(colorText).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	subtleStyle  = lipgloss.NewStyle().Foreground(colorMuted)
	labelStyle   = lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(colorText)
	accentStyle  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	dangerStyle  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)

	panelStyle = lipgloss.NewStyle().
			Background(colorSurface).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#646262")).
			Padding(1, 2)

	menuItemStyle   = lipgloss.NewStyle().Foreground(colorText)
	menuActiveStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	buttonStyle     = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#646262")).
			Padding(0, 1)

	lightSquareStyle = lipgloss.NewStyle().
				Foreground(colorBackground).
				Background(colorLightSq).
				Width(cellWidth)

	darkSquareStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorDarkSq).
			Width(cellWidth)

	selectedSquareStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorAccent)

	legalSquareStyle = lipgloss.NewStyle().
				Foreground(colorBackground).
				Background(colorSuccess)

	lastMoveStyle = lipgloss.NewStyle().
			Foreground(colorBackground).
			Background(colorGold)

	promotionButtonStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSurface).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorMuted).
				Padding(0, 2)

	promotionButtonActiveStyle = promotionButtonStyle.Copy().
					Foreground(colorBackground).
					Background(colorAccent).
					BorderForeground(colorText)
)
