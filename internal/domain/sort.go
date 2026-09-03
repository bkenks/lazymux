package domain

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// SortMode is the ordering applied to the repo list.
type SortMode string

const (
	// SortRecent puts the most recently opened repos first, never-opened last.
	SortRecent SortMode = "recent"
	// SortNameAsc orders by repo name, A→Z.
	SortNameAsc SortMode = "name-asc"
	// SortNameDesc orders by repo name, Z→A.
	SortNameDesc SortMode = "name-desc"
	// SortNamespace groups by namespace, then orders by name inside it.
	SortNamespace SortMode = "namespace"
)

// SortModes is every mode, in the order the repo list cycles through them.
var SortModes = []SortMode{SortRecent, SortNameAsc, SortNameDesc, SortNamespace}

// Sort is the mode the repo list is currently ordered by. It's package-level
// mutable to match the app's other view-state globals (ShowForge, ShowStats).
var Sort = SortRecent

var sortLabels = map[SortMode]string{
	SortRecent:    "recent",
	SortNameAsc:   "name a-z",
	SortNameDesc:  "name z-a",
	SortNamespace: "namespace",
}

// Label is the short name shown in the list title and settings screen.
func (s SortMode) Label() string {
	if l, ok := sortLabels[s]; ok {
		return l
	}
	return sortLabels[SortRecent]
}

// Next returns the mode after s in the cycle, wrapping at the end. An
// unrecognized mode cycles to the first one.
func (s SortMode) Next() SortMode {
	for i, m := range SortModes {
		if m == s {
			return SortModes[(i+1)%len(SortModes)]
		}
	}
	return SortModes[0]
}

// ParseSortMode maps a persisted value onto a known mode, falling back to the
// default for anything unrecognized.
func ParseSortMode(s string) SortMode {
	for _, m := range SortModes {
		if string(m) == s {
			return m
		}
	}
	return SortRecent
}

// SortRepos orders repo items in place by the given mode. Items that aren't
// repos are left where they are.
func SortRepos(items []list.Item, mode SortMode) {
	sort.SliceStable(items, func(i, j int) bool {
		a, aok := items[i].(Repo)
		b, bok := items[j].(Repo)
		if !aok || !bok {
			return false
		}
		return lessRepo(a, b, mode)
	})
}

func lessRepo(a, b Repo, mode SortMode) bool {
	switch mode {
	case SortNameAsc:
		return lessName(a, b)
	case SortNameDesc:
		return lessName(b, a)
	case SortNamespace:
		if ns := strings.Compare(strings.ToLower(a.Namespace()), strings.ToLower(b.Namespace())); ns != 0 {
			return ns < 0
		}
		return lessName(a, b)
	default:
		// Recency: newest first, with never-opened repos falling to the end.
		if a.LastInteracted.Equal(b.LastInteracted) {
			return lessName(a, b)
		}
		return a.LastInteracted.After(b.LastInteracted)
	}
}

// lessName compares repo names case-insensitively, falling back to the full
// path so two repos of the same name keep a stable order.
func lessName(a, b Repo) bool {
	if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
		return c < 0
	}
	return a.Path < b.Path
}
