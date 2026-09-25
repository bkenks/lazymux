package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// OpenRepoSettingsCmd reads the tags of the repo at dir, then opens the
// settings screen for its key.
func OpenRepoSettingsCmd(key, dir string) tea.Cmd {
	return func() tea.Msg {
		tags, err := repomgr.ListTags(dir)
		return events.OpenRepoSettings{Key: key, Tags: tags, TagsErr: err}
	}
}
