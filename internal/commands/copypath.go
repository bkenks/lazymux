package commands

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func CopyPathCmd(absPath string) tea.Cmd {
	return func() tea.Msg {
		if absPath == "" {
			return events.Toast{Level: events.ToastError, Msg: "no repo path to copy"}
		}
		if err := clipboard.WriteAll(absPath); err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("clipboard failed: %v", err)}
		}
		return events.Toast{Level: events.ToastInfo, Msg: "copied " + absPath}
	}
}
