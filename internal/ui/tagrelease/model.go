// Package tagrelease is the screen that picks a repo's next major, minor or
// patch version and, once confirmed, tags HEAD with it and pushes the tag.
package tagrelease

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/semver"
	"github.com/bkenks/lazymux/internal/styles"
)

type Model struct {
	form        *huh.Form
	title       string
	dir         string
	format      semver.Format
	latest      semver.Version
	bump        semver.Bump
	isConfirmed bool
}

// New builds the screen for the repo at dir, bumping the highest of its tags
// in format, or 0.0.0 when it has none.
func New(key, dir string, format semver.Format, tags []string) *Model {
	latest, hasLatest := format.Latest(tags)
	m := &Model{title: "Tag Version · " + key, dir: dir, format: format, latest: latest}

	current := "No " + format.Tag(semver.Version{}) + "-style tags yet; bumping from 0.0.0."
	if hasLatest {
		current = "Latest tag: " + format.Tag(latest)
	}
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewSelect[semver.Bump]().Title("Version to release").Description(current).
			Options(m.bumpOptions()...).Value(&m.bump),
		huh.NewConfirm().
			TitleFunc(func() string {
				return fmt.Sprintf("Tag HEAD as %s and push it to origin?", m.nextTag())
			}, &m.bump).
			Affirmative("Push").Negative("Cancel").
			Value(&m.isConfirmed).Validate(m.validateConfirmed),
	)).WithKeyMap(styles.FormKeyMap()).WithShowHelp(true).WithTheme(styles.FormTheme)
	m.resize()
	return m
}

func (m *Model) bumpOptions() []huh.Option[semver.Bump] {
	options := make([]huh.Option[semver.Bump], len(semver.Bumps))
	for i, bump := range semver.Bumps {
		label := fmt.Sprintf("%-5s  %s", bump, m.format.Tag(m.latest.Bump(bump)))
		options[i] = huh.NewOption(label, bump)
	}
	return options
}

func (m *Model) nextTag() string { return m.format.Tag(m.latest.Bump(m.bump)) }

// validateConfirmed stops a push of a tag git would reject, which a
// hand-edited tag prefix or suffix can produce.
func (m *Model) validateConfirmed(isConfirmed bool) error {
	if !isConfirmed {
		return nil
	}
	return repomgr.CheckTagName(m.nextTag())
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
		return m, commands.SetState(domain.StateMain)
	case huh.StateCompleted:
		if !m.isConfirmed {
			return m, commands.SetState(domain.StateMain)
		}
		tag := m.nextTag()
		pushing := events.Toast{Level: events.ToastInfo, Msg: "pushing " + tag + " to origin…"}
		return m, tea.Batch(
			commands.SetState(domain.StateMain),
			func() tea.Msg { return pushing },
			commands.ReleaseTagCmd(m.dir, tag),
		)
	}
	return m, cmd
}

func (m *Model) resize() { m.form = styles.FitForm(m.form, m.title) }

func (m *Model) View() tea.View { return tea.NewView(styles.RenderFormScreen(m.title, m.form)) }
