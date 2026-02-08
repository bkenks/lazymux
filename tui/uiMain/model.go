package uiMain

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface
//
// uiMain (tea.Model):
//	- Model (UI) for listing the repos from ghq and allowing the user to open them with Lazygit

type Model struct {
	List			list.Model
	frameWidth		int
	frameHeight		int
}


func New() *Model {
	x, y := constants.DocStyle.GetFrameSize()
	widthBuffer 	:= constants.WindowSize.Width-x
	heightBuffer 	:= constants.WindowSize.Height-y

	constants.RepoList = constants.RefreshRepos() // Pull new repos and set them to the global RepoList var
	newList := list.New(
		constants.RepoList, // []list.Item containing the parsed list of repos from ghq
		list.NewDefaultDelegate(), // Default list.Item styling
		widthBuffer, heightBuffer) // Height & Width
	newList.Title = "Repositories"
	newList.AdditionalShortHelpKeys = constants.DefaultKeyMap.Bindings
	newList.AdditionalFullHelpKeys = constants.DefaultKeyMap.Bindings

	return &Model{
		List: newList,
		frameWidth: widthBuffer,
		frameHeight: heightBuffer,
	}
}


func (m Model) Init() tea.Cmd { return nil }


func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd


	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedRepo := m.List.SelectedItem()
			var fullRepoPath string
			if repo, ok := selectedRepo.(constants.Repo); ok {
				fullRepoPath = constants.GetFullRepoPath(repo.Path)
			}
			
			cmd := commands.OpenLazygitAction(fullRepoPath)
			
			return m, cmd
		case "c":
			cmd = commands.CloneRepoDialog() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd
		case "d":
			cmd = commands.ConfirmDeleteDialog() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd
		case "b":
			cmd = commands.BulkCloneRepoDialog() // Send message to ModelManager to change state to CloneRepoUI
			return m, cmd 
		}
	case tea.WindowSizeMsg:
		x, y := constants.DocStyle.GetFrameSize()
		m.frameWidth 	= constants.WindowSize.Width-x
		m.frameHeight 	= constants.WindowSize.Height-y
		m.List.SetSize(m.frameWidth, m.frameHeight)
	}


	/////////////////////////////////////
	// Output
	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
	// End "Output"
	/////////////////////////////////////
}


func (m Model) View() string { return m.List.View() }

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
