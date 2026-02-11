package confirm

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type choice int

const (
	choiceYes choice = iota
	choiceNo
)

type Model struct {
	RepoPath string
	cursor   choice
}

func New() *Model {
	return &Model{
		cursor: choiceNo, // default to "No" for safety
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.ConfirmKeyMap.Proceed):
			cmds = append(
				cmds,
				commands.DeleteRepoCmd(m.RepoPath),
				commands.SetState(domain.StateMain),
			)
		case key.Matches(msg, constants.ConfirmKeyMap.Exit):
			cmds = append(cmds,
				commands.SetState(domain.StateMain),
			)
		}
	}

	cmd := tea.Batch(cmds...)
	return m, cmd
}

func (m *Model) View() string {

	title := styles.DialogTitleStyle.Render(
		"Delete Repository")

	subtitle := styles.DialogSubtitleStyle.Render(
		"Are you sure you want to delete this repository?")

	repoPath := styles.DialogRepoPath.Render(m.RepoPath)

	helpKeys := styles.DialogHelpStyle.Render(
		domain.FormatBindingsInline(
			constants.ConfirmKeyMap.HelpBinds(constants.Short),
		),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		repoPath,
		helpKeys,
	)

	renderedDialog := styles.DialogStyle.Render(content)

	placedContent := lipgloss.Place(
		constants.WindowSize.Width,
		constants.WindowSize.Height,
		lipgloss.Center,
		lipgloss.Center,
		renderedDialog,
	)

	return placedContent
}
