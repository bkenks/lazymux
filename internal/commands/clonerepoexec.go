package commands

import (
	"os/exec"

	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	tea "github.com/charmbracelet/bubbletea"
)

// CloneReposExecCmd clones each pending repo against its real URL via
// tea.ExecProcess (so git can prompt for credentials with the terminal
// attached). The placeholder-origin + insteadOf rewrite is applied afterwards,
// once the clone succeeds, in the CloneRepoExec handler.
func CloneReposExecCmd(clones []repomgr.PendingClone) tea.Cmd {
	var cmds []tea.Cmd
	baseDir := cfg().BaseDir

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
