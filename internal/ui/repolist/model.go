package repolist

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	List     list.Model
	RepoList []list.Item
}

func New() *Model {
	w, h := SizeBuffer()
	// Height 3 so a row shows the repo name, its path, and the "forge:" line.
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)
	newList := list.New(
		[]list.Item{},
		delegate,
		w, h,
	)
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

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		w, h := SizeBuffer()
		m.List.SetSize(w, h)

	case tea.KeyMsg:
		// Suppress repo-action keybinds while the list's filter input has focus —
		// otherwise typing "r" in a search filter triggers a refresh.
		if m.List.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, constants.RepoListKeyMap.Select):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.LazygitCmd(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.Clone):
			cmds = append(cmds, commands.SetState(domain.StateCloneRepo))

		case key.Matches(msg, constants.RepoListKeyMap.Delete):
			if ConvertToRepoType(m.List.SelectedItem()).AbsPath == "" {
				break
			}
			cmds = append(cmds, commands.SetState(domain.StateConfirmDelete))

		case key.Matches(msg, constants.RepoListKeyMap.VSCode):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.OpenInVSCode(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.Settings):
			cmds = append(cmds, commands.SetState(domain.StateSettings))

		case key.Matches(msg, constants.RepoListKeyMap.Refresh):
			cmds = append(cmds, commands.RefreshReposCmd())

		case key.Matches(msg, constants.RepoListKeyMap.CopyPath):
			cmds = append(cmds, commands.CopyPathCmd(AbsRepoPath(m.List.SelectedItem())))

		case key.Matches(msg, constants.RepoListKeyMap.Shell):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.OpenShellCmd(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.PullAll):
			cmds = append(cmds,
				m.List.NewStatusMessage("Pulling all repos…"),
				commands.PullAllReposCmd(),
			)

		case key.Matches(msg, constants.RepoListKeyMap.Forges):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.Path == "" {
				break
			}
			cmds = append(cmds, commands.OpenRepoForgesCmd(repo.Path))

		case key.Matches(msg, constants.RepoListKeyMap.Registry):
			cmds = append(cmds, commands.SetState(domain.StateForgeRegistry))
		}

	case events.PullAllReposComplete:
		cmds = append(cmds, m.List.NewStatusMessage(formatPullSummary(msg)))
	}

	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string { return m.List.View() }

// UpdateRepoList replaces the list's items. It returns the command from
// list.SetItems, which must be run so the filtered view is recomputed when a
// filter is applied — dropping it leaves the filter phrase set but the results
// empty ("0 results found") after a refresh (e.g. returning from lazygit).
func (m *Model) UpdateRepoList(repoList []list.Item) tea.Cmd {
	return m.List.SetItems(repoList)
}
