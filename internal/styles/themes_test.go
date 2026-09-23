package styles

import (
	"testing"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
)

func TestApplyRecolorsPaletteStyles(t *testing.T) {
	t.Cleanup(func() { Apply("default", true) })

	Apply("mono", true)
	mono := themes["mono"]
	if got := TerminalFrameStyle.GetBorderTopForeground(); got != mono.Purple.dark {
		t.Errorf("TerminalFrameStyle border = %v, want mono's %v", got, mono.Purple.dark)
	}
	if got := MenuTitle.GetBackground(); got != mono.DarkPurple.dark {
		t.Errorf("MenuTitle background = %v, want %v", got, mono.DarkPurple.dark)
	}
}

func TestApplyPicksBackgroundVariant(t *testing.T) {
	t.Cleanup(func() { Apply("default", true) })

	for _, isDark := range []bool{false, true} {
		Apply("default", isDark)
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
	t.Cleanup(func() { Apply("default", true) })

	for _, isDark := range []bool{false, true} {
		Apply("default", isDark)
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

func TestDefaultThemeAppliedAtInit(t *testing.T) {
	if DarkPink != themes["default"].DarkPink.dark {
		t.Error("package colors aren't the default dark theme before Apply is called")
	}
	if FormBoxStyle.GetBorderTopForeground() == nil {
		t.Error("palette styles are unset until Apply is called")
	}
}
