// Package forgeregistry is the settings screen for managing the forge registry
// — the list of git hosts (name + host) that repos can be linked to.
package forgeregistry

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Add, Edit, Delete, Save, Field, Exit key.Binding
}

var keys = keyMap{
	Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:   key.NewBinding(key.WithKeys("e", "enter"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Save:   key.NewBinding(key.WithKeys("enter")),
	Field:  key.NewBinding(key.WithKeys("tab")),
	Exit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "save & back")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.Add, keys.Edit, keys.Delete, keys.Exit}
}

// forgeItem is a registry entry shown in the list.
type forgeItem struct {
	name, host string
	uses       int
}

func (i forgeItem) Title() string { return i.name }
func (i forgeItem) Description() string {
	if i.uses > 0 {
		return fmt.Sprintf("%s  · %d repo(s)", i.host, i.uses)
	}
	return i.host
}
func (i forgeItem) FilterValue() string { return i.name }

type Model struct {
	list   list.Model
	forges []config.Forge

	// repos is a working copy of the repo→forge links. Deleting or renaming a
	// forge cascades into it so no repo is left pointing at a missing forge;
	// it's emitted with ForgesChanged on exit.
	repos map[string]config.RepoLink

	// inUse maps a forge name to how many repos link it, shown per row.
	inUse map[string]int

	editing    bool
	editIdx    int // -1 = adding new
	nameInput  textinput.Model
	hostInput  textinput.Model
	nameActive bool
	err        string
}

