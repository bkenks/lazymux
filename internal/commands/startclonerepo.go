package commands

import (
	"strings"

	"github.com/bkenks/lazymux/internal/events"
	tea "github.com/charmbracelet/bubbletea"
)

func StartCloneReposCmd(repoUrlsChunk string) tea.Cmd {
	repoUrls := strings.Split(strings.TrimSpace(repoUrlsChunk), "\n")

	return func() tea.Msg {
		return events.StartRepoClone{
			RepoUrls: repoUrls,
		}
	}
}
