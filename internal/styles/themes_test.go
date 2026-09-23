package styles

import (
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
)

func TestApplyRecolorsPaletteStyles(t *testing.T) {
	t.Cleanup(func() { Apply("default", true, nil) })

	Apply("mono", true, nil)
	mono := themes["mono"]
	if got := TerminalFrameStyle.GetBorderTopForeground(); got != mono.Purple.dark {
		t.Errorf("TerminalFrameStyle border = %v, want mono's %v", got, mono.Purple.dark)
	}
	if got := MenuTitle.GetBackground(); got != mono.DarkPurple.dark {
		t.Errorf("MenuTitle background = %v, want %v", got, mono.DarkPurple.dark)
	}
}

func TestApplyPicksBackgroundVariant(t *testing.T) {
	t.Cleanup(func() { Apply("default", true, nil) })

	for _, isDark := range []bool{false, true} {
		Apply("default", isDark, nil)
		want := themes["default"].DullGrey.resolve(isDark)
		if IsDark != isDark {
			t.Errorf("Apply(isDark=%v) left IsDark = %v", isDark, IsDark)
		}
		if got := UnselectedButton.GetBackground(); got != want {
			t.Errorf("isDark=%v: UnselectedButton background = %v, want %v", isDark, got, want)
		}
	}
}

func TestWidgetsFollowBackground(t *testing.T) {
	t.Cleanup(func() { Apply("default", true, nil) })

	for _, isDark := range []bool{false, true} {
		Apply("default", isDark, nil)
		wantRow := list.NewDefaultItemStyles(isDark).NormalTitle.GetForeground()
		if got := NewDelegate().Styles.NormalTitle.GetForeground(); got != wantRow {
			t.Errorf("isDark=%v: delegate title = %v, want %v", isDark, got, wantRow)
		}
		wantStatus := list.DefaultStyles(isDark).StatusBar.GetForeground()
		l := NewList(nil, NewDelegate(), 10, 10)
		if got := l.Styles.StatusBar.GetForeground(); got != wantStatus {
			t.Errorf("isDark=%v: list status bar = %v, want %v", isDark, got, wantStatus)
		}
		wantInput := textinput.DefaultStyles(isDark).Focused.Placeholder.GetForeground()
		ti := NewTextInput()
		if got := ti.Styles().Focused.Placeholder.GetForeground(); got != wantInput {
			t.Errorf("isDark=%v: text input placeholder = %v, want %v", isDark, got, wantInput)
		}
		if got := FormTheme.Theme(!isDark); got.Focused.Title.GetForeground() !=
			FormTheme.Theme(isDark).Focused.Title.GetForeground() {
			t.Errorf("isDark=%v: FormTheme follows huh's guess instead of IsDark", isDark)
		}
	}
}

func TestAccentReplacesThemeAndWidgetAccents(t *testing.T) {
	t.Cleanup(func() { Apply("default", true, nil) })

	accent, err := ParseAccent("#123456")
	if err != nil {
		t.Fatal(err)
	}
	Apply("default", true, accent)

	for name, got := range map[string]any{
		"MenuTitle background":   MenuTitle.GetBackground(),
		"FormBoxStyle border":    FormBoxStyle.GetBorderTopForeground(),
		"selected row title":     NewDelegate().Styles.SelectedTitle.GetForeground(),
		"selected row border":    NewDelegate().Styles.SelectedDesc.GetBorderLeftForeground(),
		"list title background":  NewList(nil, NewDelegate(), 10, 10).Styles.Title.GetBackground(),
		"form field title":       FormTheme.Theme(true).Focused.Title.GetForeground(),
		"form focused button bg": FormTheme.Theme(true).Focused.FocusedButton.GetBackground(),
	} {
		if got != accent {
			t.Errorf("%s = %v, want the accent %v", name, got, accent)
		}
	}

	Apply("default", true, nil)
	want := list.NewDefaultItemStyles(true).SelectedTitle.GetForeground()
	if got := NewDelegate().Styles.SelectedTitle.GetForeground(); got != want {
		t.Errorf("clearing the accent left the selected row %v, want the default %v", got, want)
	}
}

func TestParseAccent(t *testing.T) {
	if c, err := ParseAccent(""); c != nil || err != nil {
		t.Errorf("ParseAccent(\"\") = %v, %v, want no accent and no error", c, err)
	}
	for _, bad := range []string{"123456", "#12345", "#zzzzzz", "blue"} {
		if _, err := ParseAccent(bad); err == nil {
			t.Errorf("ParseAccent(%q) = nil error, want one", bad)
		}
	}
}

func TestDefaultThemeAppliedAtInit(t *testing.T) {
	if DarkPink != themes["default"].DarkPink.dark {
		t.Error("package colors aren't the default dark theme before Apply is called")
	}
	if FormBoxStyle.GetBorderTopForeground() == nil {
		t.Error("palette styles are unset until Apply is called")
	}
}
