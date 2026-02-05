package main

import (
	"fmt"
	"os"

	"github.com/bkenks/lazymux/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type manager struct {
	Model tea.Model
}

func initalManagerModel() manager {
	return manager{
		Model: models.InitialRepoListModel(),
	}
}

func (m manager) Init() tea.Cmd {
	return nil
}

func (m manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m manager) View() string {
	return docStyle.Render(m.Model.View())
}



func main() {
	m := initalManagerModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _,err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}