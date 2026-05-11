package app

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/pkg/settings"
)

// settings keys persisted in config.toml
const (
	skEditor        = "editor"
	skProtocol      = "default_protocol"
	skConfirmDelete = "confirm_delete"
	skShowFullPath  = "show_full_path"
)

var editorOptions = []string{"codium", "code", "nvim", "vim", "hx", "zed", "idea"}
var protocolOptions = []string{"https", "ssh"}

func indexOrZero(opts []string, want string) int {
	for i, v := range opts {
		if v == want {
			return i
		}
	}
	return 0
}

func buildSettingsItems(cfg config.Config) []settings.Setting {
	return []settings.Setting{
		settings.NewSelect(skEditor, "Editor", editorOptions, indexOrZero(editorOptions, cfg.Tools.Editor)),
		settings.NewSelect(skProtocol, "Default clone protocol", protocolOptions, indexOrZero(protocolOptions, cfg.Behavior.DefaultProtocol)),
		settings.NewToggle(skConfirmDelete, "Confirm before deleting", cfg.Behavior.ConfirmDelete),
		settings.NewToggle(skShowFullPath, "Show full path on rows", cfg.UI.ShowFullPath),
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
	}
	commands.SetDeps(m.cfg)
}
