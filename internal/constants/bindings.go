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

var unsetText = "not set"

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
			unsetText,
		),
	),
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(
			tea.KeyEsc.String(),
			unsetText,
		),
	),
}

func (k defaultKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(
			helpType,             // Short or Full Help
			DefaultKeyMap.Select, // key.Binding
			"select",             // Short Help
			"select",             // Full Help
		),
		SetOnHelpType(
			helpType,
			DefaultKeyMap.Exit,
			"exit",
			"exit",
		),
	}

	return func() []key.Binding { return bindsWithHelp }
}

// End "Default Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Repo List Key Map

type repoListKeyMap struct {
	Select key.Binding
	Clone  key.Binding
	Delete key.Binding
	VSCode key.Binding
}

var RepoListKeyMap = repoListKeyMap{
	Select: key.NewBinding(
		key.WithKeys("u"),            // actual keybindings
		key.WithHelp("u", unsetText), // corresponding help text
	),
	Clone: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", unsetText),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", unsetText),
	),
	VSCode: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", unsetText),
	),
}

func (k repoListKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(
			helpType,              // Short or Full Help
			RepoListKeyMap.Select, // key.Binding
			"lazygit",             // Short Help
			"open with lazygit",   // Full Help
		),
		SetOnHelpType(
			helpType,
			RepoListKeyMap.Clone,
			"clone",
			"clone new repos",
		),
		SetOnHelpType(
			helpType,
			RepoListKeyMap.Delete,
			"delete",
			"delete repo",
		),
		SetOnHelpType(
			helpType,
			RepoListKeyMap.VSCode,
			"vscode",
			"open in vscode",
		),
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
		key.WithHelp("ctrl+p", unsetText),
	),
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), unsetText),
	),
}

func (k confirmKeyMap) HelpBinds(helpType HelpType) func() []key.Binding {
	bindsWithHelp := []key.Binding{
		SetOnHelpType(
			helpType,                // Short or Full Help
			ConfirmKeyMap.Proceed,   // key.Binding
			"proceed",               // Short Help
			"proceed with deleting", // Full Help
		),
		SetOnHelpType(
			helpType,
			ConfirmKeyMap.Exit,
			"back",
			"back to menu",
		),
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
		SetOnHelpType(
			helpType,                // Short or Full Help
			CloneRepoKeyMap.Proceed, // key.Binding
			"proceed",               // Short Help
			"proceed with cloning",  // Full Help
		),
		SetOnHelpType(
			helpType,
			CloneRepoKeyMap.Exit,
			"back",
			"back to menu",
		),
	}

	return func() []key.Binding { return bindsWithHelp }
}

// End "Clone Repo Key Map"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
