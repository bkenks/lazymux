package uiConfirm

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
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

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "left", "h", "up", "k":
			m.cursor = choiceYes

		case "right", "l", "down", "j":
			m.cursor = choiceNo

		case "enter":
			if m.cursor == choiceYes {
				cmd = commands.DeleteRepoAction(m.RepoPath)
				cmds = append(cmds, cmd)
			}

			cmd = commands.SetState(commands.StateMain)
			cmds = append(cmds, cmd)

		case "esc":
			cmd = commands.SetState(commands.StateMain)
			cmds = append(cmds, cmd)
		}
	}


	return m, nil
}

func (m Model) View() string {

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
