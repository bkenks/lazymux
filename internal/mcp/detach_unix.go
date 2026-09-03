//go:build !windows

package mcp

import (
	"os/exec"
	"syscall"
)

// detach puts the server child in its own session, so it survives the shell
// that started it.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
