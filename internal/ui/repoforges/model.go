// Package repoforges is the per-repo screen for changing which forges a repo
// is linked to, which one is primary, and the URL scheme. Saving re-renders the
// repo's placeholder origin + insteadOf rule.
package repoforges

import (
	"fmt"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up, Down, Toggle, Primary, Scheme, Exit key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Toggle:  key.NewBinding(key.WithKeys(" ")),
	Primary: key.NewBinding(key.WithKeys("p")),
	Scheme:  key.NewBinding(key.WithKeys("s")),
	Exit:    key.NewBinding(key.WithKeys("esc")),
}

type Model struct {
	repoKey string
	forges  []config.Forge
	link    config.RepoLink
	cursor  int
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
	m := &Model{repoKey: repoKey, forges: forges, link: link}
	for i, f := range forges {
		if f.Name == link.Primary {
			m.cursor = i
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

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(km, keys.Exit):
		k, link := m.repoKey, m.link
		return m, tea.Batch(
			func() tea.Msg { return events.RepoLinkChanged{Key: k, Link: link} },
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
	case key.Matches(km, keys.Toggle):
		m.toggle()
	case key.Matches(km, keys.Primary):
		m.setPrimary()
	case key.Matches(km, keys.Scheme):
		if m.link.Scheme == repomgr.SchemeSSH {
			m.link.Scheme = repomgr.SchemeHTTPS
		} else {
			m.link.Scheme = repomgr.SchemeSSH
		}
	}
	return m, nil
}

func (m *Model) toggle() {
	if m.cursor >= len(m.forges) {
		return
	}
	name := m.forges[m.cursor].Name
	if m.has(name) {
		out := m.link.Forges[:0]
		for _, x := range m.link.Forges {
			if x != name {
				out = append(out, x)
			}
		}
		m.link.Forges = out
		if m.link.Primary == name {
			m.link.Primary = ""
			if len(m.link.Forges) > 0 {
				m.link.Primary = m.link.Forges[0]
			}
		}
	} else {
		m.link.Forges = append(m.link.Forges, name)
		if m.link.Primary == "" {
			m.link.Primary = name
		}
	}
}

func (m *Model) setPrimary() {
	if m.cursor >= len(m.forges) {
		return
	}
	name := m.forges[m.cursor].Name
	if !m.has(name) {
		m.link.Forges = append(m.link.Forges, name)
	}
	m.link.Primary = name
}

func (m *Model) View() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		styles.MenuTitle.Render("Repo Forges"),
		styles.MenuSubStyle.Render(m.repoKey),
		styles.MenuSubStyle.Render("scheme: "+schemeLabel(m.link.Scheme)),
	)

	var rows []string
	if len(m.forges) == 0 {
		rows = append(rows, styles.MenuSubStyle.Render("  no forges in registry — add some in settings"))
	}
	for i, f := range m.forges {
		check := "[ ]"
		if m.has(f.Name) {
			check = "[x]"
		}
		star := " "
		if m.link.Primary == f.Name {
			star = "★"
		}
		line := fmt.Sprintf(" %s %s %s (%s)", star, check, f.Name, f.Host)
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Render(line + "  ◄")
		}
		rows = append(rows, line)
	}
	body := strings.Join(rows, "\n")

	footer := styles.MenuHelpStyle.Render(
		"↑/↓ move • space toggle • p primary • s scheme • esc save & back")

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
}

func schemeLabel(s string) string {
	if s == repomgr.SchemeSSH {
		return "ssh (git@…)"
	}
	return "https"
}
