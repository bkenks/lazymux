package repolist

import (
	"charm.land/bubbles/v2/list"
	"github.com/bkenks/gitkeeper/internal/domain"
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
