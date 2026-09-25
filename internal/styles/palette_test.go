package styles

import (
	"image/color"
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

func lightness(t *testing.T, c color.Color) float64 {
	t.Helper()
	cc, ok := colorful.MakeColor(c)
	if !ok {
		t.Fatalf("%v is not a color", c)
	}
	l, _, _ := cc.OkLch()
	return l
}

func TestRampRunsFromTintToShadeThroughTheBase(t *testing.T) {
	for _, hex := range []string{"#EE6FF8", "#123456", "#777777", "#FFFF00", "#000000"} {
		base := mustHex(hex)
		r := newRamp(base)
		if r[baseStep] != lipgloss.Color(base.Hex()) {
			t.Errorf("%s: base step = %v, want the base itself", hex, r[baseStep])
		}
		for i := 1; i < rampSize; i++ {
			if lightness(t, r[i]) > lightness(t, r[i-1])+0.005 {
				t.Errorf("%s: step %d is lighter than step %d", hex, i, i-1)
			}
		}
		if lightness(t, r[0])-lightness(t, r[rampSize-1]) < 0.5 {
			t.Errorf("%s: ramp spans %v to %v, want tint to shade", hex, r[0], r[rampSize-1])
		}
	}
}

func TestRampKeepsTheBaseHue(t *testing.T) {
	base := mustHex("#EE6FF8")
	_, _, wantHue := base.OkLch()
	for i, c := range newRamp(base) {
		cc, _ := colorful.MakeColor(c)
		if _, chroma, hue := cc.OkLch(); chroma > 0.02 && (hue-wantHue > 10 || wantHue-hue > 10) {
			t.Errorf("step %d hue = %.1f, want near the base's %.1f", i, hue, wantHue)
		}
	}
}

func TestApplyDerivesColorsFromEachBase(t *testing.T) {
	t.Cleanup(func() { Apply(DefaultPalette, true) })

	p, err := NewPalette("#123456", "#AA3300", "#808070")
	if err != nil {
		t.Fatal(err)
	}
	Apply(p, true)

	for name, got := range map[string]any{
		"MenuTitle background":   MenuTitle.GetBackground(),
		"list title background":  NewList(nil, NewDelegate(), 10, 10).Styles.Title.GetBackground(),
		"form focused button bg": FormTheme.Theme(true).Focused.FocusedButton.GetBackground(),
	} {
		if want := lipgloss.Color(p.Main.Hex()); got != want {
			t.Errorf("%s = %v, want the main color %v", name, got, want)
		}
	}
	for name, got := range map[string]any{
		"selected row title": NewDelegate().Styles.SelectedTitle.GetForeground(),
		"repo path":          DialogRepoPath.GetForeground(),
		"filter prompt":      NewList(nil, NewDelegate(), 10, 10).Styles.Filter.Focused.Prompt.GetForeground(),
		"form selector":      FormTheme.Theme(true).Focused.SelectSelector.GetForeground(),
	} {
		if want := lipgloss.Color(p.Accent.Hex()); got != want {
			t.Errorf("%s = %v, want the accent %v", name, got, want)
		}
	}
	accent, gray := newRamp(p.Accent), newRamp(p.Gray)
	if got := NewDelegate().Styles.SelectedDesc.GetForeground(); got != accent[7] {
		t.Errorf("selected row desc = %v, want a darker accent %v", got, accent[7])
	}
	if got := NewDelegate().Styles.NormalDesc.GetForeground(); got != gray[baseStep] {
		t.Errorf("row desc = %v, want the gray %v", got, gray[baseStep])
	}
}

func TestApplyPicksBackgroundVariant(t *testing.T) {
	t.Cleanup(func() { Apply(DefaultPalette, true) })

	Apply(DefaultPalette, true)
	darkText := NewDelegate().Styles.NormalTitle.GetForeground()
	Apply(DefaultPalette, false)
	lightText := NewDelegate().Styles.NormalTitle.GetForeground()
	if IsDark {
		t.Error("Apply(isDark=false) left IsDark true")
	}
	if lightness(t, darkText) < 0.8 || lightness(t, lightText) > 0.4 {
		t.Errorf("row text is %v on dark and %v on light, want light on dark and dark on light",
			darkText, lightText)
	}
}

func TestOnMainIsReadable(t *testing.T) {
	t.Cleanup(func() { Apply(DefaultPalette, true) })

	for _, hex := range []string{"#FFFF99", "#1A1A40"} {
		p := DefaultPalette
		p.Main = mustHex(hex)
		Apply(p, true)
		if gap := lightness(t, OnMain) - lightness(t, Main); gap < 0.4 && gap > -0.4 {
			t.Errorf("main %s: title text %v is too close to its background", hex, OnMain)
		}
	}
}

func TestWidgetsFollowBackground(t *testing.T) {
	t.Cleanup(func() { Apply(DefaultPalette, true) })

	for _, isDark := range []bool{false, true} {
		Apply(DefaultPalette, isDark)
		wantInput := textinput.DefaultStyles(isDark).Focused.Placeholder.GetForeground()
		ti := NewTextInput()
		if got := ti.Styles().Focused.Placeholder.GetForeground(); got != wantInput {
			t.Errorf("isDark=%v: text input placeholder = %v, want %v", isDark, got, wantInput)
		}
		if got := FormTheme.Theme(!isDark); got.Focused.Title.GetForeground() !=
			FormTheme.Theme(isDark).Focused.Title.GetForeground() {
			t.Errorf("isDark=%v: FormTheme follows huh's guess instead of IsDark", isDark)
		}
		wantPadding := list.DefaultStyles(isDark).StatusBar.GetPaddingLeft()
		if got := NewList(nil, NewDelegate(), 10, 10).Styles.StatusBar.GetPaddingLeft(); got != wantPadding {
			t.Errorf("isDark=%v: list status bar lost its layout", isDark)
		}
	}
}

func TestNewPalette(t *testing.T) {
	p, err := NewPalette("", "", "")
	if err != nil || p != DefaultPalette {
		t.Errorf("NewPalette with no colors = %v, %v, want the default and no error", p, err)
	}

	p, err = NewPalette("#75F", "blue", "")
	if err == nil {
		t.Error("NewPalette with an invalid accent returned no error")
	}
	if p.Main != mustHex("#7755FF") || p.Accent != DefaultPalette.Accent {
		t.Errorf("NewPalette kept %v, want the valid main and the default accent", p)
	}
}

func TestParseColor(t *testing.T) {
	for _, bad := range []string{"", "123456", "#12345", "#zzzzzz", "blue"} {
		if _, err := ParseColor(bad); err == nil {
			t.Errorf("ParseColor(%q) = nil error, want one", bad)
		}
	}
}

func TestDefaultPaletteAppliedAtInit(t *testing.T) {
	if Main != lipgloss.Color(DefaultPalette.Main.Hex()) {
		t.Error("package colors aren't the default palette before Apply is called")
	}
	if FormBoxStyle.GetBorderTopForeground() == nil {
		t.Error("palette styles are unset until Apply is called")
	}
}
