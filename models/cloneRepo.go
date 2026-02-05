package models

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CloneRepo struct {
	windowWidth  int
	windowHeight int
	Model textinput.Model
}

func InitialCloneRepoModel() CloneRepo {
	ti := textinput.New()
	ti.Placeholder = "ex: git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return CloneRepo{
		Model: ti,
	}
}

func (m CloneRepo) Init() tea.Cmd {
	return textinput.Blink
}

func (m CloneRepo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd 
}

func (m CloneRepo) View() string {
	foreStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#FF06B7")).
		Padding(0, 1)

	boldStyle := lipgloss.NewStyle().Bold(true)
	title := boldStyle.Render("Enter your repository url\n")
	content := m.Model.View()
	footer := "\n(esc to quit)\n"
	layout := lipgloss.JoinVertical(lipgloss.Left, title, content, footer)

	return foreStyle.Render(layout)
}