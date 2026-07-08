// Package forgeselect is the clone-time screen where the user picks which
// forges host each repo being cloned, and which one is primary. It seeds the
// selection from an auto-match on the clone URL's host and lets the user add a
// new forge inline from that URL.
package forgeselect

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Toggle, Primary, Scheme, Add, Confirm, Exit key.Binding
}

var keys = keyMap{
	Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Primary: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "primary")),
	Scheme:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scheme")),
	Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add forge")),
	Confirm: key.NewBinding(key.WithKeys("enter", "ctrl+p"), key.WithHelp("enter", "next")),
	Exit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.Toggle, keys.Primary, keys.Scheme, keys.Add, keys.Confirm, keys.Exit}
}

// forgeItem is a registry forge as shown in the selection list.
type forgeItem struct {
	name, host       string
	checked, primary bool
}

func (i forgeItem) Title() string {
	box := styles.GlyphCheckOff
	if i.checked {
		box = styles.GlyphCheckOn
	}
	t := box + " " + i.name
	if i.primary {
		t += " " + styles.GlyphPrimary
	}
	return t
}
func (i forgeItem) Description() string {
	if i.primary {
		return i.host + "  · primary"
	}
	return i.host
}
func (i forgeItem) FilterValue() string { return i.name }

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

	forges := make([]config.Forge, len(cfg.Forges))
	copy(forges, cfg.Forges)

	w, h := sizeBuffer()
	l := list.New(nil, list.NewDefaultDelegate(), w, h)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{forges: forges, pending: pending, nameInput: ti, list: l}
	m.refresh()
	m.selectPrimary()
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
	items := make([]list.Item, len(m.forges))
	for i, f := range m.forges {
		items[i] = forgeItem{name: f.Name, host: f.Host, checked: p.HasForge(f.Name), primary: p.Primary == f.Name}
	}
	idx := m.list.Index()
	m.list.SetItems(items)
	m.list.Select(idx)
	m.list.Title = fmt.Sprintf("Link Forges · repo %d/%d · %s · %s",
		m.idx+1, len(m.pending), p.URL.Key(), schemeLabel(p.Scheme))
}

func (m *Model) selectPrimary() {
	for i, f := range m.forges {
		if f.Name == m.cur().Primary {
			m.list.Select(i)
			return
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		return m, func() tea.Msg { return events.ForgeSelectComplete{} }
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		w, h := sizeBuffer()
		m.list.SetSize(w, h)
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	if m.adding {
		return m.updateAdding(km)
	}

	switch {
	case key.Matches(km, keys.Exit):
		return m, func() tea.Msg { return events.SetState{State: domain.StateMain} }
	case key.Matches(km, keys.Toggle):
		m.toggleSelected()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Primary):
		m.setPrimarySelected()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Scheme):
		if m.cur().Scheme == repomgr.SchemeSSH {
			m.cur().Scheme = repomgr.SchemeHTTPS
		} else {
			m.cur().Scheme = repomgr.SchemeSSH
		}
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
	p := m.cur()
	if p.HasForge(f.Name) {
		p.Forges = removeStr(p.Forges, f.Name)
		if p.Primary == f.Name {
			p.Primary = ""
			if len(p.Forges) > 0 {
				p.Primary = p.Forges[0]
			}
		}
	} else {
		p.Forges = append(p.Forges, f.Name)
		if p.Primary == "" {
			p.Primary = f.Name
		}
	}
	m.err = ""
}

func (m *Model) setPrimarySelected() {
	f, ok := m.selectedForge()
	if !ok {
		return
	}
	p := m.cur()
	if !p.HasForge(f.Name) {
		p.Forges = append(p.Forges, f.Name)
	}
	p.Primary = f.Name
	m.err = ""
}

func (m *Model) confirm() (tea.Model, tea.Cmd) {
	if m.cur().Primary == "" {
		m.err = "pick a primary forge (p) before continuing"
		return m, nil
	}
	if m.idx < len(m.pending)-1 {
		m.idx++
		m.err = ""
		m.refresh()
		m.selectPrimary()
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
			m.setPrimarySelected()
			m.refresh()
			return m, nil
		}
	}
	m.adding = true
	m.nameInput.SetValue(suggestName(host))
	m.nameInput.CursorEnd()
	return m, m.nameInput.Focus()
}

func (m *Model) updateAdding(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(km, keys.Exit):
		m.adding = false
		m.nameInput.Blur()
		return m, nil
	case key.Matches(km, keys.Confirm):
		name := strings.TrimSpace(m.nameInput.Value())
		host := m.cur().URL.Host
		if name == "" {
			m.err = "forge name can't be empty"
			return m, nil
		}
		if _, exists := forgeByName(m.forges, name); exists {
			m.err = fmt.Sprintf("forge %q already exists", name)
			return m, nil
		}
		f := config.Forge{Name: name, Host: host}
		m.forges = append(m.forges, f)
		m.newForges = append(m.newForges, f)
		m.adding = false
		m.nameInput.Blur()
		m.refresh()
		m.list.Select(len(m.forges) - 1)
		m.setPrimarySelected()
		m.refresh()
		return m, nil
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(km)
	return m, cmd
}

func (m *Model) View() string {
	if len(m.pending) == 0 {
		return ""
	}
	view := m.list.View()

	var extra []string
	if m.err != "" {
		extra = append(extra, styles.ToastErrorStyle.Render(m.err))
	}
	if m.adding {
		extra = append(extra, styles.Subtle("add forge name  ")+m.nameInput.View()+
			"   "+styles.Subtle("enter save · esc cancel"))
	}
	if len(extra) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, view, lipgloss.JoinVertical(lipgloss.Left, extra...))
	}
	return view
}

// helpers

func sizeBuffer() (w, h int) {
	x, y := styles.DocStyle.GetFrameSize()
	w = constants.WindowSize.Width - x
	h = constants.WindowSize.Height - y - constants.FooterReservedLines
	if h < 1 {
		h = 1
	}
	return
}

func schemeLabel(s string) string {
	if s == repomgr.SchemeSSH {
		return "ssh"
	}
	return "https"
}

func suggestName(host string) string {
	if i := strings.IndexByte(host, '.'); i > 0 {
		return host[:i]
	}
	return host
}

func removeStr(s []string, v string) []string {
	out := s[:0]
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

func forgeByName(forges []config.Forge, name string) (config.Forge, bool) {
	for _, f := range forges {
		if f.Name == name {
			return f, true
		}
	}
	return config.Forge{}, false
}
