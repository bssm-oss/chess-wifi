package tui

import (
	"testing"

	"github.com/bssm-oss/chess-wifi/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHostErrorIgnoredAfterCancel(t *testing.T) {
	m := newModel()
	listener, err := session.StartHost("Host", 0)
	if err != nil {
		t.Fatalf("StartHost returned error: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	m.listener = listener
	m.screen = screenWaiting

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := updated.(*model)
	if m2.screen != screenMenu {
		t.Fatalf("expected menu after cancel, got %s", m2.screen)
	}

	updated, _ = m2.Update(hostErrorMsg{listener: listener, err: errString("host listener closed")})
	m2 = updated.(*model)
	if m2.screen != screenMenu {
		t.Fatalf("expected stale host error to be ignored, got %s", m2.screen)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
