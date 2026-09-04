package app

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/pkg/settings"
)

// settings keys persisted in config.toml
const (
	skEditor        = "editor"
	skProtocol      = "default_protocol"
	skConfirmDelete = "confirm_delete"
	skShowFullPath  = "show_full_path"
	skShowForge     = "show_forge"
	skShowStats     = "show_stats"
	skSortMode      = "sort_mode"
)

// sortOptions are the repo list orderings offered in the settings screen, in
// the same order the list's sort key cycles through them.
var sortOptions = sortModeStrings()

func sortModeStrings() []string {
	opts := make([]string, 0, len(domain.SortModes))
	for _, m := range domain.SortModes {
		opts = append(opts, string(m))
	}
	return opts
}

var protocolOptions = []string{"https", "ssh"}

// equalStrings reports whether two forge-name lists hold the same names in the
// same order.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// forgeHost returns the host of the named forge in the given slice, or "" if
// it isn't present.
func forgeHost(forges []config.Forge, name string) string {
	for _, f := range forges {
		if f.Name == name {
			return f.Host
		}
	}
	return ""
}

func indexOrZero(opts []string, want string) int {
	for i, v := range opts {
		if v == want {
			return i
		}
	}
	return 0
}

// validateEditorCommand resolves an editor command the way exec.Command will
// when a repo is opened, so a value the settings screen accepts is a value that
// actually runs. The resolved path comes back as the confirmation hint.
func validateEditorCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("editor cannot be empty")
	}
	if strings.ContainsAny(command, " \t") {
		return "", errors.New("editor takes a command name only, no arguments")
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("%q not found on PATH", command)
	}
	return path, nil
}

func buildSettingsItems(cfg config.Config) []settings.Setting {
	return []settings.Setting{
		settings.NewText(skEditor, "Editor", cfg.Tools.Editor, validateEditorCommand),
		settings.NewSelect(skProtocol, "Default clone protocol", protocolOptions, indexOrZero(protocolOptions, cfg.Behavior.DefaultProtocol)),
		settings.NewToggle(skConfirmDelete, "Confirm before deleting", cfg.Behavior.ConfirmDelete),
		settings.NewToggle(skShowFullPath, "Show full path on rows", cfg.UI.ShowFullPath),
		settings.NewToggle(skShowForge, "Show forge label on rows", cfg.UI.ShowForge),
		settings.NewToggle(skShowStats, "Show git stats on rows", cfg.UI.ShowStats),
		settings.NewSelect(skSortMode, "Sort repos by", sortOptions, indexOrZero(sortOptions, cfg.UI.SortMode)),
	}
}

// applySettingChange mutates the in-memory cfg and propagates to commands.SetDeps
// so subsequent commands pick up the new value immediately.
func (m *ModelManager) applySettingChange(msg settings.SettingChanged) {
	switch msg.Key {
	case skEditor:
		m.cfg.Tools.Editor = msg.Setting.ValueString()
	case skProtocol:
		m.cfg.Behavior.DefaultProtocol = msg.Setting.ValueString()
	case skConfirmDelete:
		if v, ok := msg.Setting.Value().(bool); ok {
			m.cfg.Behavior.ConfirmDelete = v
		}
	case skShowFullPath:
		if v, ok := msg.Setting.Value().(bool); ok {
			m.cfg.UI.ShowFullPath = v
		}
	case skShowForge:
		if v, ok := msg.Setting.Value().(bool); ok {
			m.cfg.UI.ShowForge = v
			domain.ShowForge = v // apply to the live repo list immediately
			m.main.SyncForgeVisibility()
		}
	case skShowStats:
		if v, ok := msg.Setting.Value().(bool); ok {
			m.cfg.UI.ShowStats = v
			domain.ShowStats = v // apply to the live repo list immediately
		}
	case skSortMode:
		m.cfg.UI.SortMode = msg.Setting.ValueString()
		domain.Sort = domain.ParseSortMode(m.cfg.UI.SortMode)
		m.main.Resort() // reorder the live list; its cmd only matters while filtering
	}
	commands.SetDeps(m.cfg)
}
