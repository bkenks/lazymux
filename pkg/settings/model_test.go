package settings

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func onlyAccepts(want string) Validator {
	return func(candidate string) (string, error) {
		if candidate != want {
			return "", errors.New("nope")
		}
		return "resolved " + candidate, nil
	}
}

func newTextModel(t *testing.T, value string) *Model {
	t.Helper()
	m := New("Settings", []Setting{NewText("editor", "Editor", value, onlyAccepts("ok"))}, 80, 24, 0, 0)
	return &m
}

func press(t *testing.T, m *Model, msg tea.KeyMsg) tea.Cmd {
	t.Helper()
	_, cmd := m.Update(msg)
	return cmd
}

func typeRunes(t *testing.T, m *Model, s string) {
	t.Helper()
	for _, r := range s {
		press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

// changedMsg unwraps the SettingChanged a command produces, descending into a
// tea.Batch because Update batches the list's own command alongside it.
func changedMsg(cmd tea.Cmd) (SettingChanged, bool) {
	if cmd == nil {
		return SettingChanged{}, false
	}
	switch msg := cmd().(type) {
	case SettingChanged:
		return msg, true
	case tea.BatchMsg:
		for _, sub := range msg {
			if changed, ok := changedMsg(sub); ok {
				return changed, true
			}
		}
	}
	return SettingChanged{}, false
}

func TestEnterOpensEditorOnTextSetting(t *testing.T) {
	m := newTextModel(t, "codium")

	cmd := press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.editing {
		t.Fatal("enter on a Text setting did not open the editor")
	}
	if _, ok := changedMsg(cmd); ok {
		t.Error("opening the editor emitted SettingChanged")
	}
	if got := m.input.Value(); got != "codium" {
		t.Errorf("input seeded with %q, want %q", got, "codium")
	}
}

func TestEditorCommitsValidValue(t *testing.T) {
	m := newTextModel(t, "")
	press(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	typeRunes(t, m, "ok")

	cmd := press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.editing {
		t.Fatal("editor stayed open after a valid commit")
	}
	changed, ok := changedMsg(cmd)
	if !ok {
		t.Fatal("valid commit emitted no SettingChanged")
	}
	if changed.Key != "editor" || changed.Setting.ValueString() != "ok" {
		t.Errorf("SettingChanged{%q, %q}, want {editor, ok}", changed.Key, changed.Setting.ValueString())
	}
	if got := m.Settings()[0].ValueString(); got != "ok" {
		t.Errorf("stored value %q, want ok", got)
	}
}

func TestEditorViewShowsValidationStatus(t *testing.T) {
	m := newTextModel(t, "")
	press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	typeRunes(t, m, "nah")
	if view := m.View(); !strings.Contains(view, "\u2717 nope") {
		t.Errorf("view missing the rejection status, got:\n%s", view)
	}

	press(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	press(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	press(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	typeRunes(t, m, "ok")
	if view := m.View(); !strings.Contains(view, "\u2713 resolved ok") {
		t.Errorf("view missing the acceptance hint, got:\n%s", view)
	}
}

func TestEditorRejectsInvalidValue(t *testing.T) {
	m := newTextModel(t, "")
	press(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	typeRunes(t, m, "nah")

	cmd := press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.editing {
		t.Fatal("editor closed on an invalid commit")
	}
	if _, ok := changedMsg(cmd); ok {
		t.Error("invalid commit emitted SettingChanged")
	}
	if got := m.Settings()[0].ValueString(); got != "" {
		t.Errorf("stored value %q, want it unchanged", got)
	}
}

func TestEscCancelsEditorWithoutSaving(t *testing.T) {
	m := newTextModel(t, "codium")
	press(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	typeRunes(t, m, "ok")

	cmd := press(t, m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.editing {
		t.Fatal("esc did not close the editor")
	}
	if _, ok := changedMsg(cmd); ok {
		t.Error("cancelling emitted SettingChanged")
	}
	if got := m.Settings()[0].ValueString(); got != "codium" {
		t.Errorf("stored value %q, want codium", got)
	}
}

func TestEditorSwallowsKeysBoundOnTheList(t *testing.T) {
	m := newTextModel(t, "")
	press(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	typeRunes(t, m, "hl")

	if got := m.input.Value(); got != "hl" {
		t.Errorf("input holds %q, want %q — list bindings stole the keys", got, "hl")
	}
}

func TestSelectSettingStillCycles(t *testing.T) {
	m := New("Settings", []Setting{NewSelect("protocol", "Protocol", []string{"https", "ssh"}, 0)}, 80, 24, 0, 0)

	cmd := press(t, &m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.editing {
		t.Fatal("enter on a Select opened the text editor")
	}
	changed, ok := changedMsg(cmd)
	if !ok {
		t.Fatal("cycling a Select emitted no SettingChanged")
	}
	if changed.Setting.ValueString() != "ssh" {
		t.Errorf("cycled to %q, want ssh", changed.Setting.ValueString())
	}
}
