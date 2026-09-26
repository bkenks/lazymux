package commands

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestHandoffWaitsForEnterAndKeepsTheCommandError(t *testing.T) {
	var output bytes.Buffer
	paused := &keybindHandoff{
		cmd: exec.Command("sh", "-c", "echo hello; exit 3"), shouldPause: true,
	}
	paused.SetStdin(strings.NewReader("\nleftover"))
	paused.SetStdout(&output)
	paused.SetStderr(&output)

	err := paused.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Errorf("Run() = %v, want the command's exit status 3", err)
	}
	got := ansi.Strip(output.String())
	want := []string{"loading…", "hello", "press enter to return to gitkeeper"}
	for _, part := range want {
		index := strings.Index(got, part)
		if index < 0 {
			t.Fatalf("output = %q, want %q in it, in order %q", got, part, want)
		}
		got = got[index+len(part):]
	}
	if !strings.HasSuffix(output.String(), ansi.EraseEntireScreen+ansi.CursorHomePosition) {
		t.Errorf("output = %q, want it to end by clearing the loading screen", output.String())
	}
}
