package commands

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
)

func resolveShell() string {
	if s := cfg().Tools.Shell; s != "" {
		return s
	}
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/sh"
}

// ShellCommand builds the user's shell running command in dir.
func ShellCommand(command, dir string) *exec.Cmd {
	cmd := exec.Command(resolveShell(), "-c", command)
	cmd.Dir = dir
	return cmd
}

func OpenShellCmd(absPath string) tea.Cmd {
	if absPath == "" {
		return func() tea.Msg {
			return events.Toast{Level: events.ToastError, Msg: "no repo path to open shell in"}
		}
	}

	shell := resolveShell()
	cmdBuilder := exec.Command(shell)
	cmdBuilder.Dir = absPath

	return tea.ExecProcess(
		cmdBuilder,
		func(err error) tea.Msg { return events.CmdComplete{Err: err} },
	)
}
