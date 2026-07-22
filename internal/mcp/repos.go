// Package mcp exposes the lazymux repo inventory over the Model Context
// Protocol, so an LLM can answer "which repo does this request belong to?"
// and record what it learns back into .lazymux.json.
package mcp

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/repomgr"
)

// RepoInfo is a single repo as an LLM sees it: where it lives, what it's for,
// and which forge hosts it.
type RepoInfo struct {
	Key       string `json:"key" jsonschema:"stable identifier, '<namespace>/<name>'"`
	Name      string `json:"name" jsonschema:"directory name of the repo"`
	Namespace string `json:"namespace,omitempty" jsonschema:"owner/org path the repo sits under"`
	Path      string `json:"path" jsonschema:"absolute path to the repo on disk"`

	Purpose string `json:"purpose,omitempty" jsonschema:"one-line summary of what this repo is for"`
	Context string `json:"context,omitempty" jsonschema:"longer notes: stack, conventions, when to use it"`

	Forges  []string `json:"forges,omitempty" jsonschema:"registered forges hosting this repo"`
	Primary string   `json:"primary,omitempty" jsonschema:"forge that origin currently resolves to"`

	LastInteracted string `json:"lastInteracted,omitempty" jsonschema:"RFC3339 time the user last opened this repo, if ever"`
}

// Described reports whether the repo has any purpose/context recorded.
func (r RepoInfo) Described() bool { return r.Purpose != "" || r.Context != "" }

// inventory lists every managed repo, newest interaction first so the repos
// the user actually works in lead the list.
func inventory(cfg config.Config) ([]RepoInfo, error) {
	repos, err := repomgr.ListMeta(cfg)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", cfg.BaseDir, err)
	}
	infos := make([]RepoInfo, 0, len(repos))
	for _, r := range repos {
		link := cfg.Repos[r.Path]
		info := RepoInfo{
			Key:       r.Path,
			Name:      r.Name,
			Namespace: r.Namespace(),
			Path:      r.AbsPath,
			Purpose:   link.Purpose,
			Context:   link.Context,
			Forges:    r.Forges,
			Primary:   r.Primary,
		}
		if !r.LastInteracted.IsZero() {
			info.LastInteracted = r.LastInteracted.UTC().Format(timeLayout)
		}
		infos = append(infos, info)
	}
	sort.SliceStable(infos, func(i, j int) bool {
		if infos[i].LastInteracted != infos[j].LastInteracted {
			return infos[i].LastInteracted > infos[j].LastInteracted
		}
		return infos[i].Key < infos[j].Key
	})
	return infos, nil
}

const timeLayout = "2006-01-02T15:04:05Z"

// find returns the repo with the given key, or an error naming near misses so
// a model that guessed the key wrong can correct itself in one turn.
func find(infos []RepoInfo, key string) (RepoInfo, error) {
	for _, r := range infos {
		if r.Key == key {
			return r, nil
		}
	}
	var suggestions []string
	for _, r := range infos {
		if strings.EqualFold(r.Name, key) || strings.Contains(r.Key, key) {
			suggestions = append(suggestions, r.Key)
		}
	}
	if len(suggestions) > 0 {
		return RepoInfo{}, fmt.Errorf("no repo with key %q; did you mean %s?",
			key, strings.Join(suggestions, ", "))
	}
	return RepoInfo{}, fmt.Errorf("no repo with key %q; call list_repositories for valid keys", key)
}

// search ranks repos against a free-text query. Terms are matched
// case-insensitively against the fields most likely to identify a repo, and
// weighted so a name hit outranks an incidental mention in the context blurb.
// Repos matching no term are dropped.
//
// Ranking is by how many distinct query terms a repo matches first, and only
// then by weighted score: a repo covering every word of "docs api" beats one
// that merely mentions "api" three times.
func search(infos []RepoInfo, query string) []RepoInfo {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return infos
	}
	type scored struct {
		info    RepoInfo
		matched int
		weight  int
	}
	var hits []scored
	for _, r := range infos {
		hit := scored{info: r}
		for _, term := range terms {
			if s := termScore(r, term); s > 0 {
				hit.matched++
				hit.weight += s
			}
		}
		if hit.matched > 0 {
			hits = append(hits, hit)
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].matched != hits[j].matched {
			return hits[i].matched > hits[j].matched
		}
		return hits[i].weight > hits[j].weight
	})
	out := make([]RepoInfo, len(hits))
	for i, h := range hits {
		out[i] = h.info
	}
	return out
}

func termScore(r RepoInfo, term string) int {
	weights := []struct {
		field  string
		weight int
	}{
		{r.Name, 5},
		{r.Key, 3},
		{r.Purpose, 3},
		{r.Context, 1},
	}
	score := 0
	for _, w := range weights {
		if strings.Contains(strings.ToLower(w.field), term) {
			score += w.weight
		}
	}
	return score
}

// setDescription writes purpose/context for one repo back to .lazymux.json.
// The config is re-read immediately before the write so a concurrently running
// TUI's edits to unrelated fields survive.
func setDescription(key, purpose, context string) (RepoInfo, error) {
	cfg := config.Load()
	if cfg.LoadWarning != "" {
		return RepoInfo{}, fmt.Errorf("refusing to write over an unreadable config: %s", cfg.LoadWarning)
	}
	infos, err := inventory(cfg)
	if err != nil {
		return RepoInfo{}, err
	}
	info, err := find(infos, key)
	if err != nil {
		return RepoInfo{}, err
	}
	if _, err := os.Stat(info.Path); err != nil {
		return RepoInfo{}, fmt.Errorf("repo %q is listed but missing at %s", key, info.Path)
	}

	link := cfg.Repos[key]
	if purpose != "" {
		link.Purpose = purpose
		info.Purpose = purpose
	}
	if context != "" {
		link.Context = context
		info.Context = context
	}
	cfg.Repos[key] = link

	if err := config.Save(cfg); err != nil {
		return RepoInfo{}, fmt.Errorf("writing %s: %w", config.Path(), err)
	}
	return info, nil
}
