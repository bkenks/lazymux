package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/pkg/settings"
)

// newPersistedApp writes cfg to a temp config file and starts the app on it.
func newPersistedApp(t *testing.T, cfg config.Config) *ModelManager {
	t.Helper()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(t.TempDir(), ".lazymux.json"))
	cfg.BaseDir = t.TempDir()
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	return New(config.Load(), "test")
}

func describedRepoConfig() config.Config {
	cfg := config.Default()
	cfg.Forges = []config.Forge{{Name: "github", Host: "github.com"}}
	cfg.Repos["me/demo"] = config.RepoLink{
		Upstreams: []string{"github"}, Origin: "github", Purpose: "demo purpose", Context: "demo context",
	}
	return cfg
}

func TestForgeRegistryEditsKeepRepoDescriptions(t *testing.T) {
	m := newPersistedApp(t, describedRepoConfig())

	m.Update(events.ForgesChanged{
		Forges: []config.Forge{{Name: "gh", Host: "github.com"}},
		Repos:  map[string]config.RepoLink{"me/demo": {Upstreams: []string{"gh"}, Origin: "gh"}},
	})

	link := config.Load().Repos["me/demo"]
	if link.Purpose != "demo purpose" || link.Context != "demo context" {
		t.Errorf("description lost: %+v", link)
	}
	if link.Origin != "gh" {
		t.Errorf("forge rename not applied: %+v", link)
	}
}

func TestAppSaveKeepsDescriptionWrittenByAnotherProcess(t *testing.T) {
	cfg := describedRepoConfig()
	cfg.Repos["me/demo"] = config.RepoLink{Upstreams: []string{"github"}, Origin: "github"}
	m := newPersistedApp(t, cfg)

	if _, err := config.Update(func(c *config.Config) {
		link := c.Repos["me/demo"]
		link.Purpose = "set over MCP"
		c.Repos["me/demo"] = link
	}); err != nil {
		t.Fatal(err)
	}
	m.Update(events.KeybindsChanged{Keybinds: []config.Keybind{{Name: "log", Keys: "ctrl+g", Command: "git log"}}})

	if got := config.Load().Repos["me/demo"].Purpose; got != "set over MCP" {
		t.Errorf("Purpose = %q, want the MCP write to survive", got)
	}
}

func TestUnlinkingEveryForgeKeepsRepoDescription(t *testing.T) {
	m := newPersistedApp(t, describedRepoConfig())

	m.Update(events.RepoLinkChanged{Key: "me/demo", Link: config.RepoLink{}})

	link, ok := config.Load().Repos["me/demo"]
	if !ok || link.Purpose != "demo purpose" {
		t.Errorf("link = %+v ok=%v, want the description kept", link, ok)
	}
}

func TestRepoDeletedRemovesTheDeletedRepoLink(t *testing.T) {
	cfg := describedRepoConfig()
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

	m.Update(settings.SettingChanged{Key: "confirm_delete", Setting: settings.NewToggle("confirm_delete", "", false)})

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
