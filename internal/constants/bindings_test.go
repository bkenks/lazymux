package constants

import (
	"os"
	"strings"
	"testing"
)

func TestReadmeListsEveryRepoListKey(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range RepoListKeyMap.Commands() {
		row := "| `" + c.Binding.Help().Key + "`"
		if !strings.Contains(string(readme), row) {
			t.Errorf("README keybindings table has no row for %q (%s)", c.Binding.Help().Key, c.Full)
		}
	}
}

func TestAllCoversEveryCommand(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range RepoListKeyMap.All() {
		for _, k := range b.Keys() {
			if seen[k] {
				t.Errorf("key %q bound twice on the repo list", k)
			}
			seen[k] = true
		}
	}
	if len(RepoListKeyMap.All()) != len(RepoListKeyMap.Commands()) {
		t.Error("All and Commands disagree")
	}
}
