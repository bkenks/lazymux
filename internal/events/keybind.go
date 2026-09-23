package events

import "github.com/bkenks/lazymux/internal/config"

// KeybindsChanged replaces the whole custom keybind list (from the keybinds
// screen) and is persisted right away.
type KeybindsChanged struct{ Keybinds []config.Keybind }

func (KeybindsChanged) isEvent() {}

// RunKeybind runs a custom keybind's command in Dir, the selected repo's
// directory, with the whole terminal handed to it.
type RunKeybind struct {
	Keybind config.Keybind
	Dir     string
}

func (RunKeybind) isEvent() {}
