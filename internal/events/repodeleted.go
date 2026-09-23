package events

// RepoDeleted reports the result of deleting the repo stored under Key.
type RepoDeleted struct {
	Key string
	Err error
}

func (RepoDeleted) isEvent() {}
