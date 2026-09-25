package commands

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// OpenTagReleaseCmd reads the tags of the repo at dir, then opens the release
// screen for it. A repo whose tags can't be read toasts why instead.
func OpenTagReleaseCmd(key, dir string) tea.Cmd {
	return func() tea.Msg {
		tags, err := repomgr.ListTags(dir)
		if err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("reading tags: %v", err)}
		}
		return events.OpenTagRelease{Key: key, Dir: dir, Tags: tags}
	}
}

// ReleaseTagCmd creates tag at HEAD of the repo at dir and pushes it to origin.
func ReleaseTagCmd(dir, tag string) tea.Cmd {
	return func() tea.Msg {
		return events.TagReleased{Tag: tag, Err: repomgr.ReleaseTag(dir, tag)}
	}
}
