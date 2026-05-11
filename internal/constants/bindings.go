package constants

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
//// Default Key Map

type defaultKeyMap struct {
	Select key.Binding
	Exit   key.Binding
}

var DefaultKeyMap = defaultKeyMap{
	Select: key.NewBinding(
		key.WithKeys(
			tea.KeyEnter.String(),
			tea.KeySpace.String(),
		),
		key.WithHelp(
			tea.KeyEnter.String()+"/"+tea.KeySpace.String(),
			"select",
		),
	),
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), "exit"),
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
	Select   key.Binding
	Clone    key.Binding
	Delete   key.Binding
	VSCode   key.Binding
	Settings key.Binding
	Refresh  key.Binding
	CopyPath key.Binding
	Shell    key.Binding
	Quit     key.Binding
	PullAll  key.Binding
}

var RepoListKeyMap = repoListKeyMap{
	Select: key.NewBinding(
		key.WithKeys(tea.KeyTab.String()),
		key.WithHelp(tea.KeyTab.String(), "lazygit"),
	),
	Clone: key.NewBinding(
		key.WithKeys(tea.KeyCtrlN.String()),
		key.WithHelp("ctrl+n", "clone"),
	),
	Delete: key.NewBinding(
		key.WithKeys(tea.KeyCtrlBackslash.String()),
		key.WithHelp("ctrl+\\", "delete"),
	),
	VSCode: key.NewBinding(
		key.WithKeys(tea.KeyCtrlO.String()),
		key.WithHelp("ctrl+o", "editor"),
	),
	Settings: key.NewBinding(
		key.WithKeys(tea.KeyCtrlS.String()),
		key.WithHelp("ctrl+s", "settings"),
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
	Quit: key.NewBinding(
		key.WithKeys("q", tea.KeyCtrlC.String()),
		key.WithHelp("q", "quit"),
	),
	PullAll: key.NewBinding(
		key.WithKeys(tea.KeyCtrlP.String()),
		key.WithHelp("ctrl+p", "pull all"),
	),
}

func (k repoListKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, RepoListKeyMap.Select, "lazygit", "open with lazygit"),
		SetOnHelpType(helpType, RepoListKeyMap.VSCode, "editor", "open in editor"),
		SetOnHelpType(helpType, RepoListKeyMap.Shell, "shell", "shell in repo dir"),
		SetOnHelpType(helpType, RepoListKeyMap.CopyPath, "copy", "copy path"),
		SetOnHelpType(helpType, RepoListKeyMap.Refresh, "refresh", "refresh list"),
		SetOnHelpType(helpType, RepoListKeyMap.Clone, "clone", "clone new repos"),
		SetOnHelpType(helpType, RepoListKeyMap.PullAll, "pull all", "git pull every repo (skips conflicts)"),
		SetOnHelpType(helpType, RepoListKeyMap.Delete, "delete", "delete repo"),
		SetOnHelpType(helpType, RepoListKeyMap.Settings, "settings", "open settings"),
		SetOnHelpType(helpType, RepoListKeyMap.Quit, "quit", "quit"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Repo List Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Confirm Key Map

type confirmKeyMap struct {
	Proceed key.Binding
	Exit    key.Binding
}

var ConfirmKeyMap = confirmKeyMap{
	Proceed: key.NewBinding(
		key.WithKeys(tea.KeyCtrlP.String()),
		key.WithHelp("ctrl+p", "proceed"),
	),
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), "back"),
	),
}

func (k confirmKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, ConfirmKeyMap.Proceed, "proceed", "proceed with deleting"),
		SetOnHelpType(helpType, ConfirmKeyMap.Exit, "back", "back to menu"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Confirm Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Clone Repo Key Map

type cloneRepoKeyMap struct {
	Exit    key.Binding
	Proceed key.Binding
}

var CloneRepoKeyMap = cloneRepoKeyMap{
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), "back"),
	),
	Proceed: key.NewBinding(
		key.WithKeys(tea.KeyCtrlP.String()),
		key.WithHelp("ctrl+p", "proceed"),
	),
}

func (k cloneRepoKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(helpType, CloneRepoKeyMap.Proceed, "proceed", "proceed with cloning"),
		SetOnHelpType(helpType, CloneRepoKeyMap.Exit, "back", "back to menu"),
	}
	return func() []key.Binding { return bindsWithHelp }
}

// End "Clone Repo Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
