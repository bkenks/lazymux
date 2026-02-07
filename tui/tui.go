package tui

import (
	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
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

type sessionState int

const (
	stateMain sessionState = iota
	stateConfirmDelete
	stateCloneRepo
)

type ModelManager struct {
	state 			sessionState
	main 			uiMain.Model
	confirmDelete 	uiConfirm.Model
	cloneRepo 		uiCloneRepo.Model
}

func New() *ModelManager {
	return &ModelManager{  // Not sure why using pointer but I saw it somewhere (prolly performace). Don't have to use, works without
		state: stateMain, // Initial state of TUI
		main: *uiMain.New(), // Main Model (List)
		cloneRepo: *uiCloneRepo.New(), // CloneRepo Model (Dialog)
		confirmDelete: *uiConfirm.New(), // Delete Repo Confirmation (Dialog)
	}
}

func (m ModelManager) Init() tea.Cmd { return nil }

func (m ModelManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	/////////////////////////////////////
	// Set func-scoped cmds
	var cmd tea.Cmd
	var cmds []tea.Cmd

	/////////////////////////////////////
	// UI Manager
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		constants.WindowSize = msg
	case commands.MsgCloneRepoDialog:
		m.cloneRepo = *uiCloneRepo.New()
		m.state = stateCloneRepo
	case commands.MsgConfirmDeleteDialog:
		m.confirmDelete = *uiConfirm.New()
		m.state = stateConfirmDelete
	case commands.MsgQuitRepoDialog:
		m.state = stateMain
	case commands.MsgConfirmDeleteDialogQuit:
		m.state = stateMain
	/////////////////////////////////////
	case commands.MsgGhqGet:
		m.state = stateMain
		constants.RepoList = constants.RefreshRepos()
		m.main = *uiMain.New()
	case commands.MsgGhqRm:
		constants.RepoList = constants.RefreshRepos()
		m.main = *uiMain.New()
	}
	// End "UI Manager"
	/////////////////////////////////////


	/////////////////////////////////////
	// Input & Model Router
	switch m.state {
	case stateMain:
		newMain, newCmd := m.main.Update(msg)
		mainModel, ok := newMain.(uiMain.Model)
		if !ok {
			panic("could not perform assertion on main model")
		}
		m.main = mainModel
		cmd = newCmd
	case stateConfirmDelete:
		newCD, newCmd := m.confirmDelete.Update(msg)
		cdModel, ok := newCD.(uiConfirm.Model)
		if !ok {
			panic("could not perform assertion on confirm model")
		}

		selectedRepo := m.main.List.SelectedItem()
		if repo, ok := selectedRepo.(constants.Repo); ok {
			cdModel.RepoPath = repo.Path
		}
		
		m.confirmDelete = cdModel
		cmd = newCmd
	case stateCloneRepo:
		newCloneRepo, newCmd := m.cloneRepo.Update(msg)
		cloneRepoModel, ok := newCloneRepo.(uiCloneRepo.Model)
		if !ok {
			panic("could not perform assertion on uiCloneRepo model")
		}
		m.cloneRepo = cloneRepoModel
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
	case stateCloneRepo:
		currentView = m.cloneRepo.View()
	case stateConfirmDelete:
		currentView = m.confirmDelete.View()
	default:
		currentView = m.main.View()
	}

	return constants.DocStyle.Render(currentView)
}

// End "Interface: tea.Model"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////