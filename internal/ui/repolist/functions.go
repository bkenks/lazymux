package repolist

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/list"
)

func ConvertToRepoType(i list.Item) domain.Repo {
	if domainRepo, ok := i.(domain.Repo); ok {
		return domainRepo
	}
	return domain.Repo{}
}

func AbsRepoPath(i list.Item) string {
	return ConvertToRepoType(i).AbsPath
}

func SizeBuffer() (width, height int) {
	x, y := styles.DocStyle.GetFrameSize()
	width = constants.WindowSize.Width - x
	height = constants.WindowSize.Height - y - constants.FooterReservedLines
	if height < 1 {
		height = 1
	}
	return width, height
}

func formatPullSummary(msg events.PullAllReposComplete) string {
	if len(msg.Skipped) == 0 {
		return fmt.Sprintf("Pulled %d repos.", msg.Pulled)
	}

	names := make([]string, 0, len(msg.Skipped))
	for _, s := range msg.Skipped {
		names = append(names, filepath.Base(s.RepoPath))
	}
	return fmt.Sprintf("Pulled %d, skipped %d: %s",
		msg.Pulled, len(msg.Skipped), strings.Join(names, ", "))
}
