package keybinds

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/events"
)

func newTestModel() *Model {
	cfg := config.Default()
	cfg.Keybinds = []config.Keybind{
		{Name: "Git log", Keys: "ctrl+l", Command: "git log"},
		{Name: "Status", Keys: "ctrl+g", Command: "git status"},
	}
	return New(cfg, []string{"n", "S", "ctrl+c"})
}

func TestValidateKeys(t *testing.T) {
	m := newTestModel()
	m.editIndex = -1

	for _, taken := range []string{"n", "shift + s", "ctrl + c", "ctrl + l"} {
		if err := m.validateKeys(taken); err == nil {
			t.Errorf("validateKeys(%q) = nil, want a clash error", taken)
		}
	}
	for _, bad := range []string{"", "ctrl", "ctrl + banana"} {
		if err := m.validateKeys(bad); err == nil {
			t.Errorf("validateKeys(%q) = nil, want a parse error", bad)
		}
	}
	if err := m.validateKeys("ctrl + p"); err != nil {
		t.Errorf("validateKeys(ctrl + p) = %v, want nil", err)
	}

	m.editIndex = 0
	if err := m.validateKeys("ctrl + l"); err != nil {
		t.Errorf("editing a keybind rejected its own keys: %v", err)
	}
}

// press sends msg and feeds every resulting message back into the model, the
// way the bubbletea runtime would, returning any KeybindsChanged emitted.
func press(m *Model, msg tea.Msg) []events.KeybindsChanged {
	var changed []events.KeybindsChanged
	queue := []tea.Msg{msg}
	for steps := 0; len(queue) > 0 && steps < 50; steps++ {
		next := queue[0]
		queue = queue[1:]
		if c, ok := next.(events.KeybindsChanged); ok {
			changed = append(changed, c)
			continue
		}
		_, cmd := m.Update(next)
		queue = append(queue, runCmd(cmd)...)
	}
	return changed
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

var (
	deleteKey = tea.KeyPressMsg{Code: '\\', Mod: tea.ModCtrl}
	enterKey  = tea.KeyPressMsg{Code: tea.KeyEnter}
	leftKey   = tea.KeyPressMsg{Code: tea.KeyLeft}
)

func TestDeleteNeedsConfirmation(t *testing.T) {
	m := newTestModel()

	press(m, deleteKey)
	if m.form == nil {
		t.Fatal("ctrl+\\ did not open the delete prompt")
	}
	if changed := press(m, enterKey); len(changed) != 0 || len(m.keybinds) != 2 {
		t.Fatalf("answering the default No deleted a keybind: %v", m.keybinds)
	}
	if m.form != nil {
		t.Fatal("answering No left the prompt open")
	}

	press(m, deleteKey)
	press(m, leftKey)
	changed := press(m, enterKey)
	if len(m.keybinds) != 1 || m.keybinds[0].Name != "Status" {
		t.Fatalf("Yes did not delete the selected keybind: %v", m.keybinds)
	}
	if len(changed) != 1 || len(changed[0].Keybinds) != 1 {
		t.Errorf("delete emitted %v, want one KeybindsChanged with 1 keybind", changed)
	}
}

func TestNewKeybindIsSavedCanonical(t *testing.T) {
	m := newTestModel()

	press(m, tea.KeyPressMsg{Code: 'n', Text: "n"})
	for _, field := range []string{"Tests", "Ctrl + Shift + T", "go test ./..."} {
		for _, r := range field {
			press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		press(m, enterKey)
	}
	press(m, leftKey)
	changed := press(m, enterKey)

	if len(changed) != 1 {
		t.Fatalf("form emitted %d KeybindsChanged, want 1 (form open: %v)", len(changed), m.form != nil)
	}
	want := config.Keybind{
		Name: "Tests", Keys: "ctrl+shift+t", Command: "go test ./...", ReturnOnExit: true,
	}
	if got := m.keybinds[len(m.keybinds)-1]; got != want {
		t.Errorf("saved %+v, want %+v", got, want)
	}
}

func TestReturnOnExitIgnoresYAndN(t *testing.T) {
	m := newTestModel()

	press(m, tea.KeyPressMsg{Code: 'n', Text: "n"})
	for _, field := range []string{"Tests", "ctrl + t", "go test ./..."} {
		for _, r := range field {
			press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		press(m, enterKey)
	}
	for _, r := range "yYnN" {
		if changed := press(m, tea.KeyPressMsg{Code: r, Text: string(r)}); len(changed) != 0 {
			t.Fatalf("pressing %q saved the keybind", r)
		}
	}
	press(m, leftKey)
	press(m, enterKey)

	if got := m.keybinds[len(m.keybinds)-1]; !got.ReturnOnExit {
		t.Errorf("saved %+v, want ReturnOnExit toggled on by left", got)
	}
}

func TestSaveKeyNeedsEveryFieldValid(t *testing.T) {
	m := newTestModel()
	saveKey := tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}

	press(m, tea.KeyPressMsg{Code: 'n', Text: "n"})
	for _, r := range "Tests" {
		press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if changed := press(m, saveKey); len(changed) != 0 || m.form == nil {
		t.Fatalf("saved a keybind with no keys or command: %v", changed)
	}

	press(m, enterKey)
	for _, field := range []string{"ctrl + t", "go test ./..."} {
		for _, r := range field {
			press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		press(m, enterKey)
	}
	changed := press(m, saveKey)

	if len(changed) != 1 {
		t.Fatalf("form emitted %d KeybindsChanged, want 1 (form open: %v)", len(changed), m.form != nil)
	}
	want := config.Keybind{Name: "Tests", Keys: "ctrl+t", Command: "go test ./..."}
	if got := m.keybinds[len(m.keybinds)-1]; got != want {
		t.Errorf("saved %+v, want %+v", got, want)
	}
}
