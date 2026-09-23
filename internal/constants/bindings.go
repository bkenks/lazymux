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

type keyMap interface {
	HelpBinds()
}

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
//// Default Key Map

type defaultKeyMap struct {
	Select key.Binding
	Exit   key.Binding
}

var DefaultKeyMap = defaultKeyMap{
	Select: key.NewBinding(
		key.WithKeys(
			"enter",
			"space",
		),
		key.WithHelp(
			"enter/space",
			"select",
		),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit"),
	),
}

func (k defaultKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, DefaultKeyMap.Select, "select", "select"),
		SetOnHelpType(helpType, DefaultKeyMap.Exit, "exit", "exit"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Default Key Map"
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
	Forges      key.Binding
	Registry    key.Binding
	ToggleForge key.Binding
	ToggleStats key.Binding
	CycleSort   key.Binding
}

var RepoListKeyMap = repoListKeyMap{
	Clone: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "clone"),
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
		key.WithKeys(","),
		key.WithHelp(",", "settings"),
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
		key.WithKeys("ctrl+shift+k"),
		key.WithHelp("ctrl+shift+k", "keybinds"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	PullAll: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pull all"),
	),
	Forges: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "forges"),
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

// All returns every repo-list binding, for checking custom keybinds against.
func (k repoListKeyMap) All() []key.Binding {
	return []key.Binding{
		k.Clone, k.Delete, k.VSCode, k.Settings, k.Refresh, k.CopyPath, k.Shell, k.Keybinds,
		k.Quit, k.PullAll, k.Forges, k.Registry, k.ToggleForge, k.ToggleStats, k.CycleSort,
	}
}

func (k repoListKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	// Short help is the always-visible bar — keep it to the essentials.
	// Everything shows in the full help (press ?).
	if helpType == Short {
		binds := []key.Binding{
			SetOnHelpType(Short, RepoListKeyMap.VSCode, "editor", ""),
			SetOnHelpType(Short, RepoListKeyMap.Keybinds, "keybinds", ""),
			SetOnHelpType(Short, RepoListKeyMap.Clone, "clone", ""),
			SetOnHelpType(Short, RepoListKeyMap.Forges, "forges", ""),
			SetOnHelpType(Short, RepoListKeyMap.Settings, "settings", ""),
		}
		return func() []key.Binding { return binds }
	}

	binds := []key.Binding{
		SetOnHelpType(Full, RepoListKeyMap.VSCode, "editor", "open in editor"),
		SetOnHelpType(Full, RepoListKeyMap.Shell, "shell", "shell in repo dir"),
		SetOnHelpType(Full, RepoListKeyMap.Keybinds, "keybinds", "manage custom keybinds"),
		SetOnHelpType(Full, RepoListKeyMap.CopyPath, "copy", "copy path"),
		SetOnHelpType(Full, RepoListKeyMap.Refresh, "refresh", "refresh list"),
		SetOnHelpType(Full, RepoListKeyMap.Clone, "clone", "clone new repos"),
		SetOnHelpType(Full, RepoListKeyMap.PullAll, "pull all", "git pull every repo (skips conflicts)"),
		SetOnHelpType(Full, RepoListKeyMap.Forges, "forges", "edit repo's forge links"),
		SetOnHelpType(Full, RepoListKeyMap.Registry, "registry", "manage forge registry"),
		SetOnHelpType(Full, RepoListKeyMap.ToggleForge, "forge label", "show/hide the forge label"),
		SetOnHelpType(Full, RepoListKeyMap.ToggleStats, "git stats", "show/hide branch & change counts"),
		SetOnHelpType(Full, RepoListKeyMap.CycleSort, "sort", "cycle sort order"),
		SetOnHelpType(Full, RepoListKeyMap.Delete, "delete", "delete repo"),
		SetOnHelpType(Full, RepoListKeyMap.Settings, "settings", "open settings"),
		SetOnHelpType(Full, RepoListKeyMap.Quit, "quit", "quit"),
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
