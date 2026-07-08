package domain

import "time"

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
func (r Repo) Description() string {
	if r.Primary != "" {
		return r.Path + "  ⟶ " + r.Primary
	}
	return r.Path
}
func (r Repo) FilterValue() string { return r.Name }

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
