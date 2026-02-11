package events

import "github.com/charmbracelet/bubbles/list"

type ReposRefreshed struct{ RepoList []list.Item }

func (ReposRefreshed) isEvent() {}
