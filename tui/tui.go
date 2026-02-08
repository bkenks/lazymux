package tui

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
	"github.com/bkenks/lazymux/tui/uiBulkCloneRepo"
	"github.com/bkenks/lazymux/tui/uiCloneRepo"
	"github.com/bkenks/lazymux/tui/uiConfirm"
	"github.com/bkenks/lazymux/tui/uiMain"
	tea "github.com/charmbracelet/bubbletea"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface: tea.Model
//
// ModelManager:
//	- Model for managing sub-Models (i.e other UI/Views/Screens)

type ModelManager struct {
	state 			commands.SessionState
	main 			uiMain.Model
	confirmDelete 	uiConfirm.Model
	cloneRepo 		uiCloneRepo.Model
	bulkCloneRepos	uiBulkCloneRepo.Model
}

func New() *ModelManager {
	return &ModelManager{  // Not sure why using pointer but I saw it somewhere (prolly performace). Don't have to use, works without
		state: commands.StateMain, // Initial state of TUI
		main: *uiMain.New(), // Main Model (List)
		cloneRepo: *uiCloneRepo.New(), // CloneRepo Model (Dialog)
		confirmDelete: *uiConfirm.New(), // Delete Repo Confirmation (Dialog)
		bulkCloneRepos: *uiBulkCloneRepo.New(),
	}
}

func (m ModelManager) Init() tea.Cmd { return nil }

func (m ModelManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd


	/////////////////////////////////////
	// UI Manager
	switch msg := msg.(type) {
	// Reactive Window Sizing
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
	
	// State Manager
	case commands.CommandsMsg:
		switch msg := msg.(type) {
		case commands.MsgSetState:
			m.state = msg.State

			// Initialization for each state
			switch m.state {
			case commands.StateMain:
				constants.RepoList = constants.RefreshRepos()
				m.main = *uiMain.New()
			case commands.StateConfirmDelete:
				m.confirmDelete = *uiConfirm.New()
				selectedRepo := m.main.List.SelectedItem()
				if repo, ok := selectedRepo.(constants.Repo); ok {
					m.confirmDelete.RepoPath = repo.Path
				}
			case commands.StateBulkCloneRepos:
				m.bulkCloneRepos = *uiBulkCloneRepo.New()
			}
		}
	}
	// End "UI Manager"
	/////////////////////////////////////


	/////////////////////////////////////
	// Input & Model Router
	switch m.state {
	case commands.StateMain:
		newMain, newCmd := m.main.Update(msg)
		mainModel, ok := newMain.(uiMain.Model)
		if !ok {
			panic("could not perform assertion on main model")
		}
		m.main = mainModel
		cmd = newCmd
	case commands.StateConfirmDelete:
		newCD, newCmd := m.confirmDelete.Update(msg)
		cdModel, ok := newCD.(uiConfirm.Model)
		if !ok {
			panic("could not perform assertion on confirm model")
		}
		m.confirmDelete = cdModel
		cmd = newCmd
	case commands.StateCloneRepo:
		newCloneRepo, newCmd := m.cloneRepo.Update(msg)
		cloneRepoModel, ok := newCloneRepo.(uiCloneRepo.Model)
		if !ok {
			panic("could not perform assertion on uiCloneRepo model")
		}
		m.cloneRepo = cloneRepoModel
		cmd = newCmd
	case commands.StateBulkCloneRepos:
		newBulkCloneRepo, newCmd := m.bulkCloneRepos.Update(msg)
		BulkCloneRepoModel, ok := newBulkCloneRepo.(uiBulkCloneRepo.Model)
		if !ok {
			panic("could not perform assertion on uiCloneRepo model")
		}
		m.bulkCloneRepos = BulkCloneRepoModel
		cmd = newCmd
	}
	// End "Input Router"
	/////////////////////////////////////


	/////////////////////////////////////
	// Pass Model and Cmds from sub-Models to self (main update/event loop)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ModelManager) View() string {
	var currentView string

	switch m.state {
	case commands.StateCloneRepo:
		currentView = m.cloneRepo.View()
	case commands.StateConfirmDelete:
		currentView = m.confirmDelete.View()
	case commands.StateBulkCloneRepos:
		currentView = m.bulkCloneRepos.View()
	default:
		currentView = m.main.View()
	}

	return constants.DocStyle.Render(currentView)
}

// End "Interface: tea.Model"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////