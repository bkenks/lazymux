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

func RefreshRepos() ([]list.Item) {
	cmd := exec.Command("ghq", "list") // Call ghq to list repositories
	out, err := cmd.Output()

	if err != nil { // Fail-Safe
		fmt.Println("Error getting Repo List:", err)
		os.Exit(1)
	}


	// string(out) → "github.com/user/Repo1\ngithub.com/user/Repo2\n"
	// strings.TrimSpace(...) → removes the final \n, giving "github.com/user/Repo1\ngithub.com/user/Repo2"
	// strings.Split(..., "\n") → splits into strings on "\n"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	Repos := make([]list.Item, 0, len(lines)) // preallocate slice (i.e. set array size)


	/////////////////////////////////////
	// Format lines (list of repos like "github.com/user/Repo1") into a []list.Item
	for _, line := range lines {
		if line == "" { // Fail-Safe
			continue
		}
		
		parts := strings.Split(line, "/") // split path ("github.com/user/Repo1") by "/"
		nameFromSplit := parts[len(parts)-1] // grab the last element from the split (which is the repo name)

		Repos = append(Repos, Repo{ // Add to a []list.Item (array of `list.Item`s)
			Name: nameFromSplit,
			Path: line,
		})
	}
	//
	/////////////////////////////////////

	return Repos
}

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
    C key.Binding
    D key.Binding
}

func (k KeyMap) Bindings() []key.Binding {
	return []key.Binding{
		k.C,
		k.D,
	}
}

var DefaultKeyMap = KeyMap{
    C: key.NewBinding(
        key.WithKeys("c"),        // actual keybindings
        key.WithHelp("c", "clone repo"), // corresponding help text
    ),
	D: key.NewBinding(
        key.WithKeys("d"),        // actual keybindings
        key.WithHelp("d", "delete repo"), // corresponding help text
    ),
}


// End "Bindings"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////