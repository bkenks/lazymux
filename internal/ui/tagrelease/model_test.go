package tagrelease

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/semver"
	"github.com/bkenks/lazymux/internal/ui/formtest"
)

var (
	enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}
	downKey  = tea.KeyPressMsg{Code: tea.KeyDown}
	yesKey   = tea.KeyPressMsg{Code: 'y', Text: "y"}
	noKey    = tea.KeyPressMsg{Code: 'n', Text: "n"}
	escKey   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

// press sends msg and returns what the screen emits for the app. The release
// command itself is left unrun, since it would tag and push a real repo.
func press(m *Model, msg tea.Msg) []tea.Msg {
	return formtest.Press(m, msg, func(msg tea.Msg) bool {
		switch msg.(type) {
		case events.SetState, events.Toast, events.TagReleased:
			return true
		}
		return false
	})
}

func newTestModel(format semver.Format, tags []string) *Model {
	m := New("me/demo", "/nonexistent/me/demo", format, tags)
	press(m, m.Init())
	return m
}

// pushingToast returns the text of the "pushing" toast among emitted, or ""
// when nothing is being pushed.
func pushingToast(t *testing.T, emitted []tea.Msg) string {
	t.Helper()
	for _, msg := range emitted {
		if toast, ok := msg.(events.Toast); ok && toast.Level == events.ToastInfo {
			return toast.Msg
		}
	}
	return ""
}

func TestConfirmedBumpPushesNextTag(t *testing.T) {
	cases := []struct {
		downs int
		want  string
	}{
		{0, "pushing v1.2.4 to origin…"},
		{1, "pushing v1.3.0 to origin…"},
		{2, "pushing v2.0.0 to origin…"},
	}
	for _, c := range cases {
		m := newTestModel(semver.Format{Prefix: "v"}, []string{"v1.2.3", "v1.0.0", "other"})
		for range c.downs {
			press(m, downKey)
		}
		press(m, enterKey)

		emitted := press(m, yesKey)

		if got := pushingToast(t, emitted); got != c.want {
			t.Errorf("%d downs: toast = %q, want %q", c.downs, got, c.want)
		}
		if !leavesToRepoList(emitted) {
			t.Errorf("%d downs: emitted %v, want a return to the repo list", c.downs, emitted)
		}
	}
}

func TestFirstReleaseBumpsFromZero(t *testing.T) {
	m := newTestModel(semver.Format{Prefix: "mypkg/v", Suffix: "-x"}, []string{"v9.9.9"})
	press(m, downKey)
	press(m, enterKey)

	if got := pushingToast(t, press(m, yesKey)); got != "pushing mypkg/v0.1.0-x to origin…" {
		t.Errorf("toast = %q, want the first minor release", got)
	}
}

func TestDecliningPushesNothing(t *testing.T) {
	m := newTestModel(semver.Format{Prefix: "v"}, nil)
	press(m, enterKey)

	emitted := press(m, noKey)

	if got := pushingToast(t, emitted); got != "" {
		t.Errorf("toast = %q, want nothing pushed", got)
	}
	if !leavesToRepoList(emitted) {
		t.Errorf("emitted %v, want a return to the repo list", emitted)
	}
}

func TestEscLeavesWithoutPushing(t *testing.T) {
	m := newTestModel(semver.Format{Prefix: "v"}, nil)

	emitted := press(m, escKey)

	if len(emitted) != 1 || !leavesToRepoList(emitted) {
		t.Errorf("emitted %v, want only a return to the repo list", emitted)
	}
}

func TestInvalidTagBlocksPush(t *testing.T) {
	m := newTestModel(semver.Format{Prefix: "bad prefix/"}, nil)
	press(m, enterKey)

	if emitted := press(m, yesKey); len(emitted) != 0 {
		t.Errorf("emitted %v for an invalid tag name, want nothing", emitted)
	}
}

func leavesToRepoList(emitted []tea.Msg) bool {
	for _, msg := range emitted {
		if state, ok := msg.(events.SetState); ok && state.State == domain.StateMain {
			return true
		}
	}
	return false
}
