package events

type CloneRepoExec struct{ Err error }

func (CloneRepoExec) isEvent() {}
