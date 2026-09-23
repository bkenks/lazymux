//go:build windows

package mcp

import (
	"errors"

	"github.com/bkenks/lazymux/internal/config"
)

var errDaemonUnsupported = errors.New("daemon mode is unix-only; use `lazymux mcp serve`")

// Start is unsupported on Windows, which lacks the signals and file locks the
// background server relies on. Run `lazymux mcp serve` under a supervisor.
func Start(config.Config) error { return errDaemonUnsupported }

// Stop is unsupported on Windows; see Start.
func Stop() error { return errDaemonUnsupported }

// Running always reports no background server on Windows, where none can be
// started.
func Running() int { return 0 }

// claimPID records nothing on Windows: without start and stop, no command
// reads the pidfile.
func claimPID() (func(), error) { return func() {}, nil }
