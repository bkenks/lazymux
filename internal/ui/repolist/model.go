package repolist

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface
//
// uiMain (tea.Model):
//	- Model (UI) for listing the repos from ghq and allowing the user to open them with Lazygit

type Model struct {
	List     list.Model
	RepoList []list.Item
}

func New() *Model {
	commands.RefreshReposCmd()

	w, h := sizeBuffer()
	newList := list.New(
		[]list.Item{},             // []list.Item containing the parsed list of repos from ghq
		list.NewDefaultDelegate(), // Default list.Item styling
		w, h)                      // Width & Height
	newList.Title = "Repositories"
	newList.AdditionalShortHelpKeys = constants.RepoListKeyMap.HelpBinds(constants.Short)
	newList.AdditionalFullHelpKeys = constants.RepoListKeyMap.HelpBinds(constants.Full)

	return &Model{
		List: newList,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	/////////////////////////////////////
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		w, h := sizeBuffer()
		m.List.SetSize(w, h)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.RepoListKeyMap.Select):
			selectedRepo := m.List.SelectedItem()
			if repo, ok := selectedRepo.(domain.Repo); ok {
				fullRepoPath := GetFullRepoPath(repo.Path)

				cmds = append(
					cmds,
					commands.TeaCmdBuilder("lazygit", "-p", fullRepoPath),
				)
			}
		case key.Matches(msg, constants.RepoListKeyMap.Clone):
			cmds = append(
				cmds,
				commands.SetState(domain.StateCloneRepo),
			)
		case key.Matches(msg, constants.RepoListKeyMap.Delete):
			cmds = append(
				cmds,
				commands.SetState(domain.StateConfirmDelete),
			)
		}
	}
	/////////////////////////////////////

	/////////////////////////////////////
	// Output
	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
	// End "Output"
	/////////////////////////////////////
}

func sizeBuffer() (width, height int) {
	x, y := styles.DocStyle.GetFrameSize()
	widthBuffer := constants.WindowSize.Width - x
	heightBuffer := constants.WindowSize.Height - y
	return widthBuffer, heightBuffer
}

func (m *Model) View() string { return m.List.View() }

func (m *Model) UpdateRepoList(repoList []list.Item) {
	m.List.SetItems(repoList)
}

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
