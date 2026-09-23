//go:build !windows

package terminal

import (
	"os/exec"
	"syscall"
)

// setSessionLeader gives the process its own session with the pty as its
// controlling terminal, so job control and ctrl+c reach it.
func setSessionLeader(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
}
