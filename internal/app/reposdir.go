package app

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/commands"
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/domain"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/ui/reposdir"
)

// showMainOrReposDir opens the repo list, or the repo directory prompt in its
// place while the repo directory is unset or missing.
func (m *ModelManager) showMainOrReposDir() tea.Cmd {
	problem := m.cfg.ValidateRepoRoot()
	if problem == nil {
		m.state = domain.StateMain
		m.active = &m.main
		return nil
	}
	m.state = domain.StateReposDir
	m.reposDir = reposdir.New(m.cfg, problem)
	m.active = m.reposDir
	return m.reposDir.Init()
}

// applyReposDir creates and saves the chosen repo directory, then scans it and
// returns to the repo list. $GITKEEPER_REPOS would override the saved value, so
// it is dropped for this session and the user is told to unset it.
func (m *ModelManager) applyReposDir(input string) tea.Cmd {
	dir, err := config.ParseReposDir(input)
	if err == nil {
		err = os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return tea.Batch(m.showMainOrReposDir(),
			m.toastCmd(events.ToastError, fmt.Sprintf("couldn't use %s: %v", input, err)))
	}

	notice := m.toastCmd(events.ToastInfo, "repos live in "+dir)
	if envDir := os.Getenv(config.ReposEnvVar); envDir != "" {
		_ = os.Unsetenv(config.ReposEnvVar) // Unsetenv only fails on an invalid name.
		notice = m.toastCmd(events.ToastError, fmt.Sprintf(
			"saved %s, but $%s=%s overrides it next launch; unset it", dir, config.ReposEnvVar, envDir))
	}
	if saveFailed := m.saveConfig("repo directory", func(c *config.Config) { c.ReposDir = dir }); saveFailed != nil {
		notice = saveFailed
	}
	return tea.Batch(commands.RefreshReposCmd(), commands.SetState(domain.StateMain), notice)
}
