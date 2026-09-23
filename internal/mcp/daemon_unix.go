//go:build !windows

package mcp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bkenks/lazymux/internal/config"
)

const (
	// startTimeout bounds how long `mcp start` waits for the child to bind.
	startTimeout = 5 * time.Second
	// stopTimeout bounds how long `mcp stop` waits for a graceful exit before
	// escalating to SIGKILL.
	stopTimeout = 5 * time.Second
)

// readPID returns the pid recorded in the pidfile, or 0 if there isn't a
// usable one.
func readPID() int {
	data, err := os.ReadFile(PIDPath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

// tryLock takes an exclusive flock on f without blocking, reporting false if
// another open file already holds it.
func tryLock(f *os.File) (bool, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Running returns the pid of the live server, or 0 if none is running. A
// server is live only while it holds the pidfile's lock, so a pid reused by an
// unrelated process is never mistaken for it. A pidfile nobody holds is stale
// and is removed, so a crashed server never blocks the next start.
func Running() int {
	f, err := os.Open(PIDPath())
	if err != nil {
		return 0
	}
	defer f.Close()
	locked, err := tryLock(f)
	if err != nil {
		return 0
	}
	if locked {
		os.Remove(PIDPath())
		return 0
	}
	return readPID()
}

// claimPID records the current process as the running server and holds the
// pidfile's lock until the returned release is called or the process exits.
// Release removes the pidfile if it still names this process.
func claimPID() (func(), error) {
	f, err := os.OpenFile(PIDPath(), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", PIDPath(), err)
	}
	locked, err := tryLock(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("locking %s: %w", PIDPath(), err)
	}
	if !locked {
		f.Close()
		return nil, fmt.Errorf("another server holds %s", PIDPath())
	}
	if err := writePID(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("writing %s: %w", PIDPath(), err)
	}
	return func() {
		if readPID() == os.Getpid() {
			os.Remove(PIDPath())
		}
		f.Close()
	}, nil
}

func writePID(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	_, err := f.WriteAt([]byte(strconv.Itoa(os.Getpid())), 0)
	return err
}

// detach puts the server child in its own session, so it survives the shell
// that started it.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

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
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}
	// Reap the child in the background. Without this a child that dies on
	// startup lingers as a zombie until the timeout instead of reporting at
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

// Stop signals the running server and waits for it to release the pidfile,
// escalating to SIGKILL if it ignores SIGTERM.
func Stop() error {
	pid := Running()
	if pid == 0 {
		return errors.New("not running")
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
		if Running() == 0 {
			fmt.Printf("lazymux mcp stopped (pid %d)\n", pid)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := proc.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("pid %d ignored SIGTERM and could not be killed: %w", pid, err)
	}
	fmt.Printf("lazymux mcp killed (pid %d did not exit within %s)\n", pid, stopTimeout)
	return nil
}
