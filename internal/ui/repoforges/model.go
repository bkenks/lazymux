// Package repoforges is the per-repo screen for changing which forges a repo
// is linked to, which one is primary, and the URL scheme. Saving re-renders the
// repo's placeholder origin + insteadOf rule.
package repoforges

import (
	"fmt"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type keyMap struct {
	Toggle, Primary, Scheme, Exit key.Binding
}

var keys = keyMap{
	Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
	Primary: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "primary")),
	Scheme:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scheme")),
	Exit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "save & back")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.Toggle, keys.Primary, keys.Scheme, keys.Exit}
}

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
	list    list.Model
	repoKey string
	forges  []config.Forge
	link    config.RepoLink
}

// New builds the screen for one repo. If the repo has no scheme yet it defaults
// to the config default.
func New(cfg config.Config, repoKey string) *Model {
	forges := make([]config.Forge, len(cfg.Forges))
	copy(forges, cfg.Forges)

	link := cfg.Repos[repoKey]
	if link.Scheme == "" {
		link.Scheme = cfg.Behavior.DefaultProtocol
	}

	w, h := sizeBuffer()
	l := list.New(nil, list.NewDefaultDelegate(), w, h)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{list: l, repoKey: repoKey, forges: forges, link: link}
	m.refresh()
	for i, f := range forges {
		if f.Name == link.Primary {
			m.list.Select(i)
		}
	}
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) has(name string) bool {
	for _, f := range m.link.Forges {
		if f == name {
			return true
		}
	}
	return false
}

func (m *Model) refresh() {
	items := make([]list.Item, len(m.forges))
	for i, f := range m.forges {
		items[i] = forgeItem{name: f.Name, host: f.Host, checked: m.has(f.Name), primary: m.link.Primary == f.Name}
	}
	idx := m.list.Index()
	m.list.SetItems(items)
	m.list.Select(idx)
	m.list.Title = fmt.Sprintf("Repo Forges · %s · %s", m.repoKey, schemeLabel(m.link.Scheme))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	switch {
	case key.Matches(km, keys.Exit):
		k, link := m.repoKey, m.link
		return m, tea.Batch(
			func() tea.Msg { return events.RepoLinkChanged{Key: k, Link: link} },
			func() tea.Msg { return events.SetState{State: domain.StateMain} },
		)
	case key.Matches(km, keys.Toggle):
		m.toggle()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Primary):
		m.setPrimary()
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Scheme):
		if m.link.Scheme == repomgr.SchemeSSH {
			m.link.Scheme = repomgr.SchemeHTTPS
		} else {
			m.link.Scheme = repomgr.SchemeSSH
		}
		m.refresh()
		return m, nil
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

func (m *Model) toggle() {
	f, ok := m.selectedForge()
	if !ok {
		return
	}
	if m.has(f.Name) {
		out := m.link.Forges[:0]
		for _, x := range m.link.Forges {
			if x != f.Name {
				out = append(out, x)
			}
		}
		m.link.Forges = out
		if m.link.Primary == f.Name {
			m.link.Primary = ""
			if len(m.link.Forges) > 0 {
				m.link.Primary = m.link.Forges[0]
			}
		}
	} else {
		m.link.Forges = append(m.link.Forges, f.Name)
		if m.link.Primary == "" {
			m.link.Primary = f.Name
		}
	}
}

func (m *Model) setPrimary() {
	f, ok := m.selectedForge()
	if !ok {
		return
	}
	if !m.has(f.Name) {
		m.link.Forges = append(m.link.Forges, f.Name)
	}
	m.link.Primary = f.Name
}

func (m *Model) View() string { return m.list.View() }

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
