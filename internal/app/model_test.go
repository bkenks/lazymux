package app

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
)

func TestSavedAccentRestylesTheRepoList(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", t.TempDir()+"/.lazymux.json")
	t.Cleanup(func() { styles.Apply("default", true, nil) })
	m := New(config.Default(), "test")

	edited := m.cfg.Clone()
	edited.UI.AccentColor = "#123456"
	_, cmd := m.Update(events.SettingsChanged{Config: edited})

	if got := config.Load().UI.AccentColor; got != "#123456" {
		t.Errorf("saved accent = %q, want #123456", got)
	}
	want, _ := styles.ParseAccent("#123456")
	if got := m.main.List.Styles.Title.GetBackground(); got != want {
		t.Errorf("repo list title = %v, want the new accent %v", got, want)
	}
	for _, msg := range collectMsgs(cmd) {
		if state, ok := msg.(events.SetState); ok && state.State == domain.StateMain {
			return
		}
	}
	t.Error("saving settings did not return to the repo list")
}

// collectMsgs runs cmd and any batch inside it, skipping commands that don't
// return quickly (timers), and returns the messages produced.
func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	result := make(chan tea.Msg, 1)
	go func() { result <- cmd() }()
	select {
	case msg := <-result:
		if batch, ok := msg.(tea.BatchMsg); ok {
			var msgs []tea.Msg
			for _, inner := range batch {
				msgs = append(msgs, collectMsgs(inner)...)
			}
			return msgs
		}
		return []tea.Msg{msg}
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func TestPullKeepsGoingWhileAnotherScreenIsOpen(t *testing.T) {
	t.Setenv("LAZYMUX_CONFIG", t.TempDir()+"/.lazymux.json")
	m := New(config.Default(), "test")
	m.Update(events.SetState{State: domain.StateSettings})

	results := make(chan events.PullResult, 1)
	results <- events.PullResult{RepoPath: "me/demo"}
	_, cmd := m.Update(events.PullAllStarted{Total: 1, Results: results})

	for _, msg := range collectMsgs(cmd) {
		if _, ok := msg.(events.PullResult); ok {
			return
		}
	}
	t.Error("pull results stopped being read once another screen was active")
}
