package repolist

import (
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
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
