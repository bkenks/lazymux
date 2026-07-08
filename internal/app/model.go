package app

import (
	"fmt"
	"time"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/repomgr"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/clonerepos"
	"github.com/bkenks/lazymux/internal/ui/confirm"
	"github.com/bkenks/lazymux/internal/ui/forgeregistry"
	"github.com/bkenks/lazymux/internal/ui/forgeselect"
	"github.com/bkenks/lazymux/internal/ui/repoforges"
	"github.com/bkenks/lazymux/internal/ui/repolist"
	"github.com/bkenks/lazymux/pkg/settings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const toastDuration = 4 * time.Second

type ModelManager struct {
	cfg config.Config

	state         domain.SessionState
	main          repolist.Model
	confirmDelete confirm.Model
	clonerepos    clonerepos.Model
	settingsModel settings.Model
	forgeSelect   *forgeselect.Model
	forgeRegistry *forgeregistry.Model
	repoForges    *repoforges.Model

	active tea.Model

	// clone batch progress (cloning runs after the forge-select step)
	cloneTotal int
	cloneDone  int
	cloneFail  int

	pendingDeleteKey string

	toast      string
	toastLevel events.ToastLevel
	toastSeq   int
}

func New(cfg config.Config) *ModelManager {
	commands.SetDeps(cfg)

	x, y := styles.DocStyle.GetFrameSize()
	settingsItems := buildSettingsItems(cfg)

	m := &ModelManager{
		cfg:           cfg,
		main:          *repolist.New(),
		confirmDelete: *confirm.New(),
		clonerepos:    *clonerepos.New(),
		settingsModel: settings.New("Settings", settingsItems, constants.WindowSize.Width, constants.WindowSize.Height, x, y),
	}

	m.active = &m.main
	return m
}

func (m *ModelManager) Init() tea.Cmd {
	cmds := []tea.Cmd{commands.RefreshReposCmd()}
	if w := m.cfg.LoadWarning; w != "" {
		cmds = append(cmds, m.toastCmd(events.ToastError, w))
	}
	return tea.Batch(cmds...)
}

func (m *ModelManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		constants.WindowSize = msg

	case events.Event:
		switch msg := msg.(type) {

		case events.SetState:
			m.state = msg.State

			switch m.state {
			case domain.StateMain:
				m.active = &m.main

			case domain.StateConfirmDelete:
				m.confirmDelete = *confirm.New()
				if repo, ok := m.main.List.SelectedItem().(domain.Repo); ok {
					m.confirmDelete.RepoPath = repo.Path
					m.confirmDelete.AbsPath = repo.AbsPath
					m.pendingDeleteKey = repo.Path
				}
				m.active = &m.confirmDelete

			case domain.StateCloneRepo:
				m.clonerepos = *clonerepos.New()
				m.active = &m.clonerepos

			case domain.StateSettings:
				m.active = &m.settingsModel

			case domain.StateForgeSelect:
				// m.forgeSelect is built in the StartRepoClone handler.
				if m.forgeSelect != nil {
					m.active = m.forgeSelect
				}

			case domain.StateForgeRegistry:
				m.forgeRegistry = forgeregistry.New(m.cfg)
				m.active = m.forgeRegistry

			case domain.StateRepoForges:
				// m.repoForges is built in the OpenRepoForges handler.
				if m.repoForges != nil {
					m.active = m.repoForges
				}
			}

		case events.StartRepoClone:
			// Parse each pasted URL and pre-select its matching forge, then
			// hand off to the forge-select screen before cloning.
			var pending []repomgr.PendingClone
			var bad int
			for _, raw := range msg.RepoUrls {
				if raw == "" {
					continue
				}
				p, err := repomgr.NewPendingClone(m.cfg, raw)
				if err != nil {
					bad++
					continue
				}
				pending = append(pending, p)
			}
			if bad > 0 {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("%d url(s) couldn't be parsed", bad)))
			}
			if len(pending) == 0 {
				cmds = append(cmds, commands.SetState(domain.StateMain))
				break
			}
			m.forgeSelect = forgeselect.New(m.cfg, pending)
			cmds = append(cmds, commands.SetState(domain.StateForgeSelect))

		case events.ForgeSelectComplete:
			// Persist any inline-added forges, then start the clones.
			if len(msg.NewForges) > 0 {
				m.cfg.Forges = append(m.cfg.Forges, msg.NewForges...)
				commands.SetDeps(m.cfg)
				if err := config.Save(m.cfg); err != nil {
					cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save forges: %v", err)))
				}
			}
			m.cloneTotal = len(msg.Clones)
			m.cloneDone = 0
			m.cloneFail = 0
			if m.cloneTotal == 0 {
				cmds = append(cmds, commands.SetState(domain.StateMain))
				break
			}
			cmds = append(cmds,
				commands.SetState(domain.StateMain),
				commands.CloneReposExecCmd(msg.Clones),
			)

		case events.CloneRepoExec:
			if msg.Err != nil {
				m.cloneFail++
			} else if msg.Clone.Primary != "" {
				// Clone succeeded: record the link and rewrite the repo to a
				// placeholder origin resolved to its primary forge.
				key := msg.Clone.URL.Key()
				link := msg.Clone.Link()
				m.cfg.Repos[key] = link
				if err := repomgr.RenderGitConfig(m.cfg, key, link); err != nil {
					cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("%s: %v", key, err)))
				}
			}
			if m.cloneDone < m.cloneTotal {
				m.cloneDone++
			}
			if m.cloneDone == m.cloneTotal {
				if err := config.Save(m.cfg); err != nil {
					cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save config: %v", err)))
				}
				summary := fmt.Sprintf("cloned %d/%d", m.cloneTotal-m.cloneFail, m.cloneTotal)
				cmds = append(cmds,
					commands.RefreshReposCmd(),
					m.toastCmd(events.ToastInfo, summary),
				)
			}

		case events.ForgesChanged:
			m.cfg.Forges = msg.Forges
			commands.SetDeps(m.cfg)
			if err := config.Save(m.cfg); err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save forges: %v", err)))
			}

		case events.OpenRepoForges:
			m.repoForges = repoforges.New(m.cfg, msg.Key)
			cmds = append(cmds, commands.SetState(domain.StateRepoForges))

		case events.RepoLinkChanged:
			if len(msg.Link.Forges) == 0 {
				delete(m.cfg.Repos, msg.Key)
			} else {
				m.cfg.Repos[msg.Key] = msg.Link
				if msg.Link.Primary != "" {
					if err := repomgr.RenderGitConfig(m.cfg, msg.Key, msg.Link); err != nil {
						cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("%s: %v", msg.Key, err)))
					}
				}
			}
			commands.SetDeps(m.cfg)
			if err := config.Save(m.cfg); err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save config: %v", err)))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.RepoDeleted:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("delete failed: %v", msg.Err)))
			} else {
				if m.pendingDeleteKey != "" {
					delete(m.cfg.Repos, m.pendingDeleteKey)
					m.pendingDeleteKey = ""
					_ = config.Save(m.cfg)
					commands.SetDeps(m.cfg)
				}
				cmds = append(cmds, m.toastCmd(events.ToastInfo, "repo deleted"))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.PullAllReposComplete:
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.ReposRefreshed:
			cmds = append(cmds, m.main.UpdateRepoList(msg.RepoList))

		case events.OpenInVSCodeComplete:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("editor failed: %v", msg.Err)))
			}
			cmds = append(cmds,
				commands.SetState(domain.StateMain),
				commands.RefreshReposCmd(),
			)

		case events.CmdComplete:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("command failed: %v", msg.Err)))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.Toast:
			m.toast = msg.Msg
			m.toastLevel = msg.Level
			m.toastSeq++
			seq := m.toastSeq
			cmds = append(cmds, tea.Tick(toastDuration, func(time.Time) tea.Msg {
				return events.ToastClear{Seq: seq}
			}))

		case events.ToastClear:
			if msg.Seq == 0 || msg.Seq == m.toastSeq {
				m.toast = ""
			}
		}

	case settings.SettingChanged:
		m.applySettingChange(msg)
		if err := config.Save(m.cfg); err != nil {
			cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save config: %v", err)))
		} else {
			cmds = append(cmds, m.toastCmd(events.ToastInfo, "settings saved"))
		}

	case settings.Exited:
		cmds = append(cmds, commands.SetState(domain.StateMain))
	}

	var cmd tea.Cmd
	m.active, cmd = m.active.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *ModelManager) View() string {
	body := styles.DocStyle.Render(m.active.View())
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderToast())
}

func (m *ModelManager) renderToast() string {
	width := constants.WindowSize.Width
	if m.toast == "" {
		return styles.ToastIdleStyle.Width(width).Render("")
	}
	style := styles.ToastInfoStyle
	if m.toastLevel == events.ToastError {
		style = styles.ToastErrorStyle
	}
	return style.Width(width).Render(m.toast)
}

func (m *ModelManager) toastCmd(level events.ToastLevel, msg string) tea.Cmd {
	return func() tea.Msg {
		return events.Toast{Msg: msg, Level: level}
	}
}
