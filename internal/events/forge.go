package events

import (
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// ForgeSelectComplete is emitted by the clone-time forge-select screen once the
// user has chosen forge links for every pending repo. NewForges carries any
// forges added inline (via "add forge from URL") so they get persisted.
type ForgeSelectComplete struct {
	Clones    []repomgr.PendingClone
	NewForges []config.Forge
}

func (ForgeSelectComplete) isEvent() {}

// ForgesChanged replaces the whole forge registry (from the settings screen).
// Repos carries the repo links already reconciled against the new registry —
// forges removed or renamed there are cascaded into each repo's Forges/Primary
// so no repo is left referencing a forge that no longer exists.
type ForgesChanged struct {
	Forges []config.Forge
	Repos  map[string]config.RepoLink
}

func (ForgesChanged) isEvent() {}

// RepoLinkChanged updates one repo's forge link (primary/scheme/forges) and
// re-renders its git config.
type RepoLinkChanged struct {
	Key  string
	Link config.RepoLink
}

func (RepoLinkChanged) isEvent() {}

// OpenRepoForges opens the per-repo forge editor for the given repo key.
type OpenRepoForges struct{ Key string }

func (OpenRepoForges) isEvent() {}
