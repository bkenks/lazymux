package uiBulkCloneRepo

import (
	"fmt"

	"github.com/bkenks/lazymux/constants"
	"github.com/bkenks/lazymux/tui/commands"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// reposRaw	[]string
// RepoCount	int

// TODO: CHANGE THIS TO USE ANOTHER LIST


type errMsg error

type Model struct {
	textarea textarea.Model
	err      error
}

func New() *Model {
	ti := textarea.New()
	ti.Placeholder = "git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.MaxHeight = constants.WindowSize.Height - 30
	ti.MaxWidth = constants.WindowSize.Width - 10
	ti.SetHeight(ti.MaxHeight)
	ti.SetWidth(ti.MaxWidth)

	return &Model{
		textarea: ti,
		err:      nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			cmd = commands.BulkCloneRepoAction(m.textarea.Value())
			return m, cmd
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	content := fmt.Sprintf(
		"Enter one or more repository URLs:\n\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+c clone • esc back)",
	) + "\n\n"

	placedContent := lipgloss.Place(
		constants.WindowSize.Width,
		constants.WindowSize.Height,
		lipgloss.Left,
		lipgloss.Center,
		content,
	)
	
	return placedContent
}