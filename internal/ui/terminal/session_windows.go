//go:build windows

package terminal

import (
	"os/exec"

	"github.com/charmbracelet/x/xpty"
)

// setSessionLeader is a no-op on Windows: ConPTY attaches the console itself.
func setSessionLeader(*exec.Cmd) {}

// releaseChildSide is a no-op on Windows: ConPTY owns both ends.
func releaseChildSide(xpty.Pty) {}
