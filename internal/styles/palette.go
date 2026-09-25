package styles

import (
	"errors"
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// Palette is the three base colors every UI color is derived from: Main for
// title bars and buttons, Accent for the selection and highlights, and Gray
// for text, borders and hints.
type Palette struct {
	Main, Accent, Gray colorful.Color
}

// DefaultPalette is the palette used for any base color the config leaves
// empty.
var DefaultPalette = Palette{
	Main:   mustHex("#5F5FD7"),
	Accent: mustHex("#EE6FF8"),
	Gray:   mustHex("#777777"),
}

func mustHex(hex string) colorful.Color {
	c, err := colorful.Hex(hex)
	if err != nil {
		panic(err)
	}
	return c
}

// ParseColor reads a base color written as a hex value, "#RRGGBB" or "#RGB".
func ParseColor(hex string) (colorful.Color, error) {
	c, err := colorful.Hex(hex)
	if err != nil {
		return colorful.Color{}, fmt.Errorf("%q is not a hex color like #7D56F4", hex)
	}
	return c, nil
}

// NewPalette builds a palette from hex base colors. An empty value keeps the
// default for that color, and so does an invalid one, whose error is returned
// alongside the palette.
func NewPalette(main, accent, gray string) (Palette, error) {
	p := DefaultPalette
	var errs []error
	for _, base := range []struct {
		name, hex string
		color     *colorful.Color
	}{
		{"main", main, &p.Main},
		{"accent", accent, &p.Accent},
		{"gray", gray, &p.Gray},
	} {
		if base.hex == "" {
			continue
		}
		c, err := ParseColor(base.hex)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s color: %w", base.name, err))
			continue
		}
		*base.color = c
	}
	return p, errors.Join(errs...)
}

const (
	rampSize = 11
	baseStep = 5
)

// ramp is a base color's scale from a pale tint (0) to a deep shade (10), with
// the base itself at baseStep, blended in OkLCh so every step keeps its hue. A
// base paler than the tint or deeper than the shade flattens that half.
type ramp [rampSize]color.Color

func newRamp(base colorful.Color) ramp {
	lightness, chroma, hue := base.OkLch()
	tint := colorful.OkLch(max(0.97, lightness), chroma*0.15, hue)
	shade := colorful.OkLch(min(0.18, lightness), chroma*0.4, hue)

	var r ramp
	for i := range r {
		var c colorful.Color
		switch {
		case i < baseStep:
			c = tint.BlendOkLch(base, float64(i)/baseStep)
		case i == baseStep:
			c = base
		default:
			c = base.BlendOkLch(shade, float64(i-baseStep)/(rampSize-1-baseStep))
		}
		r[i] = lipgloss.Color(c.Clamped().Hex())
	}
	return r
}

// step picks the ramp step for a dark or a light terminal background.
func (r ramp) step(isDark bool, dark, light int) color.Color {
	if isDark {
		return r[dark]
	}
	return r[light]
}

// readableOn picks the end of r that contrasts more with base.
func (r ramp) readableOn(base colorful.Color) color.Color {
	if lightness, _, _ := base.OkLch(); lightness > 0.65 {
		return r[rampSize-1]
	}
	return r[0]
}

// IsDark reports whether the terminal background is dark. Apply sets it, and
// screens pass it to the bubbles and huh defaults they build.
var IsDark = true

func init() { Apply(DefaultPalette, true) }

// Apply derives every UI color from p for a light or dark terminal background
// and rebuilds every style that depends on them. Screens built earlier keep
// their styles until they are rebuilt or restyled.
func Apply(p Palette, isDark bool) {
	IsDark = isDark
	main, accent, gray := newRamp(p.Main), newRamp(p.Accent), newRamp(p.Gray)

	Main = main[baseStep]
	OnMain = main.readableOn(p.Main)
	MainText = main.step(isDark, 3, 6)
	Accent = accent[baseStep]
	AccentMuted = accent.step(isDark, 7, 3)
	Text = gray.step(isDark, 1, 10)
	Muted = gray[baseStep]
	Subdued = gray.step(isDark, 6, 4)
	Faint = gray.step(isDark, 7, 3)
	Surface = gray.step(isDark, 7, 2)
	OnSurface = gray.step(isDark, 2, 8)

	rebuildStyles()
}

func rebuildStyles() {
	MenuTitle = lipgloss.NewStyle().
		Background(Main).
		Foreground(OnMain).
		Padding(0, 1).
		Margin(1, 0, 1, 2)

	MenuSubStyle = lipgloss.NewStyle().
		Foreground(Muted).
		MarginLeft(2).
		MarginBottom(1)

	SelectedButton = ButtonStyle.
		Background(Main).
		Foreground(OnMain).
		Bold(true)

	UnselectedButton = ButtonStyle.
		Background(Surface).
		Foreground(OnSurface)

	DialogStyle = lipgloss.NewStyle().
		Padding(1, 6, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Subdued)

	FormBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(MainText).
		Padding(0, 1).
		MarginTop(1)

	DialogTitleStyle = lipgloss.NewStyle().
		Background(Main).
		Foreground(OnMain).
		Padding(0, 1).
		Margin(0, 0, 2)

	DialogRepoPath = lipgloss.NewStyle().
		Bold(true).
		MarginBottom(2).
		Foreground(Accent)

	ToastInfoStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(Subdued)

	ToastErrorStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(Accent).
		Bold(true)

	Help = newHelpModel()
}
