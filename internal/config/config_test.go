package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(dir, ".lazymux.json"))

	cfg := Default()
	cfg.Forges = []Forge{{Name: "github", Host: "github.com"}}
	cfg.Repos["bkenks/lazymux"] = RepoLink{Forges: []string{"github"}, Primary: "github", Scheme: "https"}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	got := Load()
	if got.LoadWarning != "" {
		t.Fatalf("warning: %s", got.LoadWarning)
	}
	if len(got.Forges) != 1 || got.Forges[0].Name != "github" {
		t.Errorf("forges = %+v", got.Forges)
	}
	link, ok := got.Repos["bkenks/lazymux"]
	if !ok || link.Primary != "github" || link.Scheme != "https" {
		t.Errorf("repo link = %+v ok=%v", link, ok)
	}
	if f, ok := got.ForgeByHost("GitHub.com"); !ok || f.Name != "github" {
		t.Errorf("ForgeByHost case-insensitive lookup failed: %+v %v", f, ok)
	}
}

func TestMigrateLegacyToml(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(dir, ".lazymux.json"))

	// Point the legacy lookup at a temp XDG dir holding an old config.toml.
	xdg := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdg, "lazymux"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)
	toml := "[tools]\neditor = \"nvim\"\n[ui]\ntheme = \"dracula\"\n[behavior]\ndefault_protocol = \"ssh\"\nconfirm_delete = false\n"
	if err := os.WriteFile(filepath.Join(xdg, "lazymux", "config.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}

	got := Load() // first run: no json yet → migrates the toml
	if got.Tools.Editor != "nvim" {
		t.Errorf("editor = %q, want nvim", got.Tools.Editor)
	}
	if got.UI.Theme != "dracula" {
		t.Errorf("theme = %q, want dracula", got.UI.Theme)
	}
	if got.Behavior.DefaultProtocol != "ssh" {
		t.Errorf("protocol = %q, want ssh", got.Behavior.DefaultProtocol)
	}
	if got.Behavior.ConfirmDelete {
		t.Errorf("confirmDelete should be false")
	}
	// The migrated config should now exist as json.
	if _, err := os.Stat(filepath.Join(dir, ".lazymux.json")); err != nil {
		t.Errorf("migrated json not written: %v", err)
	}
}
