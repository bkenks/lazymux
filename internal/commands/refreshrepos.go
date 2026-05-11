package commands

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func RefreshReposCmd() tea.Cmd {
	return func() tea.Msg {
		c := cfg()

		// `ghq list` gives the short paths shown to the user;
		// `ghq list --full-path` gives the on-disk absolute paths in the same order.
		shortOut, err := exec.Command(c.Tools.Ghq, "list").Output()
		if err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("ghq list failed: %v", err)}
		}
		fullOut, err := exec.Command(c.Tools.Ghq, "list", "--full-path").Output()
		if err != nil {
			return events.Toast{Level: events.ToastError, Msg: fmt.Sprintf("ghq list --full-path failed: %v", err)}
		}

		shortLines := splitNonEmpty(string(shortOut))
		fullLines := splitNonEmpty(string(fullOut))

		// Defensive: if the two listings disagree in length, fall back to mapping by suffix.
		absByShort := make(map[string]string, len(shortLines))
		if len(shortLines) == len(fullLines) {
			for i, s := range shortLines {
				absByShort[s] = fullLines[i]
			}
		} else {
			for _, s := range shortLines {
				for _, full := range fullLines {
					if strings.HasSuffix(full, s) {
						absByShort[s] = full
						break
					}
				}
			}
		}

		interactions := domain.LoadInteractions()
		repos := make([]list.Item, 0, len(shortLines))
		for _, line := range shortLines {
			parts := strings.Split(line, "/")
			repos = append(repos, domain.Repo{
				Name:           parts[len(parts)-1],
				Path:           line,
				AbsPath:        absByShort[line],
				LastInteracted: interactions[line],
			})
		}

		// Most recently interacted first; never-interacted fall to the end.
		sort.SliceStable(repos, func(i, j int) bool {
			ri := repos[i].(domain.Repo)
			rj := repos[j].(domain.Repo)
			return ri.LastInteracted.After(rj.LastInteracted)
		})

		return events.ReposRefreshed{RepoList: repos}
	}
}

func splitNonEmpty(s string) []string {
	raw := strings.Split(strings.TrimSpace(s), "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}
