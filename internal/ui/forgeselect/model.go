// Package forgeselect is the clone-time screen where the user picks which
// forges each repo being cloned is pushed to, and which single one it is
// fetched from (the origin). It seeds the
// selection from an auto-match on the clone URL's host and lets the user add a
// new forge inline from that URL.
package forgeselect

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/forgepick"
)

type keyMap struct {
	Toggle, Origin, Scheme, Add, Confirm, Exit key.Binding
}

var keys = keyMap{
	Toggle:  key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "upstream")),
	Origin:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "origin")),
	Scheme:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scheme")),
	Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add forge")),
	Confirm: key.NewBinding(key.WithKeys("enter", "ctrl+p"), key.WithHelp("enter", "next")),
	Exit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.Toggle, keys.Origin, keys.Scheme, keys.Add, keys.Confirm, keys.Exit, constants.GlobalKeyMap.Quit}
}

type Model struct {
	list      list.Model
	forges    []config.Forge // working registry (base + inline-added)
	newForges []config.Forge // inline-added, persisted on completion

	pending []repomgr.PendingClone
	idx     int

	adding    bool
	nameInput textinput.Model
	err       string
}

// New builds the screen for a batch of pending clones.
func New(cfg config.Config, pending []repomgr.PendingClone) *Model {
	ti := textinput.New()
	ti.Placeholder = "forge name"
	ti.CharLimit = 40

	forges := slices.Clone(cfg.Forges)

	w, h := styles.ContentSize(0)
	l := list.New(nil, list.NewDefaultDelegate(), w, h)
	l.Help = styles.Help
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.KeyMap.Quit = constants.ListQuit
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{forges: forges, pending: pending, nameInput: ti, list: l}
	m.refresh()
	m.selectOrigin()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) cur() *repomgr.PendingClone { return &m.pending[m.idx] }

// refresh rebuilds the list items + title from the current pending selection,
// preserving the cursor position.
func (m *Model) refresh() {
	if len(m.pending) == 0 {
		return
	}
	p := m.cur()
	idx := m.list.Index()
	m.list.SetItems(forgepick.Items(m.forges, p.RepoLink))
	m.list.Select(idx)
	m.list.Title = fmt.Sprintf("Link Forges · repo %d/%d · %s · %s",
		m.idx+1, len(m.pending), p.URL.Key(), config.NormalizeScheme(p.Scheme))
}

func (m *Model) selectOrigin() {
	if i := forgepick.IndexOf(m.forges, m.cur().Origin); i >= 0 {
		m.list.Select(i)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		return m, func() tea.Msg { return events.ForgeSelectComplete{} }
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.resize()
	}

	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	if m.adding {
		return m.updateAdding(km)
	}

	switch {
	case key.Matches(km, constants.GlobalKeyMap.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Exit):
		return m, commands.SetState(domain.StateMain)
	case key.Matches(km, keys.Toggle):
		m.toggleSelected()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Origin):
		m.setOriginSelected()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Scheme):
		m.cur().ToggleScheme()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Add):
		return m.startAdd()
	case key.Matches(km, keys.Confirm):
		return m.confirm()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) selectedForge() (config.Forge, bool) {
	i := m.list.Index()
	if i < 0 || i >= len(m.forges) {
		return config.Forge{}, false
	}
	return m.forges[i], true
}

func (m *Model) toggleSelected() {
	f, ok := m.selectedForge()
	if !ok {
		return
	}
	m.cur().ToggleUpstream(f.Name)
	m.err = ""
}

// setOriginSelected makes the highlighted forge the single fetch source for the
// current repo, adding it to the upstreams if it wasn't already one.
func (m *Model) setOriginSelected() {
	f, ok := m.selectedForge()
	if !ok {
		return
	}
	m.cur().SetOrigin(f.Name)
	m.err = ""
}

func (m *Model) confirm() (tea.Model, tea.Cmd) {
	if m.cur().Origin == "" {
		m.err = "pick an origin forge (o) before continuing"
		return m, nil
	}
	if m.idx < len(m.pending)-1 {
		m.idx++
		m.err = ""
		m.refresh()
		m.selectOrigin()
		return m, nil
	}
	clones := m.pending
	newForges := m.newForges
	return m, func() tea.Msg {
		return events.ForgeSelectComplete{Clones: clones, NewForges: newForges}
	}
}

// startAdd begins adding a forge for the current URL's host. If that host is
// already registered, it just selects it instead of prompting.
func (m *Model) startAdd() (tea.Model, tea.Cmd) {
	host := m.cur().URL.Host
	for i, f := range m.forges {
		if strings.EqualFold(f.Host, host) {
			m.list.Select(i)
			m.setOriginSelected()
			m.refresh()
			return m, nil
		}
	}
	m.adding = true
	m.resize()
	m.nameInput.SetValue(suggestName(host))
	m.nameInput.CursorEnd()
	return m, m.nameInput.Focus()
}

func (m *Model) updateAdding(km tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(km, keys.Exit):
		m.adding = false
		m.resize()
		m.nameInput.Blur()
		return m, nil
	case key.Matches(km, keys.Confirm):
		name := strings.TrimSpace(m.nameInput.Value())
		host := m.cur().URL.Host
		if name == "" {
			m.err = "forge name can't be empty"
			return m, nil
		}
		if forgepick.IndexOf(m.forges, name) >= 0 {
			m.err = fmt.Sprintf("forge %q already exists", name)
			return m, nil
		}
		f := config.Forge{Name: name, Host: host}
		m.forges = append(m.forges, f)
		m.newForges = append(m.newForges, f)
		m.adding = false
		m.resize()
		m.nameInput.Blur()
		m.refresh()
		m.list.Select(len(m.forges) - 1)
		m.setOriginSelected()
		m.refresh()
		return m, nil
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(km)
	return m, cmd
}

func (m *Model) View() tea.View {
	if len(m.pending) == 0 {
		return tea.NewView("")
	}
	view := m.list.View()
	if !m.adding && m.err == "" {
		return tea.NewView(view)
	}

	rows := []string{view}
	if m.err != "" {
		rows = append(rows, styles.ToastErrorStyle.Render(m.err))
	}
	if m.adding {
		form := lipgloss.JoinVertical(lipgloss.Left,
			styles.Subtle("add forge name"),
			m.nameInput.View(),
			styles.Subtle("enter save · esc cancel"),
		)
		rows = append(rows, styles.FormBoxStyle.Render(form))
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// helpers

// addFormLines is the vertical space the framed add-forge form occupies (3
// content rows + rounded border + top margin), reserved from the list while
// adding.
const addFormLines = 7

// resize lays out the list, leaving room for the framed add form when it's open.
func (m *Model) resize() {
	reserved := 0
	if m.adding {
		reserved = addFormLines
	}
	m.list.SetSize(styles.ContentSize(reserved))
}

func suggestName(host string) string {
	if i := strings.IndexByte(host, '.'); i > 0 {
		return host[:i]
	}
	return host
}
