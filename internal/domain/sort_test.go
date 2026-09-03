package domain

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

func repoItems() []list.Item {
	now := time.Now()
	return []list.Item{
		Repo{Name: "beta", Path: "acme/beta", LastInteracted: now.Add(-time.Hour)},
		Repo{Name: "alpha", Path: "zulu/alpha"},
		Repo{Name: "Gamma", Path: "acme/Gamma", LastInteracted: now},
	}
}

func names(items []list.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.(Repo).Name
	}
	return out
}

func TestSortRepos(t *testing.T) {
	tests := []struct {
		mode SortMode
		want []string
	}{
		// Never-interacted repos fall to the end of the recency order.
		{SortRecent, []string{"Gamma", "beta", "alpha"}},
		{SortNameAsc, []string{"alpha", "beta", "Gamma"}},
		{SortNameDesc, []string{"Gamma", "beta", "alpha"}},
		{SortNamespace, []string{"beta", "Gamma", "alpha"}},
	}
	for _, tt := range tests {
		items := repoItems()
		SortRepos(items, tt.mode)
		got := names(items)
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("%s = %v, want %v", tt.mode, got, tt.want)
				break
			}
		}
	}
}

func TestSortModeNextCycles(t *testing.T) {
	mode := SortModes[0]
	for range SortModes {
		mode = mode.Next()
	}
	if mode != SortModes[0] {
		t.Errorf("cycling every mode landed on %q, want %q", mode, SortModes[0])
	}
	if got := SortMode("bogus").Next(); got != SortModes[0] {
		t.Errorf("unknown mode cycled to %q, want %q", got, SortModes[0])
	}
}

func TestParseSortMode(t *testing.T) {
	if got := ParseSortMode("name-desc"); got != SortNameDesc {
		t.Errorf("ParseSortMode = %q, want %q", got, SortNameDesc)
	}
	if got := ParseSortMode("nonsense"); got != SortRecent {
		t.Errorf("ParseSortMode fallback = %q, want %q", got, SortRecent)
	}
}
