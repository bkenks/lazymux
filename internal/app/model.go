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
	"github.com/bkenks/lazymux/internal/ui/splash"
	"github.com/bkenks/lazymux/pkg/settings"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	toastHold     = 4 * time.Second       // full-opacity dwell before fading out
	toastFrame    = 45 * time.Millisecond // fade animation frame interval
	toastFadeStep = 0.18                  // opacity delta per fade frame
)

type toastPhase int

const (
	toastFadeIn toastPhase = iota
	toastHolding
	toastFadeOut
)

type ModelManager struct {
	cfg config.Config

	state         domain.SessionState
	splash        splash.Model
	main          repolist.Model
	confirmDelete confirm.Model
	clonerepos    clonerepos.Model
	settingsModel settings.Model
	forgeSelect   *forgeselect.Model
	forgeRegistry *forgeregistry.Model
	repoForges    *repoforges.Model

	active tea.Model

	// clone batch progress (cloning runs after the forge-select step)
	cloneTotal    int
	cloneDone     int
	cloneFail     int
	cloneProgress progress.Model

	pendingDeleteKey string

	toast        string
	toastLevel   events.ToastLevel
	toastSeq     int
	toastOpacity float64
	toastPhase   toastPhase
}

func New(cfg config.Config, version string) *ModelManager {
	commands.SetDeps(cfg)

	// Restore the persisted forge-label visibility before building the repo
	// list, so its row height is sized correctly from the start.
	domain.ShowForge = cfg.UI.ShowForge

	x, y := styles.DocStyle.GetFrameSize()
	settingsItems := buildSettingsItems(cfg)

	m := &ModelManager{
		cfg:           cfg,
		splash:        *splash.New(version),
		main:          *repolist.New(),
		confirmDelete: *confirm.New(),
		clonerepos:    *clonerepos.New(),
		settingsModel: settings.New("Settings", settingsItems, constants.WindowSize.Width, constants.WindowSize.Height, x, y),
		cloneProgress: progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
	}

	m.state = domain.StateSplash
	m.active = &m.splash
	return m
}

func (m *ModelManager) Init() tea.Cmd {
	cmds := []tea.Cmd{m.splash.Init(), commands.RefreshReposCmd()}
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
			case domain.StateSplash:
				m.active = &m.splash

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
				// Rebuild with the live window size (and current cfg) — like the
				// other screens — since the startup build ran at size 0×0.
				x, y := styles.DocStyle.GetFrameSize()
				m.settingsModel = settings.New("Settings", buildSettingsItems(m.cfg),
					constants.WindowSize.Width, constants.WindowSize.Height, x, y)
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

			// Re-broadcast the window size so the newly-active screen lays out at
			// the current dimensions — the repo list is built at size 0 behind the
			// splash and would otherwise stay unsized until the next real resize.
			if constants.WindowSize.Width > 0 {
				sz := constants.WindowSize
				cmds = append(cmds, func() tea.Msg { return sz })
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
			prevForges := m.cfg.Forges
			prevRepos := m.cfg.Repos
			m.cfg.Forges = msg.Forges
			m.cfg.Repos = msg.Repos
			commands.SetDeps(m.cfg)

			// Re-render the git remote for any repo whose primary forge changed
			// name or host (promotion after a delete, a rename, or a host edit).
			// A repo left unlinked keeps its existing remote — the insteadOf
			// rule already written still resolves.
			for key, link := range m.cfg.Repos {
				if link.Primary == "" {
					continue
				}
				newForge, ok := m.cfg.ForgeByName(link.Primary)
				if !ok {
					continue // dangling primary; leave the existing config alone
				}
				prev := prevRepos[key]
				if prev.Primary == link.Primary && forgeHost(prevForges, prev.Primary) == newForge.Host {
					continue // nothing affecting the remote changed
				}
				if err := repomgr.RenderGitConfig(m.cfg, key, link); err != nil {
					cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("%s: %v", key, err)))
				}
			}

			if err := config.Save(m.cfg); err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save forges: %v", err)))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

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
			cmds = append(cmds,
				commands.RefreshReposCmd(),
				m.toastCmd(events.ToastInfo, msg.Summary()),
			)

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
			m.toastOpacity = 0
			m.toastPhase = toastFadeIn
			cmds = append(cmds, toastTick(m.toastSeq))

		case events.ToastAnim:
			if msg.Seq != m.toastSeq {
				break // stale animation from a superseded toast
			}
			switch m.toastPhase {
			case toastFadeIn:
				if m.toastOpacity += toastFadeStep; m.toastOpacity >= 1 {
					m.toastOpacity = 1
					m.toastPhase = toastHolding
					cmds = append(cmds, tea.Tick(toastHold, toastAnimMsg(msg.Seq)))
				} else {
					cmds = append(cmds, toastTick(msg.Seq))
				}
			case toastHolding:
				m.toastPhase = toastFadeOut
				cmds = append(cmds, toastTick(msg.Seq))
			case toastFadeOut:
				if m.toastOpacity -= toastFadeStep; m.toastOpacity <= 0 {
					m.toastOpacity = 0
					m.toast = ""
				} else {
					cmds = append(cmds, toastTick(msg.Seq))
				}
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
	// The footer region is a single reserved line (FooterReservedLines). A clone
	// batch in flight owns it — showing a live gradient bar between the per-repo
	// terminal handovers — otherwise it's the toast line.
	footer := m.renderToast()
	if line := m.renderCloneProgress(); line != "" {
		footer = line
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

// renderCloneProgress draws a gradient bar while a clone batch is in flight.
// Empty once every repo is done, so the final summary toast can take the footer.
func (m *ModelManager) renderCloneProgress() string {
	if m.cloneTotal == 0 || m.cloneDone >= m.cloneTotal {
		return ""
	}
	bar := constants.WindowSize.Width - 24
	switch {
	case bar > 40:
		bar = 40
	case bar < 10:
		bar = 10
	}
	m.cloneProgress.Width = bar
	pct := float64(m.cloneDone) / float64(m.cloneTotal)
	label := styles.Subtle(fmt.Sprintf(" cloning %d/%d", m.cloneDone+1, m.cloneTotal))
	return "  " + m.cloneProgress.ViewAs(pct) + label
}

func (m *ModelManager) renderToast() string {
	width := constants.WindowSize.Width
	if m.toast == "" {
		return styles.ToastIdleStyle.Width(width).Render("")
	}
	return styles.RenderToast(m.toast, m.toastLevel == events.ToastError, m.toastOpacity, width)
}

func (m *ModelManager) toastCmd(level events.ToastLevel, msg string) tea.Cmd {
	return func() tea.Msg {
		return events.Toast{Msg: msg, Level: level}
	}
}

// toastAnimMsg returns a tea.Tick callback that emits a ToastAnim for seq.
func toastAnimMsg(seq int) func(time.Time) tea.Msg {
	return func(time.Time) tea.Msg { return events.ToastAnim{Seq: seq} }
}

// toastTick schedules the next fade frame for seq.
func toastTick(seq int) tea.Cmd { return tea.Tick(toastFrame, toastAnimMsg(seq)) }
