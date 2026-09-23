// Package repoforges is the per-repo screen for changing which forges a repo is
// pushed to, which single one it is fetched from (the origin), and the URL
// scheme. Saving re-renders the repo's placeholder origin, insteadOf rule, and
// push URLs.
package repoforges

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/forgepick"
)

type keyMap struct {
	Toggle, Origin, Scheme, Exit key.Binding
}

var keys = keyMap{
	Toggle: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "upstream")),
	Origin: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "origin")),
	Scheme: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "scheme")),
	Exit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "save & back")),
}

func helpKeys() []key.Binding {
	return []key.Binding{keys.Toggle, keys.Origin, keys.Scheme, keys.Exit, constants.GlobalKeyMap.Quit}
}

type Model struct {
	list    list.Model
	repoKey string
	forges  []config.Forge
	link    config.RepoLink
}

// New builds the screen for one repo. If the repo has no scheme yet it defaults
// to the config default.
func New(cfg config.Config, repoKey string) *Model {
	forges := slices.Clone(cfg.Forges)

	link := cfg.Repos[repoKey].Clone()
	if link.Scheme == "" {
		link.Scheme = cfg.Behavior.DefaultProtocol
	}

	w, h := styles.ContentSize(0)
	l := list.New(nil, list.NewDefaultDelegate(), w, h)
	l.SetFilteringEnabled(false)
	l.KeyMap.Quit = constants.ListQuit
	l.AdditionalShortHelpKeys = helpKeys
	l.AdditionalFullHelpKeys = helpKeys

	m := &Model{list: l, repoKey: repoKey, forges: forges, link: link}
	m.refresh()
	if i := forgepick.IndexOf(forges, link.Origin); i >= 0 {
		m.list.Select(i)
	}
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) refresh() {
	idx := m.list.Index()
	m.list.SetItems(forgepick.Items(m.forges, m.link))
	m.list.Select(idx)
	m.list.Title = fmt.Sprintf("Repo Forges · %s · %s", m.repoKey, config.NormalizeScheme(m.link.Scheme))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.list.SetSize(styles.ContentSize(0))
	}

	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(km, constants.GlobalKeyMap.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Exit):
		changed := events.RepoLinkChanged{Key: m.repoKey, Link: m.link.Clone()}
		return m, tea.Batch(
			func() tea.Msg { return changed },
			commands.SetState(domain.StateMain),
		)
	case key.Matches(km, keys.Toggle):
		if f, ok := m.selectedForge(); ok {
			m.link.ToggleUpstream(f.Name)
		}
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Origin):
		if f, ok := m.selectedForge(); ok {
			m.link.SetOrigin(f.Name)
		}
		m.refresh()
		return m, nil
	case key.Matches(km, keys.Scheme):
		m.link.ToggleScheme()
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

func (m *Model) View() tea.View { return tea.NewView(m.list.View()) }
