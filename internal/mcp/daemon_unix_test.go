//go:build !windows

package mcp

import (
	"errors"
	"io/fs"
	"os"
	"strconv"
	"testing"
)

func assertPIDFileRemoved(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(PIDPath()); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stale pidfile should be removed, stat err = %v", err)
	}
}

func TestRunningIgnoresUnlockedPIDFileNamingALiveProcess(t *testing.T) {
	newTestWorkspace(t)
	writePIDFile(t, strconv.Itoa(os.Getpid()))

	if pid := Running(); pid != 0 {
		t.Errorf("Running() = %d, want 0: no server holds the pidfile lock", pid)
	}
	assertPIDFileRemoved(t)
}

func TestRunningClearsGarbagePIDFile(t *testing.T) {
	newTestWorkspace(t)
	writePIDFile(t, "not-a-pid")

	if pid := Running(); pid != 0 {
		t.Errorf("Running() = %d, want 0 for an unparseable pidfile", pid)
	}
	assertPIDFileRemoved(t)
}

func TestRunningWithoutPIDFile(t *testing.T) {
	newTestWorkspace(t)
	if pid := Running(); pid != 0 {
		t.Errorf("Running() = %d, want 0 with no pidfile", pid)
	}
}

func TestClaimPIDMarksServerRunningUntilReleased(t *testing.T) {
	newTestWorkspace(t)
	release, err := claimPID()
	if err != nil {
		t.Fatalf("claimPID: %v", err)
	}

	if pid := Running(); pid != os.Getpid() {
		t.Errorf("Running() = %d, want this process %d while the lock is held", pid, os.Getpid())
	}
	if _, err := claimPID(); err == nil {
		t.Error("a second claimPID should fail while the first holds the lock")
	}

	release()
	if pid := Running(); pid != 0 {
		t.Errorf("Running() = %d after release, want 0", pid)
	}
	assertPIDFileRemoved(t)
}

func TestStopWhenNotRunning(t *testing.T) {
	newTestWorkspace(t)
	writePIDFile(t, strconv.Itoa(os.Getpid()))
	if err := Stop(); err == nil {
		t.Fatal("Stop with an unlocked pidfile must not signal the pid it names")
	}
}
