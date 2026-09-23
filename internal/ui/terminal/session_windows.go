//go:build windows

package terminal

import "os/exec"

// setSessionLeader is a no-op on Windows: ConPTY attaches the console itself.
func setSessionLeader(*exec.Cmd) {}
