package commands

import (
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/repomgr"
)

// CloneReposExecCmd clones each pending repo against its real URL via
// tea.ExecProcess (so git can prompt for credentials with the terminal
// attached). The placeholder-origin + insteadOf rewrite is applied afterwards,
// once the clone succeeds, in the CloneRepoExec handler.
func CloneReposExecCmd(clones []repomgr.PendingClone) tea.Cmd {
	var cmds []tea.Cmd
	baseDir := cfg().RepoRoot()

	for _, c := range clones {
		clone := c
		dest := repomgr.RepoDir(baseDir, clone.URL.Key())
		cmds = append(cmds, tea.ExecProcess(
			exec.Command("git", "clone", clone.RealURL, dest),
			func(err error) tea.Msg {
				return events.CloneRepoExec{Clone: clone, Err: err}
			},
		))
	}

	return tea.Batch(cmds...)
}
