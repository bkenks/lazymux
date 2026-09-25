package styles

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// The colors and every style built from them are set by Apply, which init runs
// with the default palette; rebuildStyles is their one definition.
var (
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Colors
	Main        color.Color
	OnMain      color.Color
	MainText    color.Color
	Accent      color.Color
	AccentMuted color.Color
	Text        color.Color
	Muted       color.Color
	Subdued     color.Color
	Faint       color.Color
	Surface     color.Color
	OnSurface   color.Color

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

	MenuTitle lipgloss.Style

	MenuHelpStyle = lipgloss.NewStyle().
			Margin(1, 0, 0, 2)

	MenuSubStyle lipgloss.Style

	// End "Menu"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Buttons

	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1)

	SelectedButton   lipgloss.Style
	UnselectedButton lipgloss.Style

	// End "Buttons"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Dialog

	DialogStyle      lipgloss.Style
	DialogTitleStyle lipgloss.Style

	// FormBoxStyle frames the inline add/edit forms on the forge screens so the
	// "you're in edit mode" state reads clearly against the list above it.
	FormBoxStyle lipgloss.Style

	DialogHelpStyle = lipgloss.NewStyle()

	DialogSubtitleStyle = lipgloss.NewStyle().
				MarginBottom(1)

	DialogRepoPath lipgloss.Style

	// End "Dialog"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Status footer / toasts

	ToastIdleStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ToastInfoStyle  lipgloss.Style
	ToastErrorStyle lipgloss.Style

	// End "Status footer / toasts"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Help
	//
	// Help is the shared bubbles/help renderer used by the non-list screens
	// (confirm, clone) so their key hints match the palette and separators the
	// list screens already render through help internally.

	Help help.Model

	// End "Help"
	///////////////////////////////////////////////////////////////////////////////////////////////////////////////
)

// newHelpModel builds a help.Model styled from the current palette. Rebuilt by
// rebuildStyles so a palette change re-colors it.
func newHelpModel() help.Model {
	h := help.New()
	h.Styles = help.DefaultStyles(IsDark)
	key := lipgloss.NewStyle().Foreground(Faint)
	desc := lipgloss.NewStyle().Foreground(Subdued)
	h.Styles.ShortKey = key
	h.Styles.FullKey = key
	h.Styles.ShortDesc = desc
	h.Styles.FullDesc = desc
	h.Styles.ShortSeparator = key
	h.Styles.FullSeparator = key
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
	GlyphOrigin   = "★"
)

// Subtle renders muted secondary text (input labels, hints). It's a function
// so it reads the current palette after styles.Apply changes it.
func Subtle(s string) string { return lipgloss.NewStyle().Foreground(Muted).Render(s) }

// RenderToast renders the footer toast at the given opacity (0..1), blending its
// text color toward the terminal background so it can fade in and out. width is
// the full footer width; opacity>=1 renders at the plain palette color.
func RenderToast(msg string, isError bool, opacity float64, width int) string {
	base, target := ToastInfoStyle, Subdued
	if isError {
		base, target = ToastErrorStyle, Accent
	}
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}

	if hex, ok := resolveHex(target); ok && opacity < 1 {
		bg := "#e4e4e4"
		if IsDark {
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

// resolveHex extracts a hex string from a palette color for RGB blending.
// ANSI-indexed colors have no hex to blend, so they report false and skip the
// fade.
func resolveHex(c color.Color) (string, bool) {
	if rgb, ok := c.(color.RGBA); ok {
		return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B), true
	}
	return "", false
}

// End "Selection-list helpers"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
