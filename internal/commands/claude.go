package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

var claudeAttachID = regexp.MustCompile(`claude attach ([0-9a-zA-Z]+)`)

func resolveClaudeWorkDir() string {
	if dir := cfg().Tools.ClaudeWorkDir; dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}

func OpenClaudeAgentsCmd() tea.Cmd {
	return execClaude(resolveClaudeWorkDir(), "agents")
}

func NewClaudeSessionCmd(absPath string) tea.Cmd {
	start := exec.Command("claude", "--bg", "-n", filepath.Base(absPath))
	start.Dir = absPath
	out, err := start.CombinedOutput()
	if err != nil {
		return claudeErrorToast(fmt.Sprintf("claude session failed: %v", err))
	}
	match := claudeAttachID.FindSubmatch(out)
	if match == nil {
		return claudeErrorToast("claude session started but its id was not found")
	}
	return execClaude(absPath, "attach", string(match[1]))
}

func execClaude(dir string, args ...string) tea.Cmd {
	cmd := exec.Command("claude", args...)
	cmd.Dir = dir
	return tea.ExecProcess(cmd, func(err error) tea.Msg { return events.CmdComplete{Err: err} })
}

func claudeErrorToast(msg string) tea.Cmd {
	return func() tea.Msg { return events.Toast{Level: events.ToastError, Msg: msg} }
}
