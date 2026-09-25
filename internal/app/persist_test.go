package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/events"
)

// newPersistedApp writes cfg to a temp config file and starts the app on it.
func newPersistedApp(t *testing.T, cfg config.Config) *ModelManager {
	t.Helper()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	cfg.ReposDir = t.TempDir()
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	return New(config.Load(), "test")
}

func taggedRepoConfig() config.Config {
	cfg := config.Default()
	cfg.Forges = []config.Forge{{Name: "github", Host: "github.com"}}
	cfg.Repos["me/demo"] = config.RepoLink{
		Upstreams: []string{"github"}, Origin: "github", TagPrefix: "v",
	}
	return cfg
}

func TestForgeRegistryEditsKeepRepoTagFormat(t *testing.T) {
	m := newPersistedApp(t, taggedRepoConfig())

	m.Update(events.ForgesChanged{
		Forges: []config.Forge{{Name: "gh", Host: "github.com"}},
		Repos:  map[string]config.RepoLink{"me/demo": {Upstreams: []string{"gh"}, Origin: "gh"}},
	})

	link := config.Load().Repos["me/demo"]
	if link.TagPrefix != "v" {
		t.Errorf("tag format lost: %+v", link)
	}
	if link.Origin != "gh" {
		t.Errorf("forge rename not applied: %+v", link)
	}
}

func TestAppSaveKeepsEditWrittenByAnotherProcess(t *testing.T) {
	cfg := taggedRepoConfig()
	cfg.Repos["me/demo"] = config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	m := newPersistedApp(t, cfg)

	if _, err := config.Update(func(c *config.Config) {
		link := c.Repos["me/demo"]
		link.TagPrefix = "set elsewhere/v"
		c.Repos["me/demo"] = link
	}); err != nil {
		t.Fatal(err)
	}
	keybinds := []config.Keybind{{Name: "log", Keys: "ctrl+g", Command: "git log"}}
	m.Update(events.KeybindsChanged{Keybinds: keybinds})

	if got := config.Load().Repos["me/demo"].TagPrefix; got != "set elsewhere/v" {
		t.Errorf("TagPrefix = %q, want the other process's write to survive", got)
	}
}

func TestUnlinkingEveryForgeKeepsRepoTagFormat(t *testing.T) {
	m := newPersistedApp(t, taggedRepoConfig())

	m.Update(events.RepoSettingsChanged{Key: "me/demo", Link: config.RepoLink{TagPrefix: "v"}})

	link, ok := config.Load().Repos["me/demo"]
	if !ok || link.TagPrefix != "v" || link.Origin != "" {
		t.Errorf("link = %+v ok=%v, want only the tag format kept", link, ok)
	}
}

func TestEmptyRepoSettingsRemoveTheRepoLink(t *testing.T) {
	m := newPersistedApp(t, taggedRepoConfig())

	m.Update(events.RepoSettingsChanged{Key: "me/demo", Link: config.RepoLink{}})

	if link, ok := config.Load().Repos["me/demo"]; ok {
		t.Errorf("link = %+v, want the empty link removed", link)
	}
}

func TestRepoSettingsSaveTagFormat(t *testing.T) {
	m := newPersistedApp(t, taggedRepoConfig())

	m.Update(events.RepoSettingsChanged{Key: "me/demo", Link: config.RepoLink{
		Upstreams: []string{"github"}, Origin: "github", TagPrefix: "mypkg/v", TagSuffix: "-x",
	}})

	link := config.Load().Repos["me/demo"]
	if link.TagPrefix != "mypkg/v" || link.TagSuffix != "-x" {
		t.Errorf("tag format = %q/%q, want it saved", link.TagPrefix, link.TagSuffix)
	}
}

func TestRepoDeletedRemovesTheDeletedRepoLink(t *testing.T) {
	cfg := taggedRepoConfig()
	cfg.Repos["me/other"] = config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	m := newPersistedApp(t, cfg)

	m.Update(events.RepoDeleted{Key: "me/other"})

	repos := config.Load().Repos
	if _, ok := repos["me/other"]; ok {
		t.Error("deleted repo's link still saved")
	}
	if _, ok := repos["me/demo"]; !ok {
		t.Error("an unrelated repo link was removed")
	}
}

func TestSettingsSaveDoesNotOverwriteUnparseableConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".lazymux.json")
	t.Setenv("LAZYMUX_CONFIG", path)
	const broken = `{"forges": [`
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(config.Load(), "test")

	edited := m.cfg.Clone()
	edited.Behavior.ConfirmDelete = false
	m.Update(events.SettingsChanged{Config: edited})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != broken {
		t.Errorf("config overwritten with %q", data)
	}
	if m.cfg.Behavior.ConfirmDelete {
		t.Error("setting should still apply for the session")
	}
}

func TestInitWarnsAboutClashingKeybinds(t *testing.T) {
	cfg := taggedRepoConfig()
	cfg.Keybinds = []config.Keybind{
		{Name: "shadow", Keys: "o", Command: "true"},
		{Name: "first", Keys: "ctrl+g", Command: "true"},
		{Name: "second", Keys: "ctrl+g", Command: "true"},
	}
	m := newPersistedApp(t, cfg)

	clashes := m.keybindClashes()
	if len(clashes) != 2 ||
		!strings.Contains(clashes[0], "shadow") || !strings.Contains(clashes[1], "second") {
		t.Errorf("clashes = %q", clashes)
	}
}
