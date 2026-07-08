package styles

import (
	"strings"

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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection-list helpers
//
// Shared glyphs + render helpers for the checkbox/selection screens (forge
// select, forge registry, repo forges) so they match the rest of the app.
// These are functions rather than pre-built styles so they read the current
// palette after styles.Apply has swapped it for the active theme.

const (
	GlyphCheckOn  = "◉"
	GlyphCheckOff = "○"
	GlyphPrimary  = "★"
	GlyphCursor   = "▸"
	HelpSep       = "  ·  "
)

// Accent renders text in the theme's accent color (used for primary/cursor).
func Accent(s string) string { return lipgloss.NewStyle().Foreground(DarkPink).Render(s) }

// Subtle renders muted secondary text (hosts, hints).
func Subtle(s string) string { return lipgloss.NewStyle().Foreground(MediumGrey).Render(s) }

// Strong renders emphasized foreground text for the focused row.
func Strong(s string) string { return lipgloss.NewStyle().Foreground(White).Bold(true).Render(s) }

// Cursor renders the row-cursor glyph (accented) when active, else a blank.
func Cursor(active bool) string {
	if active {
		return Accent(GlyphCursor)
	}
	return " "
}

// ForgeRow renders one selectable forge line, shared by the forge-select and
// repo-forges screens: "▸ ★ ◉ name        host". The primary star and a
// checked box are accented; the focused row's name is emphasized.
func ForgeRow(active, checked, primary bool, name, host string, nameWidth int) string {
	prim := " "
	if primary {
		prim = Accent(GlyphPrimary)
	}
	box := lipgloss.NewStyle().Foreground(MediumGrey).Render(GlyphCheckOff)
	if checked {
		box = Accent(GlyphCheckOn)
	}
	padded := name
	if pad := nameWidth - len([]rune(name)); pad > 0 {
		padded = name + spaces(pad)
	}
	label := padded
	if active {
		label = Strong(padded)
	}
	return " " + Cursor(active) + " " + prim + " " + box + "  " + label + "  " + Subtle(host)
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// Help renders alternating key/label pairs as a subdued, separated hint line:
// Help("↑/↓", "move", "esc", "back") → "↑/↓ move  ·  esc back" (accented keys).
func Help(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, Accent(pairs[i])+" "+Subtle(pairs[i+1]))
	}
	return strings.Join(parts, HelpSep)
}

// End "Selection-list helpers"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
