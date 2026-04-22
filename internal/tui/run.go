package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	model := newModel()
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
