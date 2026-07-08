// Package forgeregistry is the settings screen for managing the forge registry
// — the list of git hosts (name + host) that repos can be linked to.
package forgeregistry

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up, Down, Add, Edit, Delete, Save, Field, Exit key.Binding
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k")),
	Down:   key.NewBinding(key.WithKeys("down", "j")),
	Add:    key.NewBinding(key.WithKeys("a")),
	Edit:   key.NewBinding(key.WithKeys("e", "enter")),
	Delete: key.NewBinding(key.WithKeys("d")),
	Save:   key.NewBinding(key.WithKeys("enter")),
	Field:  key.NewBinding(key.WithKeys("tab")),
	Exit:   key.NewBinding(key.WithKeys("esc")),
}

type Model struct {
	forges []config.Forge
	cursor int

	// inUse maps a forge name to how many repos link it, so deletion of a
	// forge still referenced by repos can be blocked.
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

	inUse := map[string]int{}
	for _, link := range cfg.Repos {
		for _, f := range link.Forges {
			inUse[f]++
		}
	}

	name := textinput.New()
	name.Placeholder = "name (e.g. github)"
	name.CharLimit = 40
	host := textinput.New()
	host.Placeholder = "host (e.g. github.com)"
	host.CharLimit = 100

	return &Model{forges: forges, inUse: inUse, editIdx: -1, nameInput: name, hostInput: host}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.editing {
		return m.updateEditing(km)
	}

	switch {
	case key.Matches(km, keys.Exit):
		forges := m.forges
		return m, tea.Batch(
			func() tea.Msg { return events.ForgesChanged{Forges: forges} },
			func() tea.Msg { return events.SetState{State: domain.StateMain} },
		)
	case key.Matches(km, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(km, keys.Down):
		if m.cursor < len(m.forges)-1 {
			m.cursor++
		}
	case key.Matches(km, keys.Add):
		return m.startEdit(-1)
	case key.Matches(km, keys.Edit):
		if len(m.forges) > 0 {
			return m.startEdit(m.cursor)
		}
	case key.Matches(km, keys.Delete):
		if len(m.forges) > 0 {
			name := m.forges[m.cursor].Name
			if n := m.inUse[name]; n > 0 {
				m.err = fmt.Sprintf("%q is linked by %d repo(s) — repoint them (f) first", name, n)
				break
			}
			m.forges = append(m.forges[:m.cursor], m.forges[m.cursor+1:]...)
			if m.cursor >= len(m.forges) && m.cursor > 0 {
				m.cursor--
			}
			m.err = ""
		}
	}
	return m, nil
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
	// Enforce unique names (ignoring the row being edited).
	for i, f := range m.forges {
		if f.Name == name && i != m.editIdx {
			m.err = fmt.Sprintf("forge %q already exists", name)
			return m, nil
		}
	}
	f := config.Forge{Name: name, Host: host}
	if m.editIdx >= 0 {
		m.forges[m.editIdx] = f
	} else {
		m.forges = append(m.forges, f)
		m.cursor = len(m.forges) - 1
	}
	m.editing = false
	m.nameInput.Blur()
	m.hostInput.Blur()
	return m, nil
}

func (m *Model) View() string {
	header := styles.MenuTitle.Render("Forge Registry")

	nameW := 6
	for _, f := range m.forges {
		if n := len([]rune(f.Name)); n > nameW {
			nameW = n
		}
	}

	var rows []string
	if len(m.forges) == 0 {
		rows = append(rows, styles.Subtle("   no forges yet — press ")+styles.Accent("a")+styles.Subtle(" to add one"))
	}
	for i, f := range m.forges {
		active := i == m.cursor && !m.editing
		nameCol := fmt.Sprintf("%-*s", nameW, f.Name)
		if active {
			nameCol = styles.Strong(nameCol)
		}
		line := " " + styles.Cursor(active) + " " + nameCol + "  " + styles.Subtle("("+f.Host+")")
		if n := m.inUse[f.Name]; n > 0 {
			line += "  " + styles.Subtle(fmt.Sprintf("· %d repo(s)", n))
		}
		rows = append(rows, line)
	}
	body := lipgloss.NewStyle().MarginLeft(2).Render(strings.Join(rows, "\n"))

	var footer string
	if m.editing {
		verb := "edit forge"
		if m.editIdx < 0 {
			verb = "new forge"
		}
		footer = lipgloss.JoinVertical(lipgloss.Left,
			styles.Subtle(verb),
			styles.Subtle("name  ")+m.nameInput.View(),
			styles.Subtle("host  ")+m.hostInput.View(),
			styles.MenuHelpStyle.Render(styles.Help("tab", "switch field", "enter", "save", "esc", "cancel")),
		)
	} else {
		footer = styles.MenuHelpStyle.Render(styles.Help(
			"↑/↓", "move", "a", "add", "e", "edit", "d", "delete", "esc", "save & back"))
	}
	if m.err != "" {
		footer = styles.ToastErrorStyle.Render(m.err) + "\n" + footer
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}
