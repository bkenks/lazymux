package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
)

// configWithReposDir is the default config with an existing repo directory,
// so the repo list opens without the repo directory prompt.
func configWithReposDir(t *testing.T) config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.ReposDir = t.TempDir()
	return cfg
}

func isolateReposDir(t *testing.T) {
	t.Helper()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv(config.ReposEnvVar, "")
}

func TestRepoListWaitsForAnExistingReposDir(t *testing.T) {
	isolateReposDir(t)
	missing := config.Default()
	missing.ReposDir = filepath.Join(t.TempDir(), "gone")

	for name, cfg := range map[string]config.Config{
		"unset": config.Default(), "missing": missing, "existing": configWithReposDir(t),
	} {
		m := New(cfg, "test")
		m.Update(events.SetState{State: domain.StateMain})

		want := domain.StateReposDir
		if name == "existing" {
			want = domain.StateMain
		}
		if m.state != want {
			t.Errorf("%s repo directory: state = %v, want %v", name, m.state, want)
		}
	}
}

func TestChosenReposDirIsCreatedSavedAndOpensTheList(t *testing.T) {
	isolateReposDir(t)
	dir := filepath.Join(t.TempDir(), "new", "repos")
	m := New(config.Default(), "test")
	m.Update(events.SetState{State: domain.StateMain})

	_, cmd := m.Update(events.ReposDirChosen{Dir: dir})

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("repo directory not created: %v", err)
	}
	if got := config.Load().ReposDir; got != dir {
		t.Errorf("saved reposDir = %q, want %q", got, dir)
	}
	var reopened bool
	for _, msg := range collectMsgs(cmd) {
		if state, ok := msg.(events.SetState); ok && state.State == domain.StateMain {
			reopened = true
			m.Update(msg)
		}
	}
	if !reopened || m.state != domain.StateMain {
		t.Errorf("state = %v after choosing a repo directory, want the repo list", m.state)
	}
}

func TestReposDirThatCantBeCreatedKeepsThePromptOpen(t *testing.T) {
	isolateReposDir(t)
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(config.Default(), "test")

	_, cmd := m.Update(events.ReposDirChosen{Dir: filepath.Join(file, "repos")})

	if m.state != domain.StateReposDir {
		t.Errorf("state = %v, want the repo directory prompt", m.state)
	}
	var toasted bool
	for _, msg := range collectMsgs(cmd) {
		if toast, ok := msg.(events.Toast); ok && strings.Contains(toast.Msg, "couldn't use") {
			toasted = true
		}
	}
	if !toasted {
		t.Error("no error toast for a repo directory that couldn't be created")
	}
	if got := config.Load().ReposDir; got != "" {
		t.Errorf("saved reposDir = %q, want nothing saved", got)
	}
}

func TestChosenReposDirReplacesAMissingEnvOverride(t *testing.T) {
	isolateReposDir(t)
	t.Setenv(config.ReposEnvVar, filepath.Join(t.TempDir(), "gone"))
	dir := t.TempDir()
	m := New(config.Default(), "test")

	m.Update(events.ReposDirChosen{Dir: dir})
	m.Update(events.SetState{State: domain.StateMain})

	if m.state != domain.StateMain {
		t.Errorf("state = %v, want the repo list once a directory is chosen", m.state)
	}
	if got := m.cfg.RepoRoot(); got != dir {
		t.Errorf("RepoRoot() = %q, want the chosen %q", got, dir)
	}
}
