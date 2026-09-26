package events

import (
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/repomgr"
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
// forges removed or renamed there are cascaded into each repo's
// Upstreams/Origin so no repo is left referencing a forge that no longer
// exists.
type ForgesChanged struct {
	Forges []config.Forge
	Repos  map[string]config.RepoLink
}

func (ForgesChanged) isEvent() {}

// RepoSettingsChanged saves one repo's settings — its forge links
// (upstreams/origin/scheme) and tag format — and re-renders its git config.
type RepoSettingsChanged struct {
	Key  string
	Link config.RepoLink
}

func (RepoSettingsChanged) isEvent() {}

// OpenRepoSettings opens the per-repo settings screen for the given repo key.
// Tags are the repo's existing tags, for previewing its tag format; TagsErr is
// why they couldn't be read.
type OpenRepoSettings struct {
	Key     string
	Tags    []string
	TagsErr error
}

func (OpenRepoSettings) isEvent() {}
