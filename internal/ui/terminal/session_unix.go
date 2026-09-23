//go:build !windows

package terminal

import (
	"os/exec"
	"syscall"

	"github.com/charmbracelet/x/xpty"
)

// setSessionLeader gives the process its own session with the pty as its
// controlling terminal, so job control and ctrl+c reach it.
func setSessionLeader(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
}

// releaseChildSide closes lazymux's copy of the pty's child end once the child
// holds it, so reads from the pty end after the child exits and its last
// output has been read.
func releaseChildSide(pty xpty.Pty) {
	if unixPty, ok := pty.(*xpty.UnixPty); ok {
		_ = unixPty.Slave().Close()
	}
}
