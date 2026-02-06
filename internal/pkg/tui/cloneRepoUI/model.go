package cloneRepoUI

import (
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int
	Model textinput.Model
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "ex: git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return Model{
		Model: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd = cmdEsc() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd
		case "enter":
			cmd = cloneRepo(m.Model.Value())
			return m, cmd
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd 
}

func (m Model) View() string {
	foreStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#FF06B7")).
		Padding(0, 1)


	boldStyle := lipgloss.NewStyle().Bold(true)
	title := boldStyle.Render("Enter your repository url\n")
	content := m.Model.View()
	footer := "\n(esc to quit)\n"
	layout := lipgloss.JoinVertical(lipgloss.Center, title, content, footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		foreStyle.Render(layout),
	)
	
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Cmds & Msgs

type MsgEsc struct {}

func cmdEsc() tea.Cmd {
	return func() tea.Msg {
		return MsgEsc{}
	}
}

type MsgGhqGet struct { err error }

func cloneRepo (repo string) tea.Cmd {
	c := exec.Command("ghq", "get", repo)

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return MsgGhqGet{err: err}
	})

	return tea.Cmd(cmd)
}

// End "Cmds & Msgs"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions



// End "Functions"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////