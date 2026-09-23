package mcp

import (
	"path/filepath"

	"github.com/bkenks/lazymux/internal/config"
)

// The daemon's bookkeeping files sit beside the config, so $LAZYMUX_CONFIG
// keeps a dev instance fully separate from the normal one.
const (
	pidFileName = ".lazymux-mcp.pid"
	logFileName = ".lazymux-mcp.log"
)

func stateDir() string { return filepath.Dir(config.Path()) }

// PIDPath is the file holding the running server's process id.
func PIDPath() string { return filepath.Join(stateDir(), pidFileName) }

// LogPath is where a detached server's output is appended.
func LogPath() string { return filepath.Join(stateDir(), logFileName) }
