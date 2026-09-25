package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolatePaths points $HOME at a temp dir and clears every variable that
// steers lazymux's paths, returning the temp home.
func isolatePaths(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, name := range []string{"LAZYMUX_CONFIG", "LAZYMUX_ROOT", "XDG_CONFIG_HOME", "XDG_DATA_HOME"} {
		t.Setenv(name, "")
	}
	return home
}

func TestRepoRootResolutionOrder(t *testing.T) {
	home := isolatePaths(t)
	dataHome := filepath.Join(home, "data")
	envRoot := filepath.Join(home, "env-root")

	tests := []struct {
		name     string
		lazyRoot string
		dataHome string
		baseDir  string
		want     string
	}{
		{"home fallback", "", "", "", filepath.Join(home, ".local", "share", "lazymux", "repos")},
		{"relative XDG_DATA_HOME ignored", "", "data", "", filepath.Join(home, ".local", "share", "lazymux", "repos")},
		{"XDG_DATA_HOME", "", dataHome, "", filepath.Join(dataHome, "lazymux", "repos")},
		{"baseDir beats XDG_DATA_HOME", "", dataHome, "/srv/repos", "/srv/repos"},
		{"LAZYMUX_ROOT beats baseDir", envRoot, dataHome, "/srv/repos", envRoot},
		{"LAZYMUX_ROOT expands ~", "~/code", "", "", filepath.Join(home, "code")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LAZYMUX_ROOT", tt.lazyRoot)
			t.Setenv("XDG_DATA_HOME", tt.dataHome)
			if got := (Config{BaseDir: tt.baseDir}).RepoRoot(); got != tt.want {
				t.Errorf("RepoRoot() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathFollowsXDGConfigHome(t *testing.T) {
	home := isolatePaths(t)
	if got, want := Path(), filepath.Join(home, ".config", "lazymux", "config.json"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}

	t.Setenv("XDG_CONFIG_HOME", "relative")
	if got, want := Path(), filepath.Join(home, ".config", "lazymux", "config.json"); got != want {
		t.Errorf("Path() with relative XDG_CONFIG_HOME = %q, want %q", got, want)
	}

	xdg := filepath.Join(home, "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	if got, want := Path(), filepath.Join(xdg, "lazymux", "config.json"); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}

	t.Setenv("LAZYMUX_CONFIG", "/tmp/elsewhere.json")
	if got := Path(); got != "/tmp/elsewhere.json" {
		t.Errorf("Path() = %q, want the $LAZYMUX_CONFIG value", got)
	}
}

func TestFirstRunDoesNotPinTheDefaultRoot(t *testing.T) {
	home := isolatePaths(t)

	Load()
	data, err := os.ReadFile(Path())
	if err != nil {
		t.Fatalf("first run didn't write a config: %v", err)
	}
	if strings.Contains(string(data), "baseDir") {
		t.Errorf("config pins baseDir on first run:\n%s", data)
	}

	dataHome := filepath.Join(home, "data")
	t.Setenv("XDG_DATA_HOME", dataHome)
	if got, want := Load().RepoRoot(), filepath.Join(dataHome, "lazymux", "repos"); got != want {
		t.Errorf("RepoRoot() = %q, want %q", got, want)
	}
}

func writeLegacyJSON(t *testing.T, home, content string) string {
	t.Helper()
	legacy := filepath.Join(home, "lazymux", ".lazymux.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return legacy
}

func TestLoadMovesLegacyJSONAndKeepsRepos(t *testing.T) {
	home := isolatePaths(t)
	legacyRoot := filepath.Join(home, "lazymux")
	legacy := writeLegacyJSON(t, home,
		`{"baseDir": "`+legacyRoot+`", "forges": [{"name": "github", "host": "github.com"}]}`)

	cfg := Load()
	if cfg.LoadFailed {
		t.Fatalf("LoadFailed, warnings = %q", cfg.Warnings)
	}
	if got := cfg.RepoRoot(); got != legacyRoot {
		t.Errorf("RepoRoot() = %q, want the legacy root %q", got, legacyRoot)
	}
	if len(cfg.Forges) != 1 || cfg.Forges[0].Name != "github" {
		t.Errorf("forges = %+v, want the legacy forge", cfg.Forges)
	}
	if _, err := os.Stat(Path()); err != nil {
		t.Errorf("config not written to %s: %v", Path(), err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy config still at %s (err %v)", legacy, err)
	}
	if got := Load(); len(got.Forges) != 1 || got.RepoRoot() != legacyRoot {
		t.Errorf("reload lost the moved config: forges %+v, root %q", got.Forges, got.RepoRoot())
	}
}

func TestLoadMovingLegacyJSONWithoutBaseDirKeepsItsRoot(t *testing.T) {
	home := isolatePaths(t)
	writeLegacyJSON(t, home, `{}`)

	if got, want := Load().RepoRoot(), filepath.Join(home, "lazymux"); got != want {
		t.Errorf("RepoRoot() = %q, want %q", got, want)
	}
}

func TestLoadLeavesUnreadableLegacyJSONInPlace(t *testing.T) {
	home := isolatePaths(t)
	legacy := writeLegacyJSON(t, home, `{not json`)

	cfg := Load()
	if !cfg.LoadFailed {
		t.Error("LoadFailed = false, want true for an unparseable legacy config")
	}
	if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], Path()) {
		t.Errorf("Warnings = %q, want one naming the new config path", cfg.Warnings)
	}
	if _, err := os.Stat(Path()); !os.IsNotExist(err) {
		t.Errorf("a fresh config was written over the unmoved legacy one (err %v)", err)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("legacy config removed: %v", err)
	}
}

func TestLoadIgnoresLegacyJSONWhenConfigPathIsOverridden(t *testing.T) {
	home := isolatePaths(t)
	legacy := writeLegacyJSON(t, home, `{"forges": [{"name": "github", "host": "github.com"}]}`)
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(home, "custom.json"))

	if got := Load(); len(got.Forges) != 0 {
		t.Errorf("forges = %+v, want none from the legacy file", got.Forges)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("legacy config touched: %v", err)
	}
}

func TestUpdateMovesLegacyJSONBeforeWriting(t *testing.T) {
	home := isolatePaths(t)
	writeLegacyJSON(t, home, `{"forges": [{"name": "github", "host": "github.com"}]}`)

	got, err := Update(func(c *Config) { c.Tools.Editor = "vim" })
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Forges) != 1 || got.Tools.Editor != "vim" {
		t.Errorf("Update() = forges %+v, editor %q; want the legacy forge and the change",
			got.Forges, got.Tools.Editor)
	}
	if reloaded := Load(); len(reloaded.Forges) != 1 {
		t.Errorf("legacy forges lost after Update: %+v", reloaded.Forges)
	}
}

func TestUpdateRefusesToBuryUnreadableLegacyJSON(t *testing.T) {
	home := isolatePaths(t)
	writeLegacyJSON(t, home, `{not json`)

	if _, err := Update(func(c *Config) { c.Tools.Editor = "vim" }); err == nil {
		t.Error("Update() succeeded over an unreadable legacy config")
	}
	if _, err := os.Stat(Path()); !os.IsNotExist(err) {
		t.Errorf("Update wrote %s over the unmoved legacy config (err %v)", Path(), err)
	}
}
