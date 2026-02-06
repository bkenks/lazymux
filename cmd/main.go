package main

import (
	"fmt"
	"os"

	tui "github.com/bkenks/lazymux/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

// var docStyle: lipgloss.NewStyle().Margin(1, 2)

func main() {

	tui := tui.InitialModel()
	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}