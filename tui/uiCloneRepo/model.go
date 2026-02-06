package uiCloneRepo

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface: tea.Model
// - Model (UI) for the dialog to clone a new repo

type Model struct {
	width  int
	height int
	Model textinput.Model
}


func New() *Model {
	ti := textinput.New()
	ti.Placeholder = "ex: git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return &Model{
		Model: ti,
	}
}


func (m Model) Init() tea.Cmd { return textinput.Blink }


func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd = commands.QuitRepoDialog() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd
		case "enter":
			cmd = commands.CloneRepoAction(m.Model.Value())
			return m, cmd
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd 
}


func (m Model) View() string {
	title := constants.DialogTitleStyle.Render("Enter Your Repository URL:\n\n")
	textInput := m.Model.View()
	footer := "\n(esc to go back)"
	layout := lipgloss.JoinVertical(lipgloss.Left, title, textInput, footer)

	return lipgloss.Place(
		constants.WindowSize.Width,
		constants.WindowSize.Height,
		lipgloss.Left,
		lipgloss.Center,
		constants.DialogStyle.Render(layout),
	)
	
}

// End "Interface: tea.Model"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////