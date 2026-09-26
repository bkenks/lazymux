// Package reposettings is the per-repo settings screen: the prefix and suffix
// its version tags wrap around MAJOR.MINOR.PATCH, which forges it is pushed to
// (upstreams), which one it is fetched from (the origin), and the URL scheme.
// Submitting the form, or ctrl+s from any field, saves it and re-renders the
// repo's placeholder origin, insteadOf rule and push URLs; esc leaves without
// saving.
package reposettings

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/bkenks/gitkeeper/internal/commands"
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/domain"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/repomgr"
	"github.com/bkenks/gitkeeper/internal/semver"
	"github.com/bkenks/gitkeeper/internal/styles"
)

type Model struct {
	form    *huh.Form
	fields  []huh.Field
	title   string
	repoKey string
	link    config.RepoLink
	tags    []string
	tagsErr error
}

// New builds the screen for one repo. tags are its existing tags, for
// previewing the tag format, and tagsErr is why they couldn't be read. A repo
// with no scheme yet starts on the config default.
func New(cfg config.Config, repoKey string, tags []string, tagsErr error) *Model {
	link := cfg.Repos[repoKey].Clone()
	if link.Scheme == "" {
		link.Scheme = cfg.Behavior.DefaultProtocol
	}
	link.Scheme = config.NormalizeScheme(link.Scheme)

	m := &Model{
		title:   "Repo Settings · " + repoKey,
		repoKey: repoKey,
		link:    link,
		tags:    tags,
		tagsErr: tagsErr,
	}
	fields := []huh.Field{
		huh.NewInput().Title("Tag prefix").
			Description("Text before the version, e.g. v or mypkg/v.").
			Value(&m.link.TagPrefix).Validate(validateAffix),
		huh.NewInput().Title("Tag suffix").
			DescriptionFunc(m.describeTags, &m.link).
			Value(&m.link.TagSuffix).Validate(validateAffix),
	}
	m.fields = append(fields, m.forgeFields(cfg.Forges)...)
	m.form = huh.NewForm(huh.NewGroup(m.fields...)).
		WithKeyMap(styles.FormKeyMap()).WithShowHelp(false).WithTheme(styles.FormTheme)
	m.resize()
	return m
}

// forgeFields edit the repo's upstreams, origin and scheme, or explain how to
// add a forge when the registry is empty.
func (m *Model) forgeFields(forges []config.Forge) []huh.Field {
	if len(forges) == 0 {
		return []huh.Field{huh.NewNote().Title("Forges").
			Description("No forges registered. Press F on the repo list to add one.")}
	}
	upstreams := make([]huh.Option[string], len(forges))
	for i, forge := range forges {
		upstreams[i] = huh.NewOption(forge.Name+" ("+forge.Host+")", forge.Name)
	}
	const upstreamsHeaderRows = 2
	return []huh.Field{
		huh.NewMultiSelect[string]().Title("Upstreams").
			Description("Every forge a push goes to.").
			Options(upstreams...).Filterable(false).Value(&m.link.Upstreams).
			Height(upstreamsHeaderRows + len(upstreams)),
		huh.NewSelect[string]().Title("Origin").
			Description("The upstream fetch and pull read from.").
			OptionsFunc(m.originOptions, &m.link.Upstreams).Value(&m.link.Origin),
		huh.NewSelect[string]().Title("Scheme").Inline(true).
			Options(huh.NewOptions(config.SchemeHTTPS, config.SchemeSSH)...).
			Value(&m.link.Scheme),
	}
}

func (m *Model) originOptions() []huh.Option[string] {
	if len(m.link.Upstreams) == 0 {
		return []huh.Option[string]{huh.NewOption("none, pick an upstream first", "")}
	}
	return huh.NewOptions(m.link.Upstreams...)
}

// describeTags previews the tags the current prefix and suffix produce,
// starting from the repo's latest tag in that format.
func (m *Model) describeTags() string {
	format := tagFormat(m.link)
	if m.tagsErr != nil {
		return fmt.Sprintf("Tags look like %s. Couldn't read this repo's tags: %v",
			format.Tag(semver.Version{Major: 1, Minor: 2, Patch: 3}), m.tagsErr)
	}
	latest, hasLatest := format.Latest(m.tags)
	next := make([]string, len(semver.Bumps))
	for i, bump := range semver.Bumps {
		next[i] = format.Tag(latest.Bump(bump))
	}
	summary := "Next " + strings.Join(next, " / ")
	if !hasLatest {
		return "No tags in this format yet. " + summary
	}
	return "Latest " + format.Tag(latest) + ". " + summary
}

func tagFormat(link config.RepoLink) semver.Format {
	return semver.Format{
		Prefix: strings.TrimSpace(link.TagPrefix),
		Suffix: strings.TrimSpace(link.TagSuffix),
	}
}

// validateAffix rejects a prefix or suffix that can't be part of a git tag.
func validateAffix(affix string) error {
	return repomgr.CheckTagName(strings.TrimSpace(affix) + "0.0.0")
}

func (m *Model) Init() tea.Cmd { return m.form.Init() }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form.State != huh.StateNormal {
		return m, nil
	}
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.resize()
	}

	var cmd tea.Cmd
	m.form, cmd = styles.UpdateForm(m.form, m.fields, msg)

	switch m.form.State {
	case huh.StateAborted:
		return m, commands.SetState(domain.StateMain)
	case huh.StateCompleted:
		changed := events.RepoSettingsChanged{Key: m.repoKey, Link: m.edited()}
		return m, tea.Batch(
			func() tea.Msg { return changed },
			commands.SetState(domain.StateMain),
		)
	}
	return m, cmd
}

// edited is the link as the form left it, with the tag format trimmed and the
// origin kept to one of the upstreams.
func (m *Model) edited() config.RepoLink {
	link := m.link.Clone()
	format := tagFormat(link)
	link.TagPrefix, link.TagSuffix = format.Prefix, format.Suffix
	if !slices.Contains(link.Upstreams, link.Origin) {
		link.Origin = ""
		if len(link.Upstreams) > 0 {
			link.Origin = link.Upstreams[0]
		}
	}
	return link
}

func (m *Model) resize() { m.form = styles.FitForm(m.form, m.title) }

func (m *Model) View() tea.View {
	return tea.NewView(styles.RenderFormScreen(m.title, m.form, styles.SaveKey))
}
