package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Colors
	DarkPink         lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	DullGrey         lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}
	Purple           lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	VerySubduedColor lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#4b4b4b"}
	SubduedColor     lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}
	MediumGrey       lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}
	DarkPurple       lipgloss.TerminalColor = lipgloss.Color("62")
	White            lipgloss.TerminalColor = lipgloss.Color("230")

	// End "Colors"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Terminal Window

	DocStyle = lipgloss.NewStyle().
			Margin(3, 1)

	// End "Terminal Window"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Menu

	MenuTitle = lipgloss.NewStyle().
			Background(DarkPurple).
			Foreground(White).
			Padding(0, 1).
			Margin(1, 0, 1, 2)

	MenuHelpStyle = lipgloss.NewStyle().
			Margin(1, 0, 0, 2)

	MenuSubStyle = lipgloss.NewStyle().
			Foreground(MediumGrey).
			MarginLeft(2).
			MarginBottom(1)

	// End "Menu"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Buttons

	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1)

	SelectedButton = ButtonStyle.
			Background(DarkPurple).
			Foreground(White).
			Bold(true)

	UnselectedButton = ButtonStyle.
				Background(DullGrey).
				Foreground(lipgloss.Color("250"))

	// End "Buttons"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Dialog

	DialogStyle = lipgloss.NewStyle().
			Padding(1, 6, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SubduedColor)

	DialogTitleStyle = lipgloss.NewStyle().
				Background(DarkPurple).
				Foreground(White).
				Padding(0, 1).
				Margin(0, 0, 2)

	DialogHelpStyle = lipgloss.NewStyle()

	DialogSubtitleStyle = lipgloss.NewStyle().
				MarginBottom(1)

	DialogRepoPath = lipgloss.NewStyle().
			Bold(true).
			MarginBottom(2).
			Foreground(DarkPink)

	// End "Dialog"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Status footer / toasts

	ToastIdleStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ToastInfoStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(SubduedColor)

	ToastErrorStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(DarkPink).
			Bold(true)

	// End "Status footer / toasts"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
)
