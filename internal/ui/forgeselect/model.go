// Package forgeselect is the clone-time screen where the user picks which
// forges host each repo being cloned, and which one is primary. It seeds the
// selection from an auto-match on the clone URL's host and lets the user add a
// new forge inline from that URL.
package forgeselect

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up, Down, Toggle, Primary, Scheme, Add, Confirm, Exit key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Toggle:  key.NewBinding(key.WithKeys(" ")),
	Primary: key.NewBinding(key.WithKeys("p")),
	Scheme:  key.NewBinding(key.WithKeys("s")),
	Add:     key.NewBinding(key.WithKeys("a")),
	Confirm: key.NewBinding(key.WithKeys("enter", "ctrl+p")),
	Exit:    key.NewBinding(key.WithKeys("esc")),
}

type Model struct {
	placeholderHost string
	forges          []config.Forge // working registry (base + inline-added)
	newForges       []config.Forge // inline-added, persisted on completion

	pending []repomgr.PendingClone
	idx     int
	cursor  int

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

	m := &Model{
		placeholderHost: cfg.PlaceholderHost,
		forges:          forges,
		pending:         pending,
		nameInput:       ti,
	}
	m.syncCursorToPrimary()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) cur() *repomgr.PendingClone { return &m.pending[m.idx] }

// syncCursorToPrimary parks the cursor on the current pending's primary forge.
func (m *Model) syncCursorToPrimary() {
	if len(m.pending) == 0 {
		return
	}
	for i, f := range m.forges {
		if f.Name == m.cur().Primary {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		return m, func() tea.Msg { return events.ForgeSelectComplete{} }
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.adding {
		return m.updateAdding(km)
	}

	switch {
	case key.Matches(km, keys.Exit):
		return m, func() tea.Msg { return events.SetState{State: domain.StateMain} }

	case key.Matches(km, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(km, keys.Down):
		if m.cursor < len(m.forges)-1 {
			m.cursor++
		}

	case key.Matches(km, keys.Toggle):
		m.toggleCursor()

	case key.Matches(km, keys.Primary):
		m.setPrimaryCursor()

	case key.Matches(km, keys.Scheme):
		if m.cur().Scheme == repomgr.SchemeSSH {
			m.cur().Scheme = repomgr.SchemeHTTPS
		} else {
			m.cur().Scheme = repomgr.SchemeSSH
		}

	case key.Matches(km, keys.Add):
		return m.startAdd()

	case key.Matches(km, keys.Confirm):
		return m.confirm()
	}
	return m, nil
}

func (m *Model) toggleCursor() {
	if m.cursor >= len(m.forges) {
		return
	}
	name := m.forges[m.cursor].Name
	p := m.cur()
	if p.HasForge(name) {
		p.Forges = removeStr(p.Forges, name)
		if p.Primary == name {
			p.Primary = ""
			if len(p.Forges) > 0 {
				p.Primary = p.Forges[0]
			}
		}
	} else {
		p.Forges = append(p.Forges, name)
		if p.Primary == "" {
			p.Primary = name
		}
	}
	m.err = ""
}

func (m *Model) setPrimaryCursor() {
	if m.cursor >= len(m.forges) {
		return
	}
	name := m.forges[m.cursor].Name
	p := m.cur()
	if !p.HasForge(name) {
		p.Forges = append(p.Forges, name)
	}
	p.Primary = name
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
		m.syncCursorToPrimary()
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
			m.cursor = i
			m.setPrimaryCursor()
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
		m.cursor = len(m.forges) - 1
		m.setPrimaryCursor()
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
	p := m.cur()

	header := lipgloss.JoinVertical(lipgloss.Left,
		styles.MenuTitle.Render("Link Forges"),
		styles.MenuSubStyle.Render(fmt.Sprintf("repo %d/%d — %s", m.idx+1, len(m.pending), p.URL.Key())),
		styles.MenuSubStyle.Render("from "+p.RealURL),
		styles.MenuSubStyle.Render("scheme: "+schemeLabel(p.Scheme)),
	)

	var rows []string
	if len(m.forges) == 0 {
		rows = append(rows, styles.MenuSubStyle.Render("  no forges yet — press 'a' to add one from this URL"))
	}
	for i, f := range m.forges {
		check := "[ ]"
		if p.HasForge(f.Name) {
			check = "[x]"
		}
		star := " "
		if p.Primary == f.Name {
			star = "★"
		}
		line := fmt.Sprintf(" %s %s %s (%s)", star, check, f.Name, f.Host)
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Render(line + "  ◄")
		}
		rows = append(rows, line)
	}
	body := strings.Join(rows, "\n")

	var footer string
	if m.adding {
		footer = "add forge name: " + m.nameInput.View() + "\n" +
			styles.MenuHelpStyle.Render("enter confirm • esc cancel")
	} else {
		footer = styles.MenuHelpStyle.Render(
			"↑/↓ move • space toggle • p primary • s scheme • a add-forge • enter next • esc cancel")
	}
	if m.err != "" {
		footer = styles.ToastErrorStyle.Render(m.err) + "\n" + footer
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}

// helpers

func schemeLabel(s string) string {
	if s == repomgr.SchemeSSH {
		return "ssh (git@…)"
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
