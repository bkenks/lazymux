package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

func adaptive(light, dark string) compat.AdaptiveColor {
	return compat.AdaptiveColor{Light: lipgloss.Color(light), Dark: lipgloss.Color(dark)}
}

type Palette struct {
	DarkPink         color.Color
	DullGrey         color.Color
	Purple           color.Color
	VerySubduedColor color.Color
	SubduedColor     color.Color
	MediumGrey       color.Color
	DarkPurple       color.Color
	White            color.Color
}

var themes = map[string]Palette{
	"default": {
		DarkPink:         adaptive("#EE6FF8", "#EE6FF8"),
		DullGrey:         adaptive("#C2B8C2", "#4D4D4D"),
		Purple:           adaptive("#F793FF", "#AD58B4"),
		VerySubduedColor: adaptive("#DDDADA", "#4b4b4b"),
		SubduedColor:     adaptive("#9B9B9B", "#5C5C5C"),
		MediumGrey:       adaptive("#A49FA5", "#777777"),
		DarkPurple:       lipgloss.Color("62"),
		White:            lipgloss.Color("230"),
	},
	"mono": {
		DarkPink:         adaptive("#000000", "#FFFFFF"),
		DullGrey:         adaptive("#C2C2C2", "#4D4D4D"),
		Purple:           adaptive("#666666", "#AAAAAA"),
		VerySubduedColor: adaptive("#DDDDDD", "#4B4B4B"),
		SubduedColor:     adaptive("#9B9B9B", "#5C5C5C"),
		MediumGrey:       adaptive("#A4A4A4", "#777777"),
		DarkPurple:       adaptive("#000000", "#FFFFFF"),
		White:            lipgloss.Color("255"),
	},
}

func init() { Apply("default") }

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

	FormBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(0, 1).
		MarginTop(1)

	TerminalFrameStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple)

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

	Help = newHelpModel()
}
