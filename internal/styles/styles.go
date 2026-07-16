package styles

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"
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

	// FormBoxStyle frames the inline add/edit forms on the forge screens so the
	// "you're in edit mode" state reads clearly against the list above it.
	FormBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(0, 1).
			MarginTop(1)

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

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Help
	//
	// Help is the shared bubbles/help renderer used by the non-list screens
	// (confirm, clone) so their key hints match the palette and separators the
	// list screens already render through help internally.

	Help = newHelpModel()

	// End "Help"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
)

// newHelpModel builds a help.Model styled from the current palette. Rebuilt by
// rebuildStyles so a theme swap re-colors it.
func newHelpModel() help.Model {
	h := help.New()
	key := lipgloss.NewStyle().Foreground(SubduedColor)
	desc := lipgloss.NewStyle().Foreground(VerySubduedColor)
	h.Styles.ShortKey = key
	h.Styles.FullKey = key
	h.Styles.ShortDesc = desc
	h.Styles.FullDesc = desc
	h.Styles.ShortSeparator = desc
	h.Styles.FullSeparator = desc
	return h
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection-list helpers
//
// Glyphs for the checkbox/selection screens (forge select, forge registry,
// repo forges) — encoded in each list item's title so the bubbles default
// delegate handles the row/selection styling. Subtle renders the small
// muted labels those screens draw beneath their lists.

const (
	GlyphCheckOn  = "◉"
	GlyphCheckOff = "○"
	GlyphPrimary  = "★"
)

// Subtle renders muted secondary text (input labels, hints). It's a function
// so it reads the current palette after styles.Apply swaps it for the theme.
func Subtle(s string) string { return lipgloss.NewStyle().Foreground(MediumGrey).Render(s) }

// RenderToast renders the footer toast at the given opacity (0..1), blending its
// text color toward the terminal background so it can fade in and out. width is
// the full footer width; opacity>=1 renders at the plain palette color.
func RenderToast(msg string, isError bool, opacity float64, width int) string {
	base, target := ToastInfoStyle, SubduedColor
	if isError {
		base, target = ToastErrorStyle, DarkPink
	}
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}

	dark := lipgloss.HasDarkBackground()
	if hex, ok := resolveHex(target, dark); ok && opacity < 1 {
		bg := "#e4e4e4"
		if dark {
			bg = "#1c1c1c"
		}
		to, err1 := colorful.Hex(hex)
		from, err2 := colorful.Hex(bg)
		if err1 == nil && err2 == nil {
			base = base.Foreground(lipgloss.Color(from.BlendLab(to, opacity).Clamped().Hex()))
		}
	}
	return base.Width(width).Render(msg)
}

// resolveHex extracts a hex string from a palette color for RGB blending,
// choosing the light or dark variant of an AdaptiveColor. ANSI-indexed colors
// have no hex to blend, so they report false and skip the fade.
func resolveHex(c lipgloss.TerminalColor, dark bool) (string, bool) {
	switch v := c.(type) {
	case lipgloss.AdaptiveColor:
		if dark {
			return v.Dark, true
		}
		return v.Light, true
	case lipgloss.Color:
		if s := string(v); strings.HasPrefix(s, "#") {
			return s, true
		}
	}
	return "", false
}

// End "Selection-list helpers"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
