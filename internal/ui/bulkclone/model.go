package bulkclone

import (
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// reposRaw	[]string
// RepoCount	int

// TODO: CHANGE THIS TO USE ANOTHER LIST

type errMsg error

type Model struct {
	textarea    textarea.Model
	adjHeight   int // Something weird occurs in the textarea model so we have
	adjWidth    int // to set these two to obscure values to fill the window properly
	err         error
	RepoCounter int
	TotalRepos  int
}

func New() *Model {
	wBuffer, hBuffer := sizeBuffer()

	ti := textarea.New()
	ti.Placeholder = "git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.MaxHeight = hBuffer
	ti.MaxWidth = wBuffer
	ti.SetHeight(hBuffer)
	ti.SetWidth(wBuffer)

	return &Model{
		textarea:  ti,
		adjHeight: hBuffer,
		adjWidth:  wBuffer,
		err:       nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		wBuffer, hBuffer := sizeBuffer()
		m.textarea.SetWidth(wBuffer)
		m.textarea.SetHeight(hBuffer)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.CloneRepoKeyMap.Exit):
			cmds = append(cmds,
				commands.SetState(commands.StateMain),
			)
		case key.Matches(msg, constants.CloneRepoKeyMap.Proceed):
			cmds = append(cmds,
				commands.StartCloneReposCmd(m.textarea.Value()),
			)
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func headerView() string {
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		"\n\n\n\n",
		styles.MenuTitle.Render("Repository Clone"),
		styles.MenuSubStyle.
			Render("paste repository URLs here"),
	)

	return title
}

func footerView() string {
	helpKeys := styles.MenuHelpStyle.
		Render(
			domain.FormatBindingsInline(
				constants.CloneRepoKeyMap.HelpBinds(constants.Short),
			),
		)

	return helpKeys
}

func sizeBuffer() (w, h int) {
	headerHeight := lipgloss.Height(headerView())
	footerHeight := lipgloss.Height(footerView())

	heightBuffer := 0
	widthBuffer := 2

	return (constants.WindowSize.Width - widthBuffer),
		(constants.WindowSize.Height - headerHeight - footerHeight - heightBuffer)
}

func (m *Model) View() string {

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerView(),
		m.textarea.View(),
		footerView(),
	)

	placedContent := lipgloss.PlaceVertical(
		constants.WindowSize.Height,
		lipgloss.Center,
		content,
	)

	return placedContent
}
