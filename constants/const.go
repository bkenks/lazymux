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
	purple = lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}
	test = lipgloss.Color("#ffbb00")
)

var DocStyle = lipgloss.NewStyle().
	Margin(3, 1)

var DialogStyle = lipgloss.NewStyle().
	Padding(2, 2).
	Border(lipgloss.NormalBorder(), false, false, false, true).
	BorderForeground(purple)

var DialogTitleStyle = lipgloss.NewStyle().Bold(true)

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
    CtrlShiftD key.Binding
}

func (k KeyMap) Bindings() []key.Binding {
	return []key.Binding{
		k.C,
		k.CtrlShiftD,
	}
}

var DefaultKeyMap = KeyMap{
    C: key.NewBinding(
        key.WithKeys("c"),        // actual keybindings
        key.WithHelp("c", "clone repo"), // corresponding help text
    ),
	CtrlShiftD: key.NewBinding(
        key.WithKeys("ctrl+shift+d"),        // actual keybindings
        key.WithHelp("ctrl+shift+d", "delete repo"), // corresponding help text
    ),
}


// End "Bindings"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////