package main

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

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

//// Interface: list.Item
type repo struct {
	name, path string
}

func (r repo) Title() string				{ return r.name }
func (r repo) Description() string			{ return r.path }
func (r repo) FilterValue() string			{ return r.name }
//// End "Interface: list.Item"

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg:= msg.(type) {

		/////////////////////////////////////
		case tea.KeyMsg:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "enter" {
				selectedRepo := m.list.SelectedItem()
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
			m.list.SetSize(msg.Width-h, msg.Height-v)
		/////////////////////////////////////

	}

	/////////////////////////////////////

	// Output
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
	// End "Output"
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}



func main() {
	m := model{
		list: list.New(listRepos(), list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = "GitHub Repos"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _,err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}