package settings

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
)

// stubEditorOnPath writes an executable named name into a temp dir and makes
// that dir the only entry on PATH for the duration of the test.
func stubEditorOnPath(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("writing stub editor: %v", err)
	}
	t.Setenv("PATH", dir)
	return path
}

func TestValidateEditorCommandAcceptsCommandOnPath(t *testing.T) {
	want := stubEditorOnPath(t, "zed")

	if err := validateEditorCommand("zed"); err != nil {
		t.Fatalf("validateEditorCommand(zed) = %v, want no error", err)
	}
	if got := describeEditor("zed"); got != "✓ "+want {
		t.Errorf("describeEditor(zed) = %q, want the resolved path %q", got, want)
	}
}

func TestValidateEditorCommandAcceptsAbsolutePath(t *testing.T) {
	path := stubEditorOnPath(t, "zed")

	if err := validateEditorCommand(path); err != nil {
		t.Fatalf("validateEditorCommand(%q) = %v, want no error", path, err)
	}
}

func TestValidateEditorCommandRejects(t *testing.T) {
	stubEditorOnPath(t, "zed")

	for _, bad := range []string{"", "  ", "definitely-not-installed", "zed --wait"} {
		if err := validateEditorCommand(bad); err == nil {
			t.Errorf("validateEditorCommand(%q) = nil, want an error", bad)
		}
	}
}

func TestValidateAccentColor(t *testing.T) {
	for _, ok := range []string{"", "#7D56F4", "#7d56f4", "#abc", " #7D56F4 "} {
		if err := validateAccentColor(ok); err != nil {
			t.Errorf("validateAccentColor(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"7D56F4", "#7D56F", "#GGGGGG", "purple", "#7D56F4FF"} {
		if err := validateAccentColor(bad); err == nil {
			t.Errorf("validateAccentColor(%q) = nil, want an error", bad)
		}
	}
}

// press sends msg and feeds every resulting message back into the model, the
// way the bubbletea runtime would, returning the messages the screen emits
// for the app: SettingsChanged and SetState.
func press(m *Model, msg tea.Msg) []tea.Msg {
	var emitted []tea.Msg
	queue := []tea.Msg{msg}
	for steps := 0; len(queue) > 0 && steps < 200; steps++ {
		next := queue[0]
		queue = queue[1:]
		switch next.(type) {
		case events.SettingsChanged, events.SetState:
			emitted = append(emitted, next)
			continue
		}
		_, cmd := m.Update(next)
		queue = append(queue, runCmd(cmd)...)
	}
	return emitted
}

// runCmd runs cmd, dropping it if it hasn't returned within a short wait —
// those are timers such as the cursor blink, which the tests don't need.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	result := make(chan tea.Msg, 1)
	go func() { result <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-result:
	case <-time.After(50 * time.Millisecond):
		return nil
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var msgs []tea.Msg
	for _, inner := range batch {
		msgs = append(msgs, runCmd(inner)...)
	}
	return msgs
}

var enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}

// fillAccent presses enter past every field before the accent color, types
// accent, and submits, returning what the screen emitted.
func fillAccent(m *Model, accent string) []tea.Msg {
	for range 7 {
		press(m, enterKey)
	}
	for _, r := range accent {
		press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return press(m, enterKey)
}

func newTestModel(t *testing.T) *Model {
	t.Helper()
	stubEditorOnPath(t, "zed")
	cfg := config.Default()
	cfg.Tools.Editor = "zed"
	m := New(cfg)
	press(m, m.Init())
	return m
}

func TestSubmitEmitsEditedSettings(t *testing.T) {
	m := newTestModel(t)

	emitted := fillAccent(m, "#7D56F4")

	if len(emitted) != 1 {
		t.Fatalf("emitted %v, want one SettingsChanged", emitted)
	}
	changed, ok := emitted[0].(events.SettingsChanged)
	if !ok {
		t.Fatalf("emitted %T, want SettingsChanged", emitted[0])
	}
	if got := changed.Config.UI.AccentColor; got != "#7D56F4" {
		t.Errorf("accent color = %q, want #7D56F4", got)
	}
	if got := changed.Config.Tools.Editor; got != "zed" {
		t.Errorf("editor = %q, want zed kept", got)
	}
}

func TestInvalidAccentBlocksSubmit(t *testing.T) {
	m := newTestModel(t)

	if emitted := fillAccent(m, "purple"); len(emitted) != 0 {
		t.Errorf("emitted %v for an invalid accent, want nothing", emitted)
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
