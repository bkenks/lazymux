package uiMain

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/tui/commands"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Styling

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// End "Styling"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////



///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Types


// End "Types"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////



///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interfaces

/////////////////////////////////////
//// Interface: tea.Model
type Model struct {
	windowWidth  int
	windowHeight int
	List list.Model
}

func InitialModel() Model {
	return Model{
		List: list.New(listRepos(), list.NewDefaultDelegate(), 0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case commands.MsgUpdateProjectList:
		m.List.SetItems(listRepos())
	/////////////////////////////////////
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedRepo := m.List.SelectedItem()
			var fullRepoPath string
			if repo, ok := selectedRepo.(repo); ok {
				fullRepoPath = getFullRepoPath(repo.path)
			}
			
			cmd := openLazygit(fullRepoPath)
			
			return m, cmd
		case "ctrl+n":
			cmd = commands.CloneRepoDialog() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd
		}
	/////////////////////////////////////
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	/////////////////////////////////////
	}

	/////////////////////////////////////
	// Output
	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
	// End "Output"
	/////////////////////////////////////
}

func (m Model) View() string {
	return docStyle.Render(m.List.View())
}
//// End "Interface: tea.Model"
/////////////////////////////////////

/////////////////////////////////////
//// Interface: List.Item
type repo struct {
	name, path string
}

func (r repo) Title() string				{ return r.name }
func (r repo) Description() string			{ return r.path }
func (r repo) FilterValue() string			{ return r.name }
//// End "Interface: List.Item"
/////////////////////////////////////

// End "Interfaces"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Cmds & Msgs



// End "Cmds"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions

func listRepos() ([]list.Item) {
	cmd := exec.Command("ghq", "list")
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting repo List:", err)
		os.Exit(1)
	}

	/////////////////////////////////////

	// string(out) → "github.com/user/repo1\ngithub.com/user/repo2\n"
	// strings.TrimSpace(...) → removes the final \n, giving "github.com/user/repo1\ngithub.com/user/repo2"
	// strings.Split(..., "\n") → gives a slice:
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	repos := make([]list.Item, 0, len(lines)) // preallocate slice

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "/")
		nameFromSplit := parts[len(parts)-1] // last element is the repo name
		repos = append(repos, repo{
			name: nameFromSplit,
			path: line,
		})
	}

	return repos
}

func getFullRepoPath(repo string) (string) {
	cmd := exec.Command("ghq", "List", "--full-path", repo)
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting repo path:", err)
		os.Exit(1)
	}

	/////////////////////////////////////

	path := strings.TrimSpace(string(out))
	return path
}

func openLazygit(path string) tea.Cmd {
	c := exec.Command("lazygit", "-p", path)

	type lgFinishedMsg struct { err error }

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return lgFinishedMsg{err: err}
	})
	
	return tea.Cmd(cmd)
}

// End "Functions"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
