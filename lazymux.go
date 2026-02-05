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
	currentModel tea.Model
}

func (m manager) Init() tea.Cmd {
	return nil
}

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

// 	switch msg:= msg.(type) {

// 		/////////////////////////////////////
// 		case tea.KeyMsg:
// 			if msg.String() == "ctrl+c" {
// 				return m, tea.Quit
// 			}
// 			if msg.String() == "enter" {
// 				selectedRepo := m.list.SelectedItem()
// 				var fullRepoPath string
// 				if repo, ok := selectedRepo.(repo); ok {
// 					fullRepoPath = getFullRepoPath(repo.path)
// 				}
				
// 				cmd := openLazygit(fullRepoPath)
				
// 				return m, cmd
// 			}
// 		/////////////////////////////////////
// 		case tea.WindowSizeMsg:
// 			h, v := docStyle.GetFrameSize()
// 			m.list.SetSize(msg.Width-h, msg.Height-v)
// 		/////////////////////////////////////

// 	}

// 	/////////////////////////////////////

// 	// Output
// 	var cmd tea.Cmd
// 	m.list, cmd = m.list.Update(msg)
// 	return m, cmd
// 	// End "Output"
// }

// func (m model) View() string {
// 	return docStyle.Render()
// }



func main() {
	m := models.InitialModel()

	m.Model.Title = "GitHub Repos"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _,err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}