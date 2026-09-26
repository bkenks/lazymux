package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolatePaths points $HOME at a temp dir and clears every variable that
// steers gitkeeper's paths, returning the temp home.
func isolatePaths(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{"GITKEEPER_CONFIG", "GITKEEPER_REPOS", "XDG_CONFIG_HOME", "XDG_DATA_HOME"} {
		t.Setenv(name, "")
	}
	return home
}

func TestRepoRootResolutionOrder(t *testing.T) {
	home := isolatePaths(t)
	envDir := filepath.Join(home, "env-repos")

	tests := []struct {
		name     string
		envRepos string
		reposDir string
		want     string
	}{
		{"unset", "", "", ""},
		{"reposDir", "", "/srv/repos", "/srv/repos"},
		{"GITKEEPER_REPOS beats reposDir", envDir, "/srv/repos", envDir},
		{"GITKEEPER_REPOS expands ~", "~/code", "", filepath.Join(home, "code")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(ReposEnvVar, tt.envRepos)
			if got := (Config{ReposDir: tt.reposDir}).RepoRoot(); got != tt.want {
				t.Errorf("RepoRoot() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateRepoRoot(t *testing.T) {
	home := isolatePaths(t)
	file := filepath.Join(home, "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		reposDir string
		wantErr  string
	}{
		{"unset", "", "no repo directory is set"},
		{"missing", filepath.Join(home, "gone"), "doesn't exist"},
		{"a file", file, "isn't a directory"},
		{"a directory", home, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Config{ReposDir: tt.reposDir}.ValidateRepoRoot()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateRepoRoot() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateRepoRoot() = %v, want an error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseReposDir(t *testing.T) {
	home := isolatePaths(t)
	file := filepath.Join(home, "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"  ~/Development/  ", filepath.Join(home, "Development"), false},
		{"/srv/repos/../code", "/srv/code", false},
		{"", "", true},
		{"   ", "", true},
		{"relative/dir", "", true},
		{"~other/dir", "", true},
		{file, "", true},
	}
	for _, tt := range tests {
		got, err := ParseReposDir(tt.input)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("ParseReposDir(%q) = (%q, %v), want (%q, error %v)",
				tt.input, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestPathFollowsXDGConfigHome(t *testing.T) {
	home := isolatePaths(t)
	if got, want := Path(), filepath.Join(home, ".config", "gitkeeper", "config.json"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}

	t.Setenv("XDG_CONFIG_HOME", "relative")
	if got, want := Path(), filepath.Join(home, ".config", "gitkeeper", "config.json"); got != want {
		t.Errorf("Path() with relative XDG_CONFIG_HOME = %q, want %q", got, want)
	}

	xdg := filepath.Join(home, "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if got, want := Path(), filepath.Join(xdg, "gitkeeper", "config.json"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}

	t.Setenv("GITKEEPER_CONFIG", "/tmp/elsewhere.json")
	if got := Path(); got != "/tmp/elsewhere.json" {
		t.Errorf("Path() = %q, want the $GITKEEPER_CONFIG value", got)
	}
}

func TestFirstRunLeavesTheRepoDirectoryUnset(t *testing.T) {
	isolatePaths(t)

	cfg := Load()
	if err := cfg.ValidateRepoRoot(); err == nil {
		t.Errorf("first run has repo directory %q, want none so the user picks one", cfg.RepoRoot())
	}
	data, err := os.ReadFile(Path())
	if err != nil {
		t.Fatalf("first run didn't write a config: %v", err)
	}
	if strings.Contains(string(data), "reposDir") {
		t.Errorf("config sets reposDir on first run:\n%s", data)
	}
}
