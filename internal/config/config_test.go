package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCanonicalizesHandEditedKeybinds(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gitkeeper.json")
	t.Setenv("GITKEEPER_CONFIG", path)

	edited := `{"keybinds":[{"name":"log","keys":"Ctrl + G","command":"git log"},` +
		`{"name":"broken","keys":"ctrl + banana","command":"true"}]}`
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if got := cfg.Keybinds[0].Keys; got != "ctrl+g" {
		t.Errorf("Keys = %q, want ctrl+g", got)
	}
	if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], "broken") {
		t.Errorf("Warnings = %q, want one naming the unparseable keybind", cfg.Warnings)
	}
	if cfg.LoadFailed {
		t.Error("a bad keybind must not mark the whole config as failed to load")
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GITKEEPER_CONFIG", filepath.Join(dir, ".gitkeeper.json"))

	cfg := Default()
	cfg.Forges = []Forge{{Name: "github", Host: "github.com"}}
	cfg.Repos["bkenks/gitkeeper"] = RepoLink{Upstreams: []string{"github"}, Origin: "github", Scheme: "https"}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	got := Load()
	if got.LoadFailed || len(got.Warnings) > 0 {
		t.Fatalf("LoadFailed=%v warnings=%q", got.LoadFailed, got.Warnings)
	}
	if len(got.Forges) != 1 || got.Forges[0].Name != "github" {
		t.Errorf("forges = %+v", got.Forges)
	}
	link, ok := got.Repos["bkenks/gitkeeper"]
	if !ok || link.Origin != "github" || link.Scheme != "https" {
		t.Errorf("repo link = %+v ok=%v", link, ok)
	}
	if f, ok := got.ForgeByHost("GitHub.com"); !ok || f.Name != "github" {
		t.Errorf("ForgeByHost case-insensitive lookup failed: %+v %v", f, ok)
	}
}

func TestLoadMigratesForgesPrimary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitkeeper.json")
	t.Setenv("GITKEEPER_CONFIG", path)

	legacy := `{"repos":{"bkenks/gitkeeper":{"forges":["github","forgejo"],` +
		`"primary":"forgejo","scheme":"https"}}}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	link := Load().Repos["bkenks/gitkeeper"]
	if link.Origin != "forgejo" {
		t.Errorf("Origin = %q, want forgejo", link.Origin)
	}
	if len(link.Upstreams) != 2 || link.Upstreams[0] != "github" {
		t.Errorf("Upstreams = %+v", link.Upstreams)
	}
	if link.LegacyForges != nil || link.LegacyPrimary != "" {
		t.Errorf("legacy fields not cleared: %+v", link)
	}
}
