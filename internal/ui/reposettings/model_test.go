package reposettings

import (
	"errors"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/ui/formtest"
)

var (
	enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}
	spaceKey = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	escKey   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

func press(m *Model, msg tea.Msg) []tea.Msg {
	return formtest.Press(m, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case events.RepoSettingsChanged, events.SetState:
			return true
		}
		return false
	})
}

func linkedConfig() config.Config {
	cfg := config.Default()
	cfg.Forges = []config.Forge{
		{Name: "github", Host: "github.com"},
		{Name: "forgejo", Host: "fj.example.com"},
	}
	cfg.Repos["me/demo"] = config.RepoLink{
		Upstreams: []string{"github", "forgejo"}, Origin: "github", Scheme: config.SchemeSSH,
	}
	return cfg
}

func newTestModel(cfg config.Config, tags []string) *Model {
	m := New(cfg, "me/demo", tags, nil)
	press(m, m.Init())
	return m
}

// submit presses enter until the form emits, returning the saved link.
func submit(t *testing.T, m *Model) config.RepoLink {
	t.Helper()
	for range 10 {
		for _, msg := range press(m, enterKey) {
			if changed, ok := msg.(events.RepoSettingsChanged); ok {
				return changed.Link
			}
		}
	}
	t.Fatal("form never submitted")
	return config.RepoLink{}
}

func TestSubmitSavesTagFormatAndKeepsForgeLinks(t *testing.T) {
	m := newTestModel(linkedConfig(), nil)
	formtest.Type(m, " mypkg/v ")
	press(m, enterKey)
	formtest.Type(m, "-x")

	link := submit(t, m)

	if link.TagPrefix != "mypkg/v" || link.TagSuffix != "-x" {
		t.Errorf("tag format = %q/%q, want trimmed mypkg/v and -x", link.TagPrefix, link.TagSuffix)
	}
	if !slices.Equal(link.Upstreams, []string{"github", "forgejo"}) || link.Origin != "github" ||
		link.Scheme != config.SchemeSSH {
		t.Errorf("forge links = %+v, want them unchanged", link)
	}
}

func TestUncheckingOriginMovesOriginToRemainingUpstream(t *testing.T) {
	m := newTestModel(linkedConfig(), nil)
	press(m, enterKey)
	press(m, enterKey)
	press(m, spaceKey)

	link := submit(t, m)

	if !slices.Equal(link.Upstreams, []string{"forgejo"}) || link.Origin != "forgejo" {
		t.Errorf("link = %+v, want forgejo as the only upstream and origin", link)
	}
}

func TestEditedKeepsOriginAmongUpstreams(t *testing.T) {
	cases := []struct {
		upstreams  []string
		origin     string
		wantOrigin string
	}{
		{[]string{"forgejo"}, "github", "forgejo"},
		{nil, "github", ""},
		{[]string{"github", "forgejo"}, "forgejo", "forgejo"},
	}
	for _, c := range cases {
		m := New(linkedConfig(), "me/demo", nil, nil)
		m.link.Upstreams, m.link.Origin = c.upstreams, c.origin
		if got := m.edited().Origin; got != c.wantOrigin {
			t.Errorf("upstreams %v, origin %q: edited origin = %q, want %q",
				c.upstreams, c.origin, got, c.wantOrigin)
		}
	}
}

func TestInvalidPrefixBlocksSubmit(t *testing.T) {
	m := newTestModel(linkedConfig(), nil)
	formtest.Type(m, "my pkg/")

	for range 10 {
		for _, msg := range press(m, enterKey) {
			if _, ok := msg.(events.RepoSettingsChanged); ok {
				t.Fatal("saved a prefix git can't use in a tag")
			}
		}
	}
}

func TestEmptyRegistryStillSavesTagFormat(t *testing.T) {
	m := newTestModel(config.Default(), nil)
	formtest.Type(m, "v")

	if link := submit(t, m); link.TagPrefix != "v" || link.Origin != "" {
		t.Errorf("link = %+v, want only the prefix set", link)
	}
}

func TestEscLeavesWithoutSaving(t *testing.T) {
	m := newTestModel(linkedConfig(), nil)
	formtest.Type(m, "v")

	emitted := press(m, escKey)

	if len(emitted) != 1 {
		t.Fatalf("emitted %v, want one SetState", emitted)
	}
	if state, ok := emitted[0].(events.SetState); !ok || state.State != domain.StateMain {
		t.Errorf("emitted %v, want SetState to the repo list", emitted[0])
	}
}

func TestDescribeTagsPreviewsNextVersions(t *testing.T) {
	cases := []struct {
		prefix string
		tags   []string
		err    error
		want   string
	}{
		{"v", []string{"v1.2.3", "v1.0.0"}, nil, "Latest v1.2.3. Next v1.2.4 / v1.3.0 / v2.0.0"},
		{"v", []string{"1.2.3"}, nil, "No tags in this format yet. Next v0.0.1 / v0.1.0 / v1.0.0"},
		{"", nil, errors.New("boom"), "Tags look like 1.2.3. Couldn't read this repo's tags: boom"},
	}
	for _, c := range cases {
		m := New(config.Default(), "me/demo", c.tags, c.err)
		m.link.TagPrefix = c.prefix
		if got := m.describeTags(); got != c.want {
			t.Errorf("describeTags = %q, want %q", got, c.want)
		}
	}
}

func TestSaveKeySavesFromAnyField(t *testing.T) {
	m := newTestModel(linkedConfig(), nil)
	formtest.Type(m, "v")

	emitted := press(m, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})

	if len(emitted) == 0 {
		t.Fatal("ctrl+s did not save")
	}
	changed, ok := emitted[0].(events.RepoSettingsChanged)
	if !ok || changed.Link.TagPrefix != "v" || changed.Link.Origin != "github" {
		t.Errorf("emitted %v, want RepoSettingsChanged with prefix v and origin github", emitted)
	}
}
