package settings

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/ui/formtest"
)

func TestValidateColor(t *testing.T) {
	for _, ok := range []string{"", "#7D56F4", "#7d56f4", "#abc", " #7D56F4 "} {
		if err := validateColor(ok); err != nil {
			t.Errorf("validateColor(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"7D56F4", "#7D56F", "#GGGGGG", "purple", "#7D56F4FF"} {
		if err := validateColor(bad); err == nil {
			t.Errorf("validateColor(%q) = nil, want an error", bad)
		}
	}
}

// press sends msg and returns what the screen emits for the app:
// SettingsChanged and SetState.
func press(m *Model, msg tea.Msg) []tea.Msg {
	return formtest.Press(m, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case events.SettingsChanged, events.SetState:
			return true
		}
		return false
	})
}

var enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}

// fillColors presses enter past every field before the colors, types the
// dark mode main, accent and gray colors and then the light mode ones, and
// submits, returning what the screen emitted.
func fillColors(m *Model, dark, light [3]string) []tea.Msg {
	for range 6 {
		press(m, enterKey)
	}
	var emitted []tea.Msg
	for _, color := range append(dark[:], light[:]...) {
		for _, r := range color {
			press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		emitted = press(m, enterKey)
	}
	return emitted
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	m := New(config.Default())
	press(m, m.Init())
	return m
}

func TestSubmitEmitsEditedSettings(t *testing.T) {
	m := newTestModel(t)

	emitted := fillColors(m, [3]string{"#7D56F4", " #EE6FF8 ", ""}, [3]string{"", "", "#333"})

	if len(emitted) != 1 {
		t.Fatalf("emitted %v, want one SettingsChanged", emitted)
	}
	changed, ok := emitted[0].(events.SettingsChanged)
	if !ok {
		t.Fatalf("emitted %T, want SettingsChanged", emitted[0])
	}
	want := config.ColorModes{
		Dark:  config.Colors{Main: "#7D56F4", Accent: "#EE6FF8"},
		Light: config.Colors{Gray: "#333"},
	}
	if got := changed.Config.UI.Colors; got != want {
		t.Errorf("colors = %+v, want %+v", got, want)
	}
}

func TestInvalidColorBlocksSubmit(t *testing.T) {
	m := newTestModel(t)

	if emitted := fillColors(m, [3]string{}, [3]string{"", "purple", ""}); len(emitted) != 0 {
		t.Errorf("emitted %v for an invalid light accent, want nothing", emitted)
	}
}

func TestEscLeavesWithoutSaving(t *testing.T) {
	m := newTestModel(t)

	emitted := press(m, tea.KeyPressMsg{Code: tea.KeyEscape})

	if len(emitted) != 1 {
		t.Fatalf("emitted %v, want one SetState", emitted)
	}
	if state, ok := emitted[0].(events.SetState); !ok || state.State != domain.StateMain {
		t.Errorf("emitted %v, want SetState to the repo list", emitted[0])
	}
}
