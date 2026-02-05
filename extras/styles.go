package extras

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	DocStyle lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.DocStyle = lipgloss.NewStyle().
		Margin(1, 2)

	/////////////////////////////////////

	return s
}