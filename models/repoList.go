package models

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
// Interfaces

/////////////////////////////////////
//// Interface: tea.Model
type RepoList struct {
	windowWidth  int
	windowHeight int
	Model list.Model
}

func InitialRepoListModel() *RepoList {
	return &RepoList{
		Model: list.New(listRepos(), list.NewDefaultDelegate(), 0, 0),
	}
}

func (m *RepoList) Init() tea.Cmd {
	// m.Model = list.New(listRepos(), list.NewDefaultDelegate(), 0, 0)
	return nil
}

func (m *RepoList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

		/////////////////////////////////////
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				selectedRepo := m.Model.SelectedItem()
				var fullRepoPath string
				if repo, ok := selectedRepo.(repo); ok {
					fullRepoPath = getFullRepoPath(repo.path)
				}
				
				cmd := openLazygit(fullRepoPath)
				
				return m, cmd
			}
		/////////////////////////////////////
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.Model.SetSize(msg.Width-h, msg.Height-v)
			// m.windowWidth = msg.Width
			// m.windowHeight = msg.Height
		/////////////////////////////////////
	}

	/////////////////////////////////////

	// Output
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
	// End "Output"
}

func (m *RepoList) View() string {
	return docStyle.Render(m.Model.View())
}
//// End "Interface: tea.Model"
/////////////////////////////////////

/////////////////////////////////////
//// Interface: list.Item
type repo struct {
	name, path string
}

func (r repo) Title() string				{ return r.name }
func (r repo) Description() string			{ return r.path }
func (r repo) FilterValue() string			{ return r.name }
//// End "Interface: list.Item"
/////////////////////////////////////

// End "Interfaces"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////


///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Functions

func listRepos() ([]list.Item) {
	cmd := exec.Command("ghq", "list")
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error getting repo list:", err)
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
	cmd := exec.Command("ghq", "list", "--full-path", repo)
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
