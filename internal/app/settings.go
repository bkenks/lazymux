package app

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/pkg/settings"
)

// settingField ties one row of the settings screen to the config field it
// edits: build renders the row from a config, apply writes an edited row back
// into one, and show (if set) applies the new value to the live repo list.
type settingField struct {
	key   string
	build func(cfg config.Config) settings.Setting
	apply func(cfg *config.Config, s settings.Setting)
	show  func(m *ModelManager)
}

// settingFields are the settings screen's rows, in display order.
var settingFields = []settingField{
	textField("editor", "Editor", validateEditorCommand,
		func(c *config.Config) *string { return &c.Tools.Editor }),
	selectField("default_protocol", "Default clone protocol", protocolOptions,
		func(c *config.Config) *string { return &c.Behavior.DefaultProtocol }),
	toggleField("confirm_delete", "Confirm before deleting",
		func(c *config.Config) *bool { return &c.Behavior.ConfirmDelete }),
	withShow(toggleField("show_full_path", "Show full path on rows",
		func(c *config.Config) *bool { return &c.UI.ShowFullPath }),
		func(m *ModelManager) { domain.ShowFullPath = m.cfg.UI.ShowFullPath }),
	withShow(toggleField("show_forge", "Show forge label on rows",
		func(c *config.Config) *bool { return &c.UI.ShowForge }),
		func(m *ModelManager) {
			domain.ShowForge = m.cfg.UI.ShowForge
			m.main.SyncForgeVisibility()
		}),
	withShow(toggleField("show_stats", "Show git stats on rows",
		func(c *config.Config) *bool { return &c.UI.ShowStats }),
		func(m *ModelManager) { domain.ShowStats = m.cfg.UI.ShowStats }),
	withShow(selectField("sort_mode", "Sort repos by", sortOptions,
		func(c *config.Config) *string { return &c.UI.SortMode }),
		func(m *ModelManager) {
			domain.Sort = domain.ParseSortMode(m.cfg.UI.SortMode)
			m.main.Resort()
		}),
}

func toggleField(key, label string, field func(*config.Config) *bool) settingField {
	return settingField{
		key: key,
		build: func(cfg config.Config) settings.Setting {
			return settings.NewToggle(key, label, *field(&cfg))
		},
		apply: func(cfg *config.Config, s settings.Setting) { *field(cfg) = s.Value() == true },
	}
}

func selectField(
	key, label string, options []string, field func(*config.Config) *string,
) settingField {
	return settingField{
		key: key,
		build: func(cfg config.Config) settings.Setting {
			return settings.NewSelect(key, label, options, max(slices.Index(options, *field(&cfg)), 0))
		},
		apply: func(cfg *config.Config, s settings.Setting) { *field(cfg) = s.ValueString() },
	}
}

func textField(
	key, label string, validate settings.Validator, field func(*config.Config) *string,
) settingField {
	return settingField{
		key: key,
		build: func(cfg config.Config) settings.Setting {
			return settings.NewText(key, label, *field(&cfg), validate)
		},
		apply: func(cfg *config.Config, s settings.Setting) { *field(cfg) = s.ValueString() },
	}
}

func withShow(f settingField, show func(m *ModelManager)) settingField {
	f.show = show
	return f
}

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

var protocolOptions = []string{config.SchemeHTTPS, config.SchemeSSH}

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
	items := make([]settings.Setting, 0, len(settingFields))
	for _, field := range settingFields {
		items = append(items, field.build(cfg))
	}
	return items
}

// applySettingChange saves an edited setting and applies it to the live repo
// list, returning the toast that reports the outcome.
func (m *ModelManager) applySettingChange(msg settings.SettingChanged) tea.Cmd {
	i := slices.IndexFunc(settingFields, func(f settingField) bool { return f.key == msg.Key })
	if i < 0 {
		return m.toastCmd(events.ToastError, fmt.Sprintf("unknown setting %q", msg.Key))
	}
	field := settingFields[i]
	saveErr := m.saveConfig("config", func(c *config.Config) { field.apply(c, msg.Setting) })
	if field.show != nil {
		field.show(m)
	}
	if saveErr != nil {
		return saveErr
	}
	return m.toastCmd(events.ToastInfo, "settings saved")
}
