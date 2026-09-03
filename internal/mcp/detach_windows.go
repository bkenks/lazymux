//go:build windows

package mcp

import (
	"os/exec"
	"syscall"
)

// detach gives the server child its own process group, so a Ctrl+C in the
// console that started it doesn't take the server down too. Windows has no
// sessions to leave, so there is nothing else to detach from.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
