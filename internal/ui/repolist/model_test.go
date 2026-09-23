package repolist

import (
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
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

func TestNumberKeysOpenSettingsScreens(t *testing.T) {
	cases := map[rune]domain.SessionState{'1': domain.StateSettings, '2': domain.StateKeybinds}
	for r, want := range cases {
		m := New()
		_, cmd := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		if cmd == nil {
			t.Fatalf("%q did nothing; want it to open state %v", r, want)
		}
		if got, ok := cmd().(events.SetState); !ok || got.State != want {
			t.Errorf("%q produced %#v; want SetState{%v}", r, cmd(), want)
		}
		if !slices.Contains(m.ReservedKeys(), string(r)) {
			t.Errorf("%q is not reserved, so a custom keybind could take it", r)
		}
	}
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
