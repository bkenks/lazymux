package events

import "charm.land/bubbles/v2/list"

type ReposRefreshed struct{ RepoList []list.Item }

func (ReposRefreshed) isEvent() {}
