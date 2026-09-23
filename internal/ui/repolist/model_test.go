package repolist

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	switch msg := cmd().(type) {
	case tea.QuitMsg:
		return true
	case tea.BatchMsg:
		for _, inner := range msg {
			if isQuit(inner) {
				return true
			}
		}
	}
	return false
}

func TestEscNeverQuits(t *testing.T) {
	m := New()
	if _, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape}); isQuit(cmd) {
		t.Error("esc quit the app; want esc to only go back")
	}
}

func TestQQuits(t *testing.T) {
	m := New()
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"}); !isQuit(cmd) {
		t.Error("q did not quit the app")
	}
}
