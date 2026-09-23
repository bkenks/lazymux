package app

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/config"
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

func TestTerminalFillsTheWindowWithinAMargin(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	window := tea.WindowSizeMsg{Width: 80, Height: 30}
	m := New(config.Default(), "test")
	m.Update(window)
	m.Update(events.RunKeybind{Keybind: config.Keybind{Name: "run", Command: "true"}, Dir: t.TempDir()})
	m.Update(events.SetState{State: domain.StateTerminal})

	rows := strings.Split(ansi.Strip(m.View().Content), "\n")
	if len(rows) != window.Height {
		t.Fatalf("terminal renders %d rows in a %d-row window", len(rows), window.Height)
	}
	if !strings.Contains(rows[1], "run") {
		t.Errorf("row 1 = %q, want the header right below a one-row margin", rows[1])
	}
	for _, edge := range []struct {
		row  int
		want string
	}{{2, "╭"}, {window.Height - 3, "╰"}} {
		row := []rune(rows[edge.row])
		if string(row[2]) != edge.want || string(row[window.Width-3]) == " " {
			t.Errorf("row %d = %q, want the border spanning columns 2..%d",
				edge.row, rows[edge.row], window.Width-3)
		}
	}
}
