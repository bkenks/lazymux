package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".lazymux.json")
	t.Setenv("LAZYMUX_CONFIG", path)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMarksUnparseableConfigAsFailed(t *testing.T) {
	writeConfigFile(t, `{"forges": [`)

	cfg := Load()
	if !cfg.LoadFailed {
		t.Fatal("LoadFailed = false for a config that doesn't parse")
	}
	if len(cfg.Warnings) == 0 {
		t.Error("expected a warning explaining the failure")
	}
	if err := Save(cfg); err == nil {
		t.Error("Save wrote defaults over a config that failed to load")
	}
}

func TestUpdateRefusesToOverwriteUnparseableConfig(t *testing.T) {
	const broken = `{"forges": [`
	path := writeConfigFile(t, broken)

	_, err := Update(func(c *Config) { c.Tools.Editor = "vim" })
	if err == nil {
		t.Fatal("Update succeeded over a config that doesn't parse")
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != broken {
		t.Errorf("file was rewritten to %q", data)
	}
}

func TestUpdateKeepsEditsMadeByAnotherWriter(t *testing.T) {
	writeConfigFile(t, `{}`)
	stale := Load()

	if _, err := Update(func(c *Config) {
		c.Repos["ns/repo"] = RepoLink{Purpose: "written by the MCP server"}
	}); err != nil {
		t.Fatal(err)
	}

	got, err := Update(func(c *Config) { c.Tools.Editor = "vim" })
	if err != nil {
		t.Fatal(err)
	}
	if got.Repos["ns/repo"].Purpose != "written by the MCP server" {
		t.Errorf("purpose lost: %+v", got.Repos["ns/repo"])
	}
	if stale.Tools.Editor == "vim" {
		t.Error("Update mutated a previously loaded config")
	}
}

func TestUpdateCreatesMissingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", ".lazymux.json")
	t.Setenv("LAZYMUX_CONFIG", path)

	if _, err := Update(func(c *Config) { c.Tools.Editor = "vim" }); err != nil {
		t.Fatal(err)
	}
	if Load().Tools.Editor != "vim" {
		t.Error("change not persisted")
	}
}

func TestLoadNormalizesReposDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfigFile(t, `{"baseDir": "~/repos/../lazymux/"}`)

	if got, want := Load().ReposDir, filepath.Join(home, "lazymux"); got != want {
		t.Errorf("ReposDir = %q, want %q", got, want)
	}
}

func TestLoadRejectsUnknownDefaultProtocol(t *testing.T) {
	writeConfigFile(t, `{"behavior": {"defaultProtocol": "SSH"}}`)
	if got := Load().Behavior.DefaultProtocol; got != SchemeSSH {
		t.Errorf("DefaultProtocol = %q, want ssh", got)
	}

	writeConfigFile(t, `{"behavior": {"defaultProtocol": "ftp"}}`)
	cfg := Load()
	if cfg.Behavior.DefaultProtocol != SchemeHTTPS {
		t.Errorf("DefaultProtocol = %q, want https", cfg.Behavior.DefaultProtocol)
	}
	if len(cfg.Warnings) != 1 {
		t.Errorf("Warnings = %q, want one for the unknown protocol", cfg.Warnings)
	}
}

func TestConfigCloneSharesNoState(t *testing.T) {
	original := Default()
	original.Forges = []Forge{{Name: "github", Host: "github.com"}}
	original.Repos["ns/repo"] = RepoLink{Upstreams: []string{"github"}, Origin: "github"}

	clone := original.Clone()
	clone.Forges[0].Name = "changed"
	link := clone.Repos["ns/repo"]
	link.Upstreams[0] = "changed"
	clone.Repos["ns/other"] = RepoLink{}

	if original.Forges[0].Name != "github" {
		t.Error("clone shares Forges")
	}
	if original.Repos["ns/repo"].Upstreams[0] != "github" {
		t.Error("clone shares a link's Upstreams")
	}
	if _, ok := original.Repos["ns/other"]; ok {
		t.Error("clone shares the Repos map")
	}
}

func TestRepoLinkEditsNeverAliasTheOriginal(t *testing.T) {
	upstreams := make([]string, 2, 4)
	copy(upstreams, []string{"a", "b"})
	original := RepoLink{Upstreams: upstreams, Origin: "a"}

	edited := original
	edited.RemoveUpstream("a")
	edited.ToggleUpstream("c")
	edited.SetOrigin("d")

	if !slices.Equal(original.Upstreams, []string{"a", "b"}) || original.Origin != "a" {
		t.Errorf("original changed: %+v", original)
	}
	if !slices.Equal(edited.Upstreams, []string{"b", "c", "d"}) || edited.Origin != "d" {
		t.Errorf("edited = %+v", edited)
	}
}

func TestRepoLinkRemovingOriginPromotesNextUpstream(t *testing.T) {
	link := RepoLink{Upstreams: []string{"a", "b"}, Origin: "a"}
	link.ToggleUpstream("a")
	if link.Origin != "b" {
		t.Errorf("Origin = %q, want b", link.Origin)
	}
	link.ToggleUpstream("b")
	if link.Origin != "" || len(link.Upstreams) != 0 {
		t.Errorf("link = %+v, want unlinked", link)
	}
	link.ToggleUpstream("c")
	if link.Origin != "c" {
		t.Errorf("first upstream added should become origin, got %q", link.Origin)
	}
}

func TestRepoLinkRenameUpstreamDropsDuplicate(t *testing.T) {
	link := RepoLink{Upstreams: []string{"old", "new"}, Origin: "old"}
	link.RenameUpstream("old", "new")
	if !slices.Equal(link.Upstreams, []string{"new"}) || link.Origin != "new" {
		t.Errorf("link = %+v", link)
	}
}

func TestRepoLinkWithForgeLinksKeepsDescription(t *testing.T) {
	stored := RepoLink{Upstreams: []string{"a"}, Origin: "a", Purpose: "p", Context: "c"}
	edited := RepoLink{Upstreams: []string{"b"}, Origin: "b", Scheme: SchemeSSH}

	got := stored.WithForgeLinks(edited)
	if got.Purpose != "p" || got.Context != "c" || got.Origin != "b" || got.Scheme != SchemeSSH {
		t.Errorf("got %+v", got)
	}
}

func TestLoadReplacesOutOfRangeMCPPort(t *testing.T) {
	writeConfigFile(t, `{"mcp": {"port": 99999}}`)
	cfg := Load()
	if cfg.MCP.Port != DefaultMCPPort {
		t.Errorf("Port = %d, want the default", cfg.MCP.Port)
	}
	if len(cfg.Warnings) != 1 {
		t.Errorf("Warnings = %q, want one for the bad port", cfg.Warnings)
	}
}
