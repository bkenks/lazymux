package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
)

// applySettings saves the fields the settings screen edits, brings the running
// app in line with them, and returns to the repo list with a toast reporting
// the outcome.
func (m *ModelManager) applySettings(edited config.Config) tea.Cmd {
	prev := m.cfg.UI
	saveFailed := m.saveConfig("settings", func(c *config.Config) {
		c.Behavior = edited.Behavior
		c.UI.Colors = edited.UI.Colors
		c.UI.ShowFullPath = edited.UI.ShowFullPath
		c.UI.ShowForge = edited.UI.ShowForge
		c.UI.ShowStats = edited.UI.ShowStats
		c.UI.SortMode = edited.UI.SortMode
	})

	cmds := []tea.Cmd{commands.SetState(domain.StateMain), m.showSettings(prev)}
	if saveFailed != nil {
		return tea.Batch(append(cmds, saveFailed)...)
	}
	return tea.Batch(append(cmds, m.toastCmd(events.ToastInfo, "settings saved"))...)
}

// showSettings applies the saved UI settings to the repo list, restyling the
// app when the colors changed from prev.
func (m *ModelManager) showSettings(prev config.UI) tea.Cmd {
	ui := m.cfg.UI
	domain.ShowFullPath = ui.ShowFullPath
	domain.ShowStats = ui.ShowStats
	domain.ShowForge = ui.ShowForge
	m.main.SyncForgeVisibility()

	var cmds []tea.Cmd
	if ui.SortMode != prev.SortMode {
		domain.Sort = domain.ParseSortMode(ui.SortMode)
		cmds = append(cmds, m.main.Resort())
	}
	if ui.Colors != prev.Colors {
		if err := ApplyColors(ui.Colors, styles.IsDark); err != nil {
			cmds = append(cmds, m.toastCmd(events.ToastError, err.Error()))
		}
		m.main.Restyle()
		m.cloneProgress = styles.NewProgress()
	}
	return tea.Batch(cmds...)
}

// ApplyColors styles the app with the base colors for a dark or a light
// terminal background. Invalid colors fall back to the defaults and are
// reported in the returned error.
func ApplyColors(modes config.ColorModes, isDark bool) error {
	mode, colors := "light", modes.Light
	if isDark {
		mode, colors = "dark", modes.Dark
	}
	palette, err := styles.NewPalette(colors.Main, colors.Accent, colors.Gray)
	styles.Apply(palette, isDark)
	if err != nil {
		return fmt.Errorf("%s mode %w", mode, err)
	}
	return nil
}
