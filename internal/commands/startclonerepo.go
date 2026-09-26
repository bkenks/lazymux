package commands

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/events"
)

func StartCloneReposCmd(repoUrlsChunk string) tea.Cmd {
	repoUrls := strings.Split(strings.TrimSpace(repoUrlsChunk), "\n")

	return func() tea.Msg {
		return events.StartRepoClone{
			RepoUrls: repoUrls,
		}
	}
}
