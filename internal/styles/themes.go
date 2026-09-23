package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// shade is a palette color's value on a light and on a dark terminal background.
type shade struct{ light, dark color.Color }

func adaptive(light, dark string) shade {
	return shade{light: lipgloss.Color(light), dark: lipgloss.Color(dark)}
}

func fixed(c string) shade { return adaptive(c, c) }

func (s shade) resolve(isDark bool) color.Color {
	return lipgloss.LightDark(isDark)(s.light, s.dark)
}

type Palette struct {
	DarkPink         shade
	DullGrey         shade
	Purple           shade
	VerySubduedColor shade
	SubduedColor     shade
	MediumGrey       shade
	DarkPurple       shade
	White            shade
}

var themes = map[string]Palette{
	"default": {
		DarkPink:         adaptive("#EE6FF8", "#EE6FF8"),
		DullGrey:         adaptive("#C2B8C2", "#4D4D4D"),
		Purple:           adaptive("#F793FF", "#AD58B4"),
		VerySubduedColor: adaptive("#B2B2B2", "#4b4b4b"),
		SubduedColor:     adaptive("#9B9B9B", "#5C5C5C"),
		MediumGrey:       adaptive("#A49FA5", "#777777"),
		DarkPurple:       fixed("62"),
		White:            fixed("230"),
	},
	"mono": {
		DarkPink:         adaptive("#000000", "#FFFFFF"),
		DullGrey:         adaptive("#C2C2C2", "#4D4D4D"),
		Purple:           adaptive("#666666", "#AAAAAA"),
		VerySubduedColor: adaptive("#B2B2B2", "#4B4B4B"),
		SubduedColor:     adaptive("#9B9B9B", "#5C5C5C"),
		MediumGrey:       adaptive("#A4A4A4", "#777777"),
		DarkPurple:       adaptive("#000000", "#FFFFFF"),
		White:            adaptive("#FFFFFF", "#000000"),
	},
}

// IsDark reports whether the terminal background is dark. Apply sets it, and
// screens pass it to the bubbles and huh defaults they build.
var IsDark = true

func init() { Apply("default", true) }

// Apply picks the named theme's colors for a light or dark terminal background
// and rebuilds every style that depends on them. Call it before the program
// runs; frames already drawn are not re-styled.
func Apply(name string, isDark bool) {
	p, ok := themes[name]
	if !ok {
		p = themes["default"]
	}

	IsDark = isDark
	DarkPink = p.DarkPink.resolve(isDark)
	DullGrey = p.DullGrey.resolve(isDark)
	Purple = p.Purple.resolve(isDark)
	VerySubduedColor = p.VerySubduedColor.resolve(isDark)
	SubduedColor = p.SubduedColor.resolve(isDark)
	MediumGrey = p.MediumGrey.resolve(isDark)
	DarkPurple = p.DarkPurple.resolve(isDark)
	White = p.White.resolve(isDark)

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
		Foreground(adaptive("#3C3C3C", "250").resolve(IsDark))

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
