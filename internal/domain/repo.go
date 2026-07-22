package domain

import (
	"fmt"
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

	// Local git state, used to gauge whether a repo can be wiped safely. These
	// are independent signals: LocalBranches counts refs/heads; UnpushedCommits
	// counts commits reachable from local branches but absent from every
	// remote-tracking ref; UncommittedFiles counts working-tree paths that
	// aren't committed at all. The last two are disjoint categories of work a
	// delete would lose.
	LocalBranches    int
	UnpushedCommits  int
	UncommittedFiles int
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

// ShowStats controls whether Description() includes the git stats summary
// (branches, unpushed commits, uncommitted files). Toggled from the repo list
// like ShowForge.
var ShowStats = true

func (r Repo) Description() string {
	line := r.Namespace()
	if stats := r.GitStatsLabel(); ShowStats && stats != "" {
		if line != "" {
			line += "  ·  "
		}
		line += stats
	}
	if ShowForge && r.Primary != "" {
		return line + "\nforge: " + r.Primary
	}
	return line
}

// GitStatsLabel summarizes local git state for the list row: branch count plus
// any unpushed commits or uncommitted changes — the signals for whether a repo
// still holds work that a delete would lose. Empty for repos with no branches
// (e.g. a bare or freshly-initialized repo).
func (r Repo) GitStatsLabel() string {
	if r.LocalBranches == 0 {
		return ""
	}
	label := fmt.Sprintf("%d %s", r.LocalBranches, plural(r.LocalBranches, "branch", "branches"))
	if r.UnpushedCommits > 0 {
		label += fmt.Sprintf(", %d unpushed", r.UnpushedCommits)
	}
	if r.UncommittedFiles > 0 {
		label += fmt.Sprintf(", %d uncommitted %s", r.UncommittedFiles,
			plural(r.UncommittedFiles, "file", "files"))
	}
	return label
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
func (r Repo) FilterValue() string { return r.Name }

// End "Interface"
///////////////////////////////////////////////////////////////////////////////////////////////////////////////
