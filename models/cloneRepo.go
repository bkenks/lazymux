package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type CloneRepo struct {
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd 
}

func (m CloneRepo) View() string {
	return fmt.Sprintf(
		"Enter your repository url:\n\n%s\n\n%s",
		m.Model.View(),
		"(esc to quit)",
	) + "\n"
}