package styles

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	DarkPink         lipgloss.TerminalColor
	DullGrey         lipgloss.TerminalColor
	Purple           lipgloss.TerminalColor
	VerySubduedColor lipgloss.TerminalColor
	SubduedColor     lipgloss.TerminalColor
	MediumGrey       lipgloss.TerminalColor
	DarkPurple       lipgloss.TerminalColor
	White            lipgloss.TerminalColor
}

var themes = map[string]Palette{
	"default": {
		DarkPink:         lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"},
		DullGrey:         lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"},
		Purple:           lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"},
		VerySubduedColor: lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#4b4b4b"},
		SubduedColor:     lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"},
		MediumGrey:       lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"},
		DarkPurple:       lipgloss.Color("62"),
		White:            lipgloss.Color("230"),
	},
	"mono": {
		DarkPink:         lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
		DullGrey:         lipgloss.AdaptiveColor{Light: "#C2C2C2", Dark: "#4D4D4D"},
		Purple:           lipgloss.AdaptiveColor{Light: "#666666", Dark: "#AAAAAA"},
		VerySubduedColor: lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#4B4B4B"},
		SubduedColor:     lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"},
		MediumGrey:       lipgloss.AdaptiveColor{Light: "#A4A4A4", Dark: "#777777"},
		DarkPurple:       lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
		White:            lipgloss.Color("255"),
	},
}

// Apply swaps the package-level color vars and rebuilds every style that
// depends on them. Safe to call at startup before the program runs; calling
// it after the program has rendered will not retroactively re-style frames
// already drawn.
func Apply(name string) {
	p, ok := themes[name]
	if !ok {
		p = themes["default"]
	}

	DarkPink = p.DarkPink
	DullGrey = p.DullGrey
	Purple = p.Purple
	VerySubduedColor = p.VerySubduedColor
	SubduedColor = p.SubduedColor
	MediumGrey = p.MediumGrey
	DarkPurple = p.DarkPurple
	White = p.White

	rebuildStyles()
}

func rebuildStyles() {
	MenuTitle = lipgloss.NewStyle().
		Background(DarkPurple).
		Foreground(White).
		Padding(0, 1).
		Margin(1, 0, 1, 2)

	MenuSubStyle = lipgloss.NewStyle().
		Foreground(MediumGrey).
		MarginLeft(2).
		MarginBottom(1)

	SelectedButton = ButtonStyle.
		Background(DarkPurple).
		Foreground(White).
		Bold(true)

	UnselectedButton = ButtonStyle.
		Background(DullGrey).
		Foreground(lipgloss.Color("250"))

	DialogStyle = lipgloss.NewStyle().
		Padding(1, 6, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SubduedColor)

	DialogTitleStyle = lipgloss.NewStyle().
		Background(DarkPurple).
		Foreground(White).
		Padding(0, 1).
		Margin(0, 0, 2)

	DialogRepoPath = lipgloss.NewStyle().
		Bold(true).
		MarginBottom(2).
		Foreground(DarkPink)

	ToastInfoStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(SubduedColor)

	ToastErrorStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(DarkPink).
		Bold(true)
}
