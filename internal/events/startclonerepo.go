package events

type StartRepoClone struct{ RepoUrls []string }

func (StartRepoClone) isEvent() {}
