package events

import "github.com/bkenks/lazymux/internal/repomgr"

// CloneRepoExec reports the result of one repo clone, carrying the pending
// clone so the handler can apply the placeholder/insteadOf rewrite and record
// the forge link once the clone succeeds.
type CloneRepoExec struct {
	Clone repomgr.PendingClone
	Err   error
}

func (CloneRepoExec) isEvent() {}
