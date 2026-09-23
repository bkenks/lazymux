package repomgr

import (
	"strings"

	"github.com/bkenks/lazymux/internal/config"
)

// PendingClone is a repo the user is about to clone, together with the forge
// links they've chosen for it. It's built with auto-matched defaults and then
// adjusted in the forge-select screen before the clone runs.
type PendingClone struct {
	RealURL string  // the URL the user pasted, trimmed — cloned against directly
	URL     RepoURL // parsed form
	config.RepoLink
}

// NewPendingClone parses raw and pre-selects the forge whose host matches the
// URL (if any), making it the origin. Scheme defaults to the URL's own scheme.
func NewPendingClone(cfg config.Config, raw string) (PendingClone, error) {
	u, err := ParseRepoURL(raw)
	if err != nil {
		return PendingClone{}, err
	}
	p := PendingClone{RealURL: strings.TrimSpace(raw), URL: u}
	p.Scheme = u.Scheme
	if f, ok := cfg.ForgeByHost(u.Host); ok {
		p.SetOrigin(f.Name)
	}
	return p, nil
}

// Link converts the selection into the config record persisted for the repo.
func (p PendingClone) Link() config.RepoLink {
	link := p.RepoLink
	return link.Clone()
}