func New(cfg config.Config) *Model {
	forges := make([]config.Forge, len(cfg.Forges))
	copy(forges, cfg.Forges)

	// Deep-copy the repo links so edits here don't mutate the app's config
	// until ForgesChanged is applied on exit.
	repos := make(map[string]config.RepoLink, len(cfg.Repos))
	for k, v := range cfg.Repos {
		repos[k] = config.RepoLink{
			Forges:  append([]string(nil), v.Forges...),
			Primary: v.Primary,
			Scheme:  v.Scheme,
		}
	}

	name := textinput.New()
	name.Placeholder = "name (e.g. github)"
	name.CharLimit = 40
	host := textinput.New()
	host.Placeholder = "host (e.g. github.com)"
	host.CharLimit = 100

	w, h := sizeBuffer()
	l := list.New(nil, list.NewDefaultDelegate(), w, h)
	l.Title = "Forge Registry"
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{list: l, forges: forges, repos: repos, editIdx: -1, nameInput: name, hostInput: host}
	m.recomputeUses()
	m.refresh()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

// recomputeUses rebuilds the per-forge repo counts from the working repo copy.
func (m *Model) recomputeUses() {
	uses := map[string]int{}
	for _, link := range m.repos {
		for _, f := range link.Forges {
			uses[f]++
		}
	}
	m.inUse = uses
}

func (m *Model) refresh() {
	items := make([]list.Item, len(m.forges))
	for i, f := range m.forges {
		items[i] = forgeItem{name: f.Name, host: f.Host, uses: m.inUse[f.Name]}
	}
	idx := m.list.Index()
	m.list.SetItems(items)
	if idx >= len(items) {
		idx = len(items) - 1
	}
	if idx >= 0 {
		m.list.Select(idx)
	}
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
	if m.editing {
		return m.updateEditing(km)
	}

	switch {
	case key.Matches(km, keys.Exit):
		forges := m.forges
		repos := m.repos
		return m, tea.Batch(
			func() tea.Msg { return events.ForgesChanged{Forges: forges, Repos: repos} },
			func() tea.Msg { return events.SetState{State: domain.StateMain} },
		)
	case key.Matches(km, keys.Add):
		return m.startEdit(-1)
	case key.Matches(km, keys.Edit):
		if len(m.forges) > 0 {
			return m.startEdit(m.list.Index())
		}
	case key.Matches(km, keys.Delete):
		if len(m.forges) > 0 {
			idx := m.list.Index()
			name := m.forges[idx].Name
			// Cascade the removal into every repo that links this forge:
			// drop it, and promote another linked forge to primary (or leave
			// the repo unlinked) when it was the primary.
			m.applyForgeRemoved(name)
			m.forges = append(m.forges[:idx], m.forges[idx+1:]...)
			m.recomputeUses()
			m.err = ""
			m.refresh()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) startEdit(idx int) (tea.Model, tea.Cmd) {
	m.editing = true
	m.editIdx = idx
	m.err = ""
	if idx >= 0 {
		m.nameInput.SetValue(m.forges[idx].Name)
		m.hostInput.SetValue(m.forges[idx].Host)
	} else {
		m.nameInput.SetValue("")
		m.hostInput.SetValue("")
	}
	m.nameActive = true
	m.hostInput.Blur()
	return m, m.nameInput.Focus()
}

func (m *Model) updateEditing(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(km, keys.Exit):
		m.editing = false
		m.nameInput.Blur()
		m.hostInput.Blur()
		return m, nil
	case key.Matches(km, keys.Field):
		m.nameActive = !m.nameActive
		if m.nameActive {
			m.hostInput.Blur()
			return m, m.nameInput.Focus()
		}
		m.nameInput.Blur()
		return m, m.hostInput.Focus()
	case key.Matches(km, keys.Save):
		return m.saveEdit()
	}

	var cmd tea.Cmd
	if m.nameActive {
		m.nameInput, cmd = m.nameInput.Update(km)
	} else {
		m.hostInput, cmd = m.hostInput.Update(km)
	}
	return m, cmd
}

func (m *Model) saveEdit() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	host := strings.TrimSpace(m.hostInput.Value())
	if name == "" || host == "" {
		m.err = "name and host are required"
		return m, nil
	}
	for i, f := range m.forges {
		if f.Name == name && i != m.editIdx {
			m.err = fmt.Sprintf("forge %q already exists", name)
			return m, nil
		}
	}
	f := config.Forge{Name: name, Host: host}
	if m.editIdx >= 0 {
		old := m.forges[m.editIdx].Name
		m.forges[m.editIdx] = f
		if old != name {
			// Cascade the rename into repo links so they keep referencing this
			// forge under its new name. A host-only edit needs no cascade (the
			// app re-renders affected repos from the new registry).
			m.applyForgeRenamed(old, name)
			m.recomputeUses()
		}
	} else {
		m.forges = append(m.forges, f)
	}
	m.editing = false
	m.nameInput.Blur()
	m.hostInput.Blur()
	m.refresh()
	if m.editIdx < 0 {
		m.list.Select(len(m.forges) - 1)
	}
	return m, nil
}

// applyForgeRemoved drops forge `name` from every repo link. When it was a
// repo's primary, the first remaining linked forge is promoted; if none remain
// the repo is left unlinked (empty primary).
func (m *Model) applyForgeRemoved(name string) {
	for key, link := range m.repos {
		if !contains(link.Forges, name) {
			continue
		}
		link.Forges = without(link.Forges, name)
		if link.Primary == name {
			link.Primary = ""
			if len(link.Forges) > 0 {
				link.Primary = link.Forges[0]
			}
		}
		m.repos[key] = link
	}
}

// applyForgeRenamed rewrites forge `old` to `new` in every repo link (and their
// primary), so a rename never orphans a repo.
func (m *Model) applyForgeRenamed(old, newName string) {
	for key, link := range m.repos {
		if !contains(link.Forges, old) {
			continue
		}
		forges := make([]string, 0, len(link.Forges))
		for _, f := range link.Forges {
			if f == old {
				f = newName
			}
			if !contains(forges, f) {
				forges = append(forges, f)
			}
		}
		link.Forges = forges
		if link.Primary == old {
			link.Primary = newName
		}
		m.repos[key] = link
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func without(s []string, v string) []string {
	out := make([]string, 0, len(s))
	for _, x := range s {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

func (m *Model) View() string {
	view := m.list.View()

	var extra []string
	if m.err != "" {
		extra = append(extra, styles.ToastErrorStyle.Render(m.err))
	}
	if m.editing {
		verb := "edit forge"
		if m.editIdx < 0 {
			verb = "new forge"
		}
		extra = append(extra,
			styles.Subtle(verb),
			styles.Subtle("name  ")+m.nameInput.View(),
			styles.Subtle("host  ")+m.hostInput.View(),
			styles.Subtle("tab switch field · enter save · esc cancel"),
		)
	}
	if len(extra) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, view, lipgloss.JoinVertical(lipgloss.Left, extra...))
	}
	return view
}

func sizeBuffer() (w, h int) {
	x, y := styles.DocStyle.GetFrameSize()
	w = constants.WindowSize.Width - x
	h = constants.WindowSize.Height - y - constants.FooterReservedLines
	if h < 1 {
		h = 1
	}
	return
}
