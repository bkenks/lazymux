package uiConfirm

import (
	"github.com/bkenks/lazymux/tui/commands"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)
type Model struct {
	form huh.Form
	RepoPath string
}

func New() *Model {
	return &Model{
		form: *huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().Key("confirm").Title("Are you sure?").Affirmative("Yes").Negative("No"),
			),
		),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if form, ok := form.(*huh.Form); ok {
		m.form = *form
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		if m.form.GetBool("confirm") {
			cmd = commands.DeleteRepoAction(m.RepoPath)
		} else {
			cmd = commands.ConfirmDeleteDialogQuit()
		}
	}
	
	return m, cmd
}

func (m Model) View() string {
	return m.form.View()
}