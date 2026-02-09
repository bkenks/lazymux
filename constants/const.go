package constants

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Global Variables

var (
	WindowSize tea.WindowSizeMsg
	RepoList []list.Item
)

// End "Global Variables"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Styling

var (
	DarkPink = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	DullGrey = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}
	Purple = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	DarkPurple = lipgloss.Color("62")
	White = lipgloss.Color("230")

	ButtonStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	SelectedButton = ButtonStyle.
		Background(DarkPurple).
		Foreground(White).
		Bold(true)

	UnselectedButton = ButtonStyle.
		Background(DullGrey).
		Foreground(lipgloss.Color("250"))

	DocStyle = lipgloss.NewStyle().
		Margin(3, 1)

	DialogStyle = lipgloss.NewStyle().
		Padding(0, 6, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DullGrey)

	DialogTitleStyle = lipgloss.NewStyle().Bold(true)

	Title = lipgloss.NewStyle().
		Background(DarkPurple).
		Foreground(White).
		Padding(0, 1).
		Bold(true).
		MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
		MarginBottom(1)
)

// End "Styling"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface
//
// Repo (list.Item):
//	- Represents a repository

type Repo struct {
	Name, Path string
}

func (r Repo) Title() string				{ return r.Name }
func (r Repo) Description() string			{ return r.Path }
func (r Repo) FilterValue() string			{ return r.Name }

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////




///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions



func GetFullRepoPath(repo string) (string) {
	cmd := exec.Command("ghq", "list", "--full-path", repo)
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting repo path:", repo)
		os.Exit(1)
	}

	/////////////////////////////////////

	path := strings.TrimSpace(string(out))
	return path
}

// End "Functions"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////



///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Bindings

type KeyMap struct {
	Esc key.Binding
	Select key.Binding
	Left key.Binding
	Right key.Binding
    C key.Binding
    D key.Binding
	Confirm key.Binding
}

func (k KeyMap) Bindings() []key.Binding {
	return []key.Binding{
		k.Esc,
		k.Select,
		k.Left,
		k.Right,
		k.C,
		k.D,
		k.Confirm,
	}
}

var DefaultKeyMap = KeyMap{
	Select: key.NewBinding(
		key.WithKeys(tea.KeyEnter.String(), tea.KeySpace.String()),
		key.WithHelp("enter/space", "select"),
	),
	Esc: key.NewBinding(
		key.WithKeys(),
		key.WithHelp(tea.KeyEsc.String(), "back"),
	),
}

var UIMainKeyMap = KeyMap{
	Select: key.NewBinding(
		key.WithKeys(DefaultKeyMap.Select.Keys()...),
		key.WithHelp("enter/space", "open in lazygit"),
	),
    C: key.NewBinding(
        key.WithKeys("c"),        // actual keybindings
        key.WithHelp("c", "clone repo"), // corresponding help text
    ),
	D: key.NewBinding(
        key.WithKeys("d"),        // actual keybindings
        key.WithHelp("d", "delete repo"), // corresponding help text
    ),
}

var UIConfirm = KeyMap{
	Select: key.NewBinding(
		key.WithKeys(DefaultKeyMap.Select.Keys()...),
		key.WithHelp("enter/space", "select"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	),
	Esc: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), "back"),
	),
}

var UICloneRepo = KeyMap{
	Esc: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp(tea.KeyEsc.String(), "back"),
	),
	Confirm: key.NewBinding(
		key.WithKeys(tea.KeyCtrlJ.String()),
		key.WithHelp("ctrl+j", "clone repos"),
	),
}

// End "Bindings"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////