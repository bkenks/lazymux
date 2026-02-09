package uiConfirm

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
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

		case key.Matches(msg, constants.UIConfirm.Left):
			m.cursor = choiceYes

		case key.Matches(msg, constants.UIConfirm.Right):
			m.cursor = choiceNo

		case key.Matches(msg, constants.DefaultKeyMap.Select):
			if m.cursor == choiceYes {
				cmds = append(
					cmds,
					commands.DeleteRepoCmd(m.RepoPath),
					commands.SetState(commands.StateMain),
				)
			} else {
				cmds = append(cmds,
					commands.SetState(commands.StateMain),
				)
			}
		case key.Matches(msg, constants.DefaultKeyMap.Esc):
			cmds = append(
				cmds,
				commands.SetState(commands.StateMain),
			)
		}
	}

	cmd := tea.Batch(cmds...)
	return m, cmd
}

func (m *Model) View() string {

	title := constants.Title.Margin(1, 0, 2).Render(
		"Delete Repository")

	subtitle := constants.SubtitleStyle.Render(
		"Are you sure you want to delete this repository?")


	repoPath := constants.SubtitleStyle.
		Bold(true).
		MarginBottom(2).
		Foreground(constants.DarkPink).
		Render(m.RepoPath)

	/////////////////////////////////////

	yes := constants.UnselectedButton.Render("Yes")
	no := constants.UnselectedButton.Render("No")

	if m.cursor == choiceYes {
		yes = constants.SelectedButton.Render("Yes")
	} else {
		no = constants.SelectedButton.Render("No")
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		yes,
		no,
	)

	/////////////////////////////////////

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		repoPath,
		buttons,
	)

	renderedDialog := constants.DialogStyle.Render(content)

	placedContent := lipgloss.Place(
		constants.WindowSize.Width,
		constants.WindowSize.Height,
		lipgloss.Center,
		lipgloss.Center,
		renderedDialog,
	)

	return placedContent
}
