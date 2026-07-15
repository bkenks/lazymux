package domain

import (
	"path"
	"time"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Interface
//
// Repo (list.Item):
//	- Represents a repository

type Repo struct {
	Name, Path     string
	AbsPath        string
	LastInteracted time.Time

	// Forge links, populated from config for repos under the lazymux base dir.
	Forges  []string
	Primary string
	Scheme  string
}

func (r Repo) Title() string { return r.Name }

// Namespace is the repo's path with its own name stripped — e.g. "bkenks" for
// "bkenks/myrepo", or "foo/bar" for "foo/bar/repo". Empty for a repo sitting
// directly under the base dir.
func (r Repo) Namespace() string {
	ns := path.Dir(r.Path)
	if ns == "." {
		return ""
	}
	return ns
}

// ShowForge controls whether Description() includes the "forge:" line. It's
// toggled from the repo list; package-level to match the app's other view-state
// globals (constants.WindowSize, the style vars).
var ShowForge = true

func (r Repo) Description() string {
	if ShowForge && r.Primary != "" {
		return r.Namespace() + "\nforge: " + r.Primary
	}
	return r.Namespace()
}
func (r Repo) FilterValue() string { return r.Name }

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
