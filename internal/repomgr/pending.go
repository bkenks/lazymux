package repomgr

import "github.com/bkenks/lazymux/internal/config"

// PendingClone is a repo the user is about to clone, together with the forge
// links they've chosen for it. It's built with auto-matched defaults and then
// adjusted in the forge-select screen before the clone runs.
type PendingClone struct {
	RealURL   string  // the URL the user pasted — cloned against directly
	URL       RepoURL // parsed form
	Upstreams []string
	Origin    string
	Scheme    string
}

// NewPendingClone parses raw and pre-selects the forge whose host matches the
// URL (if any), making it the origin. Scheme defaults to the URL's own scheme.
func NewPendingClone(cfg config.Config, raw string) (PendingClone, error) {
	u, err := ParseRepoURL(raw)
	if err != nil {
		return PendingClone{}, err
	}
	p := PendingClone{RealURL: raw, URL: u, Scheme: u.Scheme}
	if f, ok := cfg.ForgeByHost(u.Host); ok {
		p.Upstreams = []string{f.Name}
		p.Origin = f.Name
	}
	return p, nil
}

// Link converts the selection into the config record persisted for the repo.
func (p PendingClone) Link() config.RepoLink {
	return config.RepoLink{Upstreams: p.Upstreams, Origin: p.Origin, Scheme: p.Scheme}
}

// HasForge reports whether name is among the selected upstream forges.
func (p PendingClone) HasForge(name string) bool {
	for _, f := range p.Upstreams {
		if f == name {
			return true
		}
	}
	return false
}
