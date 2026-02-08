package uiBulkCloneRepo

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/tui/commands"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type (
	errMsg error
)

type Model struct {
	viewport    viewport.Model
	Messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	reposRaw	[]string
	RepoCount	int
	err         error
}

func New() *Model {
	ta := textarea.New()
	ta.Placeholder = "Enter your repository URL here..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(100)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(100, 5)
	vp.SetContent(`Welcome to the Bulk Repo Cloning room!
Type a repository url and press Enter to add to the clone list.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &Model{
		textarea:    ta,
		Messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		reposRaw: 	 []string{},
		RepoCount:	 0,
		err:         nil,
	}
}

func (m Model) Init() tea.Cmd { return textarea.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		cmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.Messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.Messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			cmd = commands.BulkCloneRepoAction(m.reposRaw)
			return m, cmd
		case tea.KeyEnter:
			m.reposRaw = append(m.reposRaw, m.textarea.Value())
			m.Messages = append(m.Messages, m.senderStyle.Render(m.textarea.Value()))
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.Messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}