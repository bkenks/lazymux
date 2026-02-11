package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func RefreshReposCmd() tea.Cmd {

	cmd := exec.Command("ghq", "list") // Call ghq to list repositories
	out, err := cmd.Output()

	if err != nil { // Fail-Safe
		fmt.Println("Error getting Repo List:", err)
		os.Exit(1)
	}

	// string(out) → "github.com/user/Repo1\ngithub.com/user/Repo2\n"
	// strings.TrimSpace(...) → removes the final \n, giving "github.com/user/Repo1\ngithub.com/user/Repo2"
	// strings.Split(..., "\n") → splits into strings on "\n"
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	repos := make([]list.Item, 0, len(lines)) // preallocate slice (i.e. set array size)

	/////////////////////////////////////
	// Format lines (list of repos like "github.com/user/Repo1") into a []list.Item
	for _, line := range lines {
		if line == "" { // Fail-Safe
			continue
		}

		parts := strings.Split(line, "/")    // split path ("github.com/user/Repo1") by "/"
		nameFromSplit := parts[len(parts)-1] // grab the last element from the split (which is the repo name)

		repos = append(repos, domain.Repo{ // Add to a []list.Item (array of `list.Item`s)
			Name: nameFromSplit,
			Path: line,
		})
	}
	//
	/////////////////////////////////////

	return func() tea.Msg {
		return events.ReposRefreshed{RepoList: repos}
	}
}
