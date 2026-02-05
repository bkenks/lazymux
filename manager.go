package main

import (
	"github.com/bkenks/lazymux/models"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	mainView sessionState = iota
	cloneRepo
)

type Manager struct {
	state		sessionState
	repoList	tea.Model
	cloneRepo	tea.Model
	Model		tea.Model
}

// Init initialises the Manager on program load. It partly implements the tea.Model interface.
func (m *Manager) Init() tea.Cmd {
	m.state = mainView
	m.repoList = models.InitialRepoListModel()
	m.cloneRepo = models.InitialCloneRepoModel()
	m.Model = m.repoList
	return nil
}

// Update handles event and manages internal state. It partly implements the tea.Model interface.
func (m *Manager) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var fg, bg tea.Model
	var fgCmd, bgCmd tea.Cmd
	
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
			
		case "ctrl+n":
			m.state = cloneRepo
			return m, nil

		case "esc":
			m.state = mainView
			return m, nil
		}
	}

	
	if m.state == cloneRepo {
		fg, fgCmd = m.cloneRepo.Update(message)
		m.cloneRepo = fg
	} else {
		bg, bgCmd = m.repoList.Update(message)
		m.repoList = bg
	}

	cmds := []tea.Cmd{}
	cmds = append(cmds, fgCmd, bgCmd)

	return m, tea.Batch(cmds...)
}

// View applies and styling and handles rendering the view. It partly implements the tea.Model
// interface.
func (m *Manager) View() string {
	if m.state == cloneRepo {
		return m.cloneRepo.View()
	}
	return m.repoList.View()
}