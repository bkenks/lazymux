// Package reposdir is the screen that makes the user choose where repos live.
// The app shows it in place of the repo list whenever the repo directory is
// unset or missing, and it can only be left by choosing one or quitting.
package reposdir

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
)

const title = "Repo directory"

type Model struct {
	form *huh.Form
	dir  string
}

// New builds the prompt, explaining problem and pre-filling the current repo
// directory so a missing one can be created by just confirming it.
func New(cfg config.Config, problem error) *Model {
	m := &Model{dir: cfg.RepoRoot()}
	formKeys := huh.NewDefaultKeyMap()
	formKeys.Quit = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Where should lazymux keep your repos?").
			Description(describe(problem)).
			Placeholder("~/Development").
			Value(&m.dir).Validate(validate),
	)).WithKeyMap(formKeys).WithShowHelp(false).WithTheme(styles.FormTheme)
	m.resize()
	return m
}

func describe(problem error) string {
	return problem.Error() + ". Repos are cloned into <directory>/<namespace>/<repo>; " +
		"a directory that doesn't exist yet is created."
}

func validate(input string) error {
	_, err := config.ParseReposDir(input)
	return err
}

func (m *Model) Init() tea.Cmd { return m.form.Init() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form.State != huh.StateNormal {
		return m, nil
	}
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.resize()
	}

	model, cmd := m.form.Update(msg)
	if form, ok := model.(*huh.Form); ok {
		m.form = form
	}

	switch m.form.State {
	case huh.StateAborted:
		return m, tea.Quit
	case huh.StateCompleted:
		dir := strings.TrimSpace(m.dir)
		return m, func() tea.Msg { return events.ReposDirChosen{Dir: dir} }
	}
	return m, cmd
}

func (m *Model) resize() { m.form = styles.FitForm(m.form, title) }

func (m *Model) View() tea.View { return tea.NewView(styles.RenderFormScreen(title, m.form)) }
