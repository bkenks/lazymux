package app

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/clonerepos"
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
	state         domain.SessionState
	main          repolist.Model
	confirmDelete confirm.Model
	clonerepos    clonerepos.Model

	active tea.Model
}

func New() *ModelManager {
	m := &ModelManager{
		main:          *repolist.New(), // Main Model (List)
		confirmDelete: *confirm.New(),  // Delete Repo Confirmation (Dialog)
		clonerepos:    *clonerepos.New(),
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

	case events.Event:
		switch msg := msg.(type) {

		//// State Manager
		case events.SetState:
			m.state = msg.State

			// Initialization for each state
			switch m.state {

			case domain.StateMain:
				m.active = &m.main

			case domain.StateConfirmDelete:
				m.confirmDelete = *confirm.New()
				selectedRepo := m.main.List.SelectedItem()
				if repo, ok := selectedRepo.(domain.Repo); ok {
					m.confirmDelete.RepoPath = repo.Path
				}
				m.active = &m.confirmDelete

			case domain.StateCloneRepo:
				m.clonerepos = *clonerepos.New()
				m.active = &m.clonerepos
			}

		//// Trigger: Repos are being cloned.
		case events.StartRepoClone:
			m.clonerepos.RepoCounter = 0
			m.clonerepos.TotalRepos = len(msg.RepoUrls)

			cmds = append(cmds, commands.CloneReposExecCmd(msg.RepoUrls))
		case events.CloneRepoExec:
			if m.clonerepos.RepoCounter < m.clonerepos.TotalRepos {
				m.clonerepos.RepoCounter++
			}
			if m.clonerepos.RepoCounter == m.clonerepos.TotalRepos {
				cmds = append(cmds,
					commands.RefreshReposCmd(),
					commands.SetState(domain.StateMain),
				)
			}
		/// Trigger: A repo has been deleted.
		case events.RepoDeleted:
			cmds = append(cmds, commands.RefreshReposCmd())
		//// - Trigger: Completed pulling a list of repos from ghq
		case events.ReposRefreshed:
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
