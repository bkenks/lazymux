package events

import "github.com/bkenks/lazymux/internal/config"

// SettingsChanged carries the config as the submitted settings form left it.
// The fields that screen edits are saved from it.
type SettingsChanged struct{ Config config.Config }

func (SettingsChanged) isEvent() {}
