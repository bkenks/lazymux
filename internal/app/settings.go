package app

import (
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
		c.Tools.Editor = edited.Tools.Editor
		c.Behavior = edited.Behavior
		c.UI.AccentColor = edited.UI.AccentColor
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

// showSettings applies the saved UI settings to the repo list, restyling it
// when the accent color changed from prev.
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
	if ui.AccentColor != prev.AccentColor {
		accent, err := styles.ParseAccent(ui.AccentColor)
		if err != nil {
			cmds = append(cmds, m.toastCmd(events.ToastError, err.Error()))
		}
		styles.Apply(ui.Theme, styles.IsDark, accent)
		m.main.Restyle()
	}
	return tea.Batch(cmds...)
}
