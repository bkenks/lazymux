package constants

import (
	"charm.land/bubbles/v2/key"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers

type HelpType int

const (
	Short HelpType = iota
	Full
)

func SetOnHelpType(helpType HelpType, bind key.Binding, shortHelp string, fullHelp string) key.Binding {
	bindWithHelp := bind

	switch helpType {
	case Short:
		bindWithHelp.SetHelp(bind.Help().Key, shortHelp)
	case Full:
		bindWithHelp.SetHelp(bind.Help().Key, fullHelp)
	}
	return bindWithHelp
}

// End "Helpers"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Global Key Map

type globalKeyMap struct {
	Quit key.Binding
}

var GlobalKeyMap = globalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// ListQuit replaces the list component's own quit binding, which bubbles v2
// sets to "v". Only q quits; esc is always back.
var ListQuit = key.NewBinding(
	key.WithKeys("q"),
	key.WithHelp("q", "quit"),
)

// End "Global Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Repo List Key Map

type repoListKeyMap struct {
	Clone       key.Binding
	Delete      key.Binding
	VSCode      key.Binding
	Settings    key.Binding
	Refresh     key.Binding
	CopyPath    key.Binding
	Shell       key.Binding
	Keybinds    key.Binding
	Quit        key.Binding
	PullAll     key.Binding
	RepoConfig  key.Binding
	TagRelease  key.Binding
	Registry    key.Binding
	ToggleForge key.Binding
	ToggleStats key.Binding
	CycleSort   key.Binding
}

var RepoListKeyMap = repoListKeyMap{
	Clone: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clone"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	VSCode: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "editor"),
	),
	Settings: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "settings"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	CopyPath: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "copy path"),
	),
	Shell: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "shell"),
	),
	Keybinds: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "keybinds"),
	),
	Quit: GlobalKeyMap.Quit,
	PullAll: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pull all"),
	),
	RepoConfig: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "repo settings"),
	),
	TagRelease: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "tag version"),
	),
	Registry: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "registry"),
	),
	ToggleForge: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "forge label"),
	),
	ToggleStats: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "git stats"),
	),
	CycleSort: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "sort"),
	),
}

// RepoListCommand is one repo-list binding with its help text. InShortHelp puts
// it in the always-visible help bar; every command shows in the full help.
type RepoListCommand struct {
	Binding     key.Binding
	Short, Full string
	InShortHelp bool
}

// Commands lists every repo-list binding in help order. It is the one list the
// help bars, the --help text and the custom-keybind clash check derive from.
func (k repoListKeyMap) Commands() []RepoListCommand {
	return []RepoListCommand{
		{k.Settings, "settings", "open settings", true},
		{k.Keybinds, "keybinds", "manage custom keybinds", true},
		{k.RepoConfig, "repo settings", "edit repo's forge links & tag format", true},
		{k.VSCode, "editor", "open in editor", true},
		{k.Shell, "shell", "shell in repo dir", false},
		{k.CopyPath, "copy", "copy path", false},
		{k.Refresh, "refresh", "refresh list", false},
		{k.Clone, "clone", "clone new repos", true},
		{k.PullAll, "pull all", "git pull every repo (skips conflicts)", false},
		{k.TagRelease, "tag version", "tag & push the next major/minor/patch version", false},
		{k.Registry, "registry", "manage forge registry", false},
		{k.ToggleForge, "forge label", "show/hide the forge label", false},
		{k.ToggleStats, "git stats", "show/hide branch & change counts", false},
		{k.CycleSort, "sort", "cycle sort order", false},
		{k.Delete, "delete", "delete repo", false},
		{k.Quit, "quit", "quit", false},
	}
}

// All returns every repo-list binding, for checking custom keybinds against.
func (k repoListKeyMap) All() []key.Binding {
	commands := k.Commands()
	binds := make([]key.Binding, len(commands))
	for i, c := range commands {
		binds[i] = c.Binding
	}
	return binds
}

// HelpBinds returns the bindings for the short help bar or the full help,
// labelled for that view.
func (k repoListKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	var binds []key.Binding
	for _, c := range k.Commands() {
		if helpType == Short && !c.InShortHelp {
			continue
		}
		binds = append(binds, SetOnHelpType(helpType, c.Binding, c.Short, c.Full))
	}
	return func() []key.Binding { return binds }
}

// End "Repo List Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Confirm Key Map

type confirmKeyMap struct {
	Left     key.Binding
	Right    key.Binding
	Move     key.Binding // display-only: the ←/→ hint shown in help
	Activate key.Binding
	Proceed  key.Binding // ctrl+p: instant confirm shortcut (hidden from help)
	Exit     key.Binding
}

var ConfirmKeyMap = confirmKeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
	),
	Move: key.NewBinding(
		key.WithHelp("←/→", "select"),
	),
	Activate: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Proceed: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "delete now"),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

func (k confirmKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, ConfirmKeyMap.Move, "select", "select yes/no"),
		SetOnHelpType(helpType, ConfirmKeyMap.Activate, "confirm", "confirm selection"),
		SetOnHelpType(helpType, ConfirmKeyMap.Exit, "back", "back to menu"),
		SetOnHelpType(helpType, GlobalKeyMap.Quit, "quit", "quit"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Confirm Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Clone Repo Key Map

type cloneRepoKeyMap struct {
	Exit       key.Binding
	Proceed    key.Binding
	ToggleMode key.Binding
	CycleForge key.Binding
}

var CloneRepoKeyMap = cloneRepoKeyMap{
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Proceed: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "proceed"),
	),
	ToggleMode: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "urls/namespace"),
	),
	CycleForge: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle forge"),
	),
}

func (k cloneRepoKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, CloneRepoKeyMap.Proceed, "proceed", "proceed with cloning"),
		SetOnHelpType(helpType, CloneRepoKeyMap.ToggleMode, "namespace", "toggle urls/namespace mode"),
		SetOnHelpType(helpType, CloneRepoKeyMap.Exit, "back", "back to menu"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Clone Repo Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
