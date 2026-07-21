package clonerepos

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	modeURLs mode = iota
	modeNamespace
)

type Model struct {
	mode mode

	textarea  textarea.Model
	adjHeight int
	adjWidth  int

	forges   []config.Forge
	forgeIdx int
	nsInput  textinput.Model
	fetching bool

	err error
}

func New(cfg config.Config) *Model {
	wBuffer, hBuffer := sizeBuffer()

	ti := textarea.New()
	ti.Placeholder = "git@github.com:ispenttoo/muchtimeonthis.git..."
	ti.Focus()
	ti.MaxHeight = hBuffer
	ti.MaxWidth = wBuffer
	ti.SetHeight(hBuffer)
	ti.SetWidth(wBuffer)

	ns := textinput.New()
	ns.Placeholder = "org-or-user"
	ns.CharLimit = 100
	ns.Width = wBuffer

	return &Model{
		textarea:  ti,
		forges:    cfg.Forges,
		nsInput:   ns,
		adjHeight: hBuffer,
		adjWidth:  wBuffer,
	}
}

func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		wBuffer, hBuffer := sizeBuffer()
		m.textarea.SetWidth(wBuffer)
		m.textarea.SetHeight(hBuffer)
		m.nsInput.Width = wBuffer
		return m, nil

	case events.NamespaceCloneFailed:
		m.fetching = false
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.CloneRepoKeyMap.Exit):
			return m, commands.SetState(domain.StateMain)
		case key.Matches(msg, constants.CloneRepoKeyMap.ToggleMode):
			m.toggleMode()
			return m, nil
		}

		if m.mode == modeNamespace {
			return m.updateNamespace(msg)
		}
		return m.updateURLs(msg)
	}

	return m, nil
}

func (m *Model) toggleMode() {
	if m.mode == modeURLs {
		m.mode = modeNamespace
		m.textarea.Blur()
		m.nsInput.Focus()
	} else {
		m.mode = modeURLs
		m.nsInput.Blur()
		m.textarea.Focus()
	}
	m.err = nil
}

func (m *Model) updateURLs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, constants.CloneRepoKeyMap.Proceed) {
		return m, commands.StartCloneReposCmd(m.textarea.Value())
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m *Model) updateNamespace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, constants.CloneRepoKeyMap.CycleForge):
		if len(m.forges) > 0 {
			m.forgeIdx = (m.forgeIdx + 1) % len(m.forges)
			m.err = nil
		}
		return m, nil
	case key.Matches(msg, constants.CloneRepoKeyMap.Proceed):
		return m.startFetch()
	}
	var cmd tea.Cmd
	m.nsInput, cmd = m.nsInput.Update(msg)
	return m, cmd
}

func (m *Model) startFetch() (tea.Model, tea.Cmd) {
	if m.fetching {
		return m, nil
	}
	if len(m.forges) == 0 {
		m.err = fmt.Errorf("no forges registered — add one first (F)")
		return m, nil
	}
	namespace := strings.TrimSpace(m.nsInput.Value())
	if namespace == "" {
		m.err = fmt.Errorf("enter a namespace/org name")
		return m, nil
	}
	m.err = nil
	m.fetching = true
	forge := m.forges[m.forgeIdx]
	return m, commands.ListNamespaceReposCmd(forge, namespace)
}

func headerView(mode mode) string {
	sub := "paste repository URLs here"
	if mode == modeNamespace {
		sub = "clone every repo in a namespace/org"
	}
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		"\n\n\n\n",
		styles.MenuTitle.Render("Repository Clone"),
		styles.MenuSubStyle.Render(sub),
	)
	return title
}

func footerView(mode mode) string {
	binds := constants.CloneRepoKeyMap.HelpBinds(constants.Short)()
	if mode == modeNamespace {
		binds = append(binds, constants.CloneRepoKeyMap.CycleForge)
	}
	helpKeys := styles.MenuHelpStyle.Render(domain.FormatBindingsInline(func() []key.Binding { return binds }))
	return helpKeys
}

func sizeBuffer() (w, h int) {
	headerHeight := lipgloss.Height(headerView(modeURLs))
	footerHeight := lipgloss.Height(footerView(modeURLs))

	widthBuffer := 2

	return constants.WindowSize.Width - widthBuffer,
		constants.WindowSize.Height - headerHeight - footerHeight - constants.FooterReservedLines
}

func (m *Model) View() string {
	var body string
	if m.mode == modeNamespace {
		body = m.namespaceView()
	} else {
		body = m.textarea.View()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerView(m.mode),
		body,
		footerView(m.mode),
	)

	placedContent := lipgloss.PlaceVertical(
		constants.WindowSize.Height,
		lipgloss.Center,
		content,
	)
	return placedContent
}

// namespaceView renders the forge picker + namespace input, plus a fetching
// indicator or error once a listing attempt has run.
func (m *Model) namespaceView() string {
	forgeLabel := "no forges registered"
	if len(m.forges) > 0 {
		f := m.forges[m.forgeIdx]
		forgeLabel = fmt.Sprintf("forge: %s (%s) · tab to cycle", f.Name, f.Host)
	}

	rows := []string{
		styles.Subtle(forgeLabel),
		m.nsInput.View(),
	}
	if m.fetching {
		rows = append(rows, styles.Subtle("fetching repos…"))
	}
	if m.err != nil {
		rows = append(rows, styles.ToastErrorStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
