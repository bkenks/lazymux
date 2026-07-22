package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bkenks/lazymux/internal/config"
)

// The daemon's bookkeeping files sit beside the config, so $LAZYMUX_CONFIG
// keeps a dev instance fully separate from the normal one.
const (
	pidFileName = ".lazymux-mcp.pid"
	logFileName = ".lazymux-mcp.log"

	// startTimeout bounds how long `mcp start` waits for the child to bind.
	startTimeout = 5 * time.Second
	// stopTimeout bounds how long `mcp stop` waits for a graceful exit before
	// escalating to SIGKILL.
	stopTimeout = 5 * time.Second
)

func stateDir() string { return filepath.Dir(config.Path()) }

// PIDPath is the file holding the running server's process id.
func PIDPath() string { return filepath.Join(stateDir(), pidFileName) }

// LogPath is where a detached server's output is appended.
func LogPath() string { return filepath.Join(stateDir(), logFileName) }

// readPID returns the pid recorded in the pidfile, or 0 if there isn't a
// usable one. An unparseable file is cleared rather than left to confuse
// every later command.
func readPID() int {
	data, err := os.ReadFile(PIDPath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		os.Remove(PIDPath())
		return 0
	}
	return pid
}

// Running returns the pid of the live server, or 0 if none is running. A
// pidfile naming a dead process is treated as absent (and removed), so a
// crashed server never blocks the next start.
func Running() int {
	pid := readPID()
	if pid == 0 {
		return 0
	}
	if !alive(pid) {
		os.Remove(PIDPath())
		return 0
	}
	return pid
}

// alive reports whether a process exists. Signal 0 performs the permission
// and existence checks without delivering anything.
func alive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// WritePID records the current process as the running server.
func WritePID() error {
	return os.WriteFile(PIDPath(), []byte(strconv.Itoa(os.Getpid())), 0o644)
}

// RemovePID clears the pidfile.
func RemovePID() { os.Remove(PIDPath()) }

// Start launches a detached server process and waits for it to report that it
// bound the port, so a bind failure surfaces here rather than only in the log.
func Start(cfg config.Config) error {
	if pid := Running(); pid != 0 {
		return fmt.Errorf("already running (pid %d) on %s", pid, cfg.MCP.Endpoint())
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating the lazymux binary: %w", err)
	}
	logFile, err := os.OpenFile(LogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", LogPath(), err)
	}
	defer logFile.Close()

	cmd := exec.Command(self, "mcp", "serve")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Setsid detaches the child from this terminal's process group so it
	// survives the shell that started it.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}
	// Reap the child in the background. Without this a child that dies on
	// startup lingers as a zombie, which still answers signal 0 — so a bind
	// failure would look "alive" until the timeout instead of reporting at
	// once. If the server starts fine this goroutine simply outlives us, and
	// the child is reparented to init when we exit.
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()

	if err := waitForReady(cmd.Process.Pid, exited); err != nil {
		return err
	}
	fmt.Printf("lazymux mcp listening on %s (pid %d)\n", cfg.MCP.Endpoint(), cmd.Process.Pid)
	fmt.Printf("logs: %s\n", LogPath())
	return nil
}

// waitForReady blocks until the child publishes its pidfile — which it does
// only after binding the port — or dies, or startTimeout elapses. Waiting on
// the child's own signal rather than on the port being connectable is what
// makes "port already taken by something else" a failure instead of a
// spurious success. On failure it surfaces the tail of the log, the only
// place a detached child's error message lands.
func waitForReady(pid int, exited <-chan error) error {
	deadline := time.After(startTimeout)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		if readPID() == pid {
			return nil
		}
		select {
		case <-exited:
			return fmt.Errorf("server exited during startup:\n%s", tailLog())
		case <-deadline:
			return fmt.Errorf("server did not come up within %s:\n%s", startTimeout, tailLog())
		case <-tick.C:
		}
	}
}

func tailLog() string {
	data, err := os.ReadFile(LogPath())
	if err != nil {
		return "(no log available at " + LogPath() + ")"
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > 10 {
		lines = lines[len(lines)-10:]
	}
	return "  " + strings.Join(lines, "\n  ")
}

// Stop signals the running server and waits for it to exit, escalating to
// SIGKILL if it ignores SIGTERM.
func Stop() error {
	pid := Running()
	if pid == 0 {
		return fmt.Errorf("not running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding pid %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signalling pid %d: %w", pid, err)
	}

	deadline := time.Now().Add(stopTimeout)
	for time.Now().Before(deadline) {
		if !alive(pid) {
			RemovePID()
			fmt.Printf("lazymux mcp stopped (pid %d)\n", pid)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := proc.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("pid %d ignored SIGTERM and could not be killed: %w", pid, err)
	}
	RemovePID()
	fmt.Printf("lazymux mcp killed (pid %d did not exit within %s)\n", pid, stopTimeout)
	return nil
}
