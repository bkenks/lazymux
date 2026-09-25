package app

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/charmbracelet/x/ansi"
)

func TestScreensFitTheWindow(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	window := tea.WindowSizeMsg{Width: 80, Height: 30}

	for _, state := range []domain.SessionState{
		domain.StateSplash, domain.StateMain, domain.StateConfirmDelete,
		domain.StateCloneRepo, domain.StateSettings, domain.StateForgeRegistry,
		domain.StateKeybinds, domain.StateReposDir,
	} {
		m := New(configWithReposDir(t), "test")
		m.Update(window)
		m.Update(events.SetState{State: state})
		m.Update(window)

		if got := lipgloss.Height(m.View().Content); got > window.Height {
			t.Errorf("state %v renders %d rows in a %d-row window", state, got, window.Height)
		}
	}
}

func TestListScreensShowQuitOnce(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	window := tea.WindowSizeMsg{Width: 200, Height: 30}

	for _, state := range []domain.SessionState{
		domain.StateMain, domain.StateForgeRegistry, domain.StateKeybinds,
	} {
		m := New(configWithReposDir(t), "test")
		m.Update(window)
		m.Update(events.SetState{State: state})
		m.Update(window)

		if got := strings.Count(ansi.Strip(m.View().Content), "quit"); got != 1 {
			t.Errorf("state %v shows quit %d times, want 1", state, got)
		}
	}
}
