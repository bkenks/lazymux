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
	Select key.Binding
	Left key.Binding
	Right key.Binding
    C key.Binding
    D key.Binding
}

func (k KeyMap) Bindings() []key.Binding {
	return []key.Binding{
		k.Select,
		k.Left,
		k.Right,
		k.C,
		k.D,
	}
}

var UIMainKeyMap = KeyMap{
	Select: key.NewBinding(
		key.WithKeys("enter", "space"),
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
		key.WithKeys("enter", "space"),
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
}


// End "Bindings"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////


// case tea.KeyMsg:
// 		switch msg.String() {

// 		case "left", "h", "up", "k":
// 			m.cursor = choiceYes

// 		case "right", "l", "down", "j":
// 			m.cursor = choiceNo

// 		case "enter":
// 			if m.cursor == choiceYes {
// 				cmds = append(
// 					cmds,
// 					commands.DeleteRepoCmd(m.RepoPath),
// 					commands.SetState(commands.StateMain),
// 				)
// 			}
// 		case "esc":
// 			cmds = append(
// 				cmds,
// 				commands.SetState(commands.StateMain),
// 			)
// 		}
// 	}