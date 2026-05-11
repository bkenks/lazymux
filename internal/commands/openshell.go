package commands

import (
	"os"
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
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
