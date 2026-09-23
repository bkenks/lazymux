package styles

import "testing"

func TestApplyRecolorsPaletteStyles(t *testing.T) {
	t.Cleanup(func() { Apply("default") })

	Apply("mono")
	mono := themes["mono"]
	if got := TerminalFrameStyle.GetBorderTopForeground(); got != mono.Purple {
		t.Errorf("TerminalFrameStyle border = %v, want the mono palette's %v", got, mono.Purple)
	}
	if got := MenuTitle.GetBackground(); got != mono.DarkPurple {
		t.Errorf("MenuTitle background = %v, want %v", got, mono.DarkPurple)
	}
}

func TestDefaultThemeAppliedAtInit(t *testing.T) {
	if DarkPink != themes["default"].DarkPink {
		t.Error("package colors aren't the default theme before Apply is called")
	}
	if FormBoxStyle.GetBorderTopForeground() == nil {
		t.Error("palette styles are unset until Apply is called")
	}
}
