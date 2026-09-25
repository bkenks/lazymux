package events

// ReposDirChosen carries the repo directory the user entered, as typed.
type ReposDirChosen struct{ Dir string }

func (ReposDirChosen) isEvent() {}
