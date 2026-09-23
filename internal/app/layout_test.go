package app

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
)

func TestScreensFitTheWindow(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	window := tea.WindowSizeMsg{Width: 80, Height: 30}

	for _, state := range []domain.SessionState{
		domain.StateSplash, domain.StateMain, domain.StateConfirmDelete,
		domain.StateCloneRepo, domain.StateSettings, domain.StateForgeRegistry,
		domain.StateKeybinds,
	} {
		m := New(config.Default(), "test")
		m.Update(window)
		m.Update(events.SetState{State: state})
		m.Update(window)

		if got := lipgloss.Height(m.View().Content); got > window.Height {
			t.Errorf("state %v renders %d rows in a %d-row window", state, got, window.Height)
		}
	}
}
