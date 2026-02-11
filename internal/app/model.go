package app

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/bulkclone"
	"github.com/bkenks/lazymux/internal/ui/confirm"
	"github.com/bkenks/lazymux/internal/ui/repolist"
	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface: tea.Model
//
// ModelManager:
//	- Model for managing sub-Models (i.e other UI/Views/Screens)

type ModelManager struct {
	state          commands.SessionState
	main           repolist.Model
	confirmDelete  confirm.Model
	bulkCloneRepos bulkclone.Model

	active tea.Model
}

func New() *ModelManager {
	m := &ModelManager{
		main:           *repolist.New(), // Main Model (List)
		confirmDelete:  *confirm.New(),  // Delete Repo Confirmation (Dialog)
		bulkCloneRepos: *bulkclone.New(),
	}

	m.active = &m.main

	return m
}

func (m *ModelManager) Init() tea.Cmd { return commands.RefreshReposCmd() }

func (m *ModelManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	/////////////////////////////////////
	// UI Manager
	switch msg := msg.(type) {

	//// Reactive Window Sizing
	case tea.WindowSizeMsg:
		constants.WindowSize = msg

	case commands.CommandsMsg:
		switch msg := msg.(type) {

		//// State Manager
		case commands.MsgSetState:
			m.state = msg.State

			// Initialization for each state
			switch m.state {

			case commands.StateMain:
				m.active = &m.main

			case commands.StateConfirmDelete:
				m.confirmDelete = *confirm.New()
				selectedRepo := m.main.List.SelectedItem()
				if repo, ok := selectedRepo.(domain.Repo); ok {
					m.confirmDelete.RepoPath = repo.Path
				}
				m.active = &m.confirmDelete

			case commands.StateBulkCloneRepos:
				m.bulkCloneRepos = *bulkclone.New()
				m.active = &m.bulkCloneRepos
			}

		//// Trigger: Repos are being cloned.
		case commands.MsgStartRepoClone:
			m.bulkCloneRepos.RepoCounter = 0
			m.bulkCloneRepos.TotalRepos = len(msg.RepoUrls)

			cmds = append(cmds, commands.CloneReposExecCmd(msg.RepoUrls))
		case commands.MsgRepoCloned:
			if m.bulkCloneRepos.RepoCounter < m.bulkCloneRepos.TotalRepos {
				m.bulkCloneRepos.RepoCounter++
			}
			if m.bulkCloneRepos.RepoCounter == m.bulkCloneRepos.TotalRepos {
				cmds = append(cmds,
					commands.RefreshReposCmd(),
					commands.SetState(commands.StateMain),
				)
			}
		/// Trigger: A repo has been deleted.
		case commands.MsgRepoDeleted:
			cmds = append(cmds, commands.RefreshReposCmd())
		//// - Trigger: Completed pulling a list of repos from ghq
		case commands.MsgReposRefreshed:
			m.main.UpdateRepoList(msg.RepoList)
		}
	}
	// End "UI Manager"
	/////////////////////////////////////

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *ModelManager) View() string {
	return styles.DocStyle.Render(m.active.View())
}

// End "Interface: tea.Model"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
