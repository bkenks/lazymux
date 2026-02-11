package events

type RepoDeleted struct {
	Err error
}

func (RepoDeleted) isEvent() {}
