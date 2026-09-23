//go:build !windows

package terminal

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/charmbracelet/x/ansi"
)

func startSession(t *testing.T, command, dir string) *Model {
	t.Helper()
	constants.WindowSize = tea.WindowSizeMsg{Width: 80, Height: 24}
	m, err := New("test", commands.ShellCommand(command, dir))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(m.close)
	return m
}

func screenText(m *Model) string {
	return ansi.Strip(m.emulator.Render())
}

func waitForScreen(t *testing.T, m *Model, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(screenText(m), want) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("screen never showed %q; got:\n%s", want, screenText(m))
}

func typeText(m *Model, text string) {
	for _, r := range text {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func TestRunsCommandInDir(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := startSession(t, `basename "$(pwd)"`, dir)
	waitForScreen(t, m, filepath.Base(dir))
}

func TestForwardsKeysToProcess(t *testing.T) {
	m := startSession(t, `read line; echo "got:$line"`, t.TempDir())
	typeText(m, "hi")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	waitForScreen(t, m, "got:hi")
}

func TestOutputWaitEndsAfterEsc(t *testing.T) {
	m := startSession(t, "sleep 30", t.TempDir())
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	result := make(chan tea.Msg, 1)
	go func() {
		for {
			if msg := m.waitForOutput()(); msg == nil {
				result <- msg
				return
			}
		}
	}()
	select {
	case <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("waiting for output never ended after esc; the goroutine leaks")
	}
}

func TestStaysOpenAfterExitUntilEsc(t *testing.T) {
	m := startSession(t, "echo done; exit 3", t.TempDir())
	waitForScreen(t, m, "done")

	m.Update(m.waitForExit()())
	if !m.hasExited || m.exitErr == nil {
		t.Fatalf("hasExited=%v exitErr=%v; want exited with an error", m.hasExited, m.exitErr)
	}
	if view := m.View().Content; !strings.Contains(view, "exit status 3") ||
		!strings.Contains(view, "esc to return") {
		t.Errorf("view missing exit status or return hint:\n%s", view)
	}

	if _, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape}); cmd == nil {
		t.Error("esc returned no command; want a switch back to the repo list")
	}
}
