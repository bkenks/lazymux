package tui

import (
	"github.com/bkenks/lazymux/internal/pkg/tui/cloneRepoUI"
	"github.com/bkenks/lazymux/internal/pkg/tui/mainUI"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	stateMain sessionState = iota
	stateCloneRepo
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interfaces

type ModelManager struct {
	state 		sessionState
	main 		tea.Model
	cloneRepo 	tea.Model
	width		int
	height		int
}

func InitialModel() ModelManager {
	return ModelManager{
		state: stateMain,
		main: mainUI.InitialModel(),
		cloneRepo: cloneRepoUI.InitialModel(),
	}
}

func (m ModelManager) Init() tea.Cmd {
	return nil
}

func (m ModelManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case mainUI.MsgCtrlN:
		m.state = stateCloneRepo
	case cloneRepoUI.MsgEsc:
		m.state = stateMain
	case cloneRepoUI.MsgGhqGet:
		m.state = stateMain
		return m, mainUI.UpdateRepoList()
	}

	switch m.state {
	case stateMain:
		newMain, newCmd := m.main.Update(msg)
		mainModel, ok := newMain.(mainUI.Model)
		if !ok {
			panic("could not perform assertion on main model")
		}
		m.main = mainModel
		cmd = newCmd
	case stateCloneRepo:
		newCloneRepo, newCmd := m.cloneRepo.Update(msg)
		cloneRepoModel, ok := newCloneRepo.(cloneRepoUI.Model)
		if !ok {
			panic("could not perform assertion on cloneRepoUI model")
		}
		m.cloneRepo = cloneRepoModel
		cmd = newCmd
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ModelManager) View() string {
	switch m.state {
	case stateCloneRepo:
		return m.cloneRepo.View()
	default:
		return m.main.View()
	}
}

// End "Interfaces"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////