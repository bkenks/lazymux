package app

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/gitkeeper/internal/commands"
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/constants"
	"github.com/bkenks/gitkeeper/internal/domain"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/keybind"
	"github.com/bkenks/gitkeeper/internal/repomgr"
	"github.com/bkenks/gitkeeper/internal/semver"
	"github.com/bkenks/gitkeeper/internal/styles"
	"github.com/bkenks/gitkeeper/internal/ui/clonerepos"
	"github.com/bkenks/gitkeeper/internal/ui/confirm"
	"github.com/bkenks/gitkeeper/internal/ui/forgeregistry"
	"github.com/bkenks/gitkeeper/internal/ui/forgeselect"
	"github.com/bkenks/gitkeeper/internal/ui/keybinds"
	"github.com/bkenks/gitkeeper/internal/ui/repolist"
	"github.com/bkenks/gitkeeper/internal/ui/reposdir"
	"github.com/bkenks/gitkeeper/internal/ui/reposettings"
	"github.com/bkenks/gitkeeper/internal/ui/settings"
	"github.com/bkenks/gitkeeper/internal/ui/splash"
	"github.com/bkenks/gitkeeper/internal/ui/tagrelease"
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
	settingsModel *settings.Model
	forgeSelect   *forgeselect.Model
	forgeRegistry *forgeregistry.Model
	repoSettings  *reposettings.Model
	tagRelease    *tagrelease.Model
	keybinds      *keybinds.Model
	reposDir      *reposdir.Model

	active tea.Model

	// clone batch progress (cloning runs after the forge-select step)
	cloneTotal    int
	cloneDone     int
	cloneFail     int
	cloneProgress progress.Model

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
	domain.ShowStats = cfg.UI.ShowStats
	domain.ShowFullPath = cfg.UI.ShowFullPath
	domain.Sort = domain.ParseSortMode(cfg.UI.SortMode)

	m := &ModelManager{
		cfg:           cfg,
		splash:        *splash.New(version),
		main:          *repolist.New(),
		confirmDelete: *confirm.New(),
		clonerepos:    *clonerepos.New(cfg),
		cloneProgress: styles.NewProgress(),
	}

	m.main.SetKeybinds(cfg.Keybinds)
	m.state = domain.StateSplash
	m.active = &m.splash
	return m
}

func (m *ModelManager) Init() tea.Cmd {
	cmds := []tea.Cmd{m.splash.Init()}
	if m.cfg.ValidateRepoRoot() == nil {
		cmds = append(cmds, commands.RefreshReposCmd())
	}
	warnings := append(slices.Clone(m.cfg.Warnings), m.keybindClashes()...)
	if len(warnings) > 0 {
		cmds = append(cmds, m.toastCmd(events.ToastError, strings.Join(warnings, "; ")))
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

			case domain.StateMain, domain.StateReposDir:
				cmds = append(cmds, m.showMainOrReposDir())

			case domain.StateConfirmDelete:
				repo, ok := m.main.List.SelectedItem().(domain.Repo)
				if !m.cfg.Behavior.ConfirmDelete {
					m.state = domain.StateMain
					m.active = &m.main
					if ok {
						cmds = append(cmds, commands.DeleteRepoCmd(repo.Path, repo.AbsPath))
					}
					break
				}
				m.confirmDelete = *confirm.New()
				if ok {
					m.confirmDelete.RepoPath = repo.Path
					m.confirmDelete.AbsPath = repo.AbsPath
				}
				m.active = &m.confirmDelete

			case domain.StateCloneRepo:
				m.clonerepos = *clonerepos.New(m.cfg)
				m.active = &m.clonerepos

			case domain.StateSettings:
				m.settingsModel = settings.New(m.cfg)
				m.active = m.settingsModel
				cmds = append(cmds, m.settingsModel.Init())

			case domain.StateForgeSelect:
				// m.forgeSelect is built in the StartRepoClone handler.
				if m.forgeSelect != nil {
					m.active = m.forgeSelect
				}

			case domain.StateForgeRegistry:
				m.forgeRegistry = forgeregistry.New(m.cfg)
				m.active = m.forgeRegistry

			case domain.StateRepoSettings:
				// m.repoSettings is built in the OpenRepoSettings handler.
				if m.repoSettings != nil {
					m.active = m.repoSettings
					cmds = append(cmds, m.repoSettings.Init())
				}

			case domain.StateTagRelease:
				// m.tagRelease is built in the OpenTagRelease handler.
				if m.tagRelease != nil {
					m.active = m.tagRelease
					cmds = append(cmds, m.tagRelease.Init())
				}

			case domain.StateKeybinds:
				m.keybinds = keybinds.New(m.cfg, m.main.ReservedKeys())
				m.active = m.keybinds
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
			seen := map[string]bool{}
			for _, raw := range msg.RepoUrls {
				if strings.TrimSpace(raw) == "" {
					continue
				}
				p, err := repomgr.NewPendingClone(m.cfg, raw)
				if err != nil {
					bad++
					continue
				}
				if seen[p.URL.Key()] {
					continue // the same repo pasted twice would clone into one directory
				}
				seen[p.URL.Key()] = true
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

		case events.NamespaceCloneFailed:
			// The clone screen (still active) shows this inline too; the
			// toast just surfaces it if the user has already moved on.
			cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("%s: %v", msg.Namespace, msg.Err)))

		case events.ForgeSelectComplete:
			// Persist any inline-added forges, then start the clones.
			if len(msg.NewForges) > 0 {
				cmds = append(cmds, m.saveConfig("forges", func(c *config.Config) {
					c.Forges = append(c.Forges, msg.NewForges...)
				}))
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
			} else if msg.Clone.Origin != "" {
				// Clone succeeded: record the link and rewrite the repo to a
				// placeholder origin resolved to its origin forge, plus a push
				// URL per upstream.
				key := msg.Clone.URL.Key()
				link := msg.Clone.Link()
				cmds = append(cmds,
					m.saveConfig("repo link", func(c *config.Config) {
						c.Repos[key] = c.Repos[key].WithForgeLinks(link)
					}),
					m.renderGitConfig(key, link),
				)
			}
			if m.cloneDone < m.cloneTotal {
				m.cloneDone++
			}
			if m.cloneDone == m.cloneTotal {
				summary := fmt.Sprintf("cloned %d/%d", m.cloneTotal-m.cloneFail, m.cloneTotal)
				cmds = append(cmds,
					commands.RefreshReposCmd(),
					m.toastCmd(events.ToastInfo, summary),
				)
			}

		case events.ForgesChanged:
			prev := m.cfg.Clone()
			cmds = append(cmds, m.saveConfig("forges", func(c *config.Config) {
				c.Forges = msg.Forges
				for key, link := range msg.Repos {
					c.Repos[key] = c.Repos[key].WithForgeLinks(link)
				}
			}))

			// Re-render the git remotes for any repo whose links or forge hosts
			// changed (promotion after a delete, a rename, or a host edit). A
			// repo left unlinked keeps its existing remote — the insteadOf rule
			// already written still resolves.
			for key, link := range m.cfg.Repos {
				if _, ok := m.cfg.ForgeByName(link.Origin); !ok {
					continue // unlinked or dangling origin; leave the existing config alone
				}
				if slices.Equal(remoteHosts(prev, prev.Repos[key]), remoteHosts(m.cfg, link)) {
					continue
				}
				cmds = append(cmds, m.renderGitConfig(key, link))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.KeybindsChanged:
			cmds = append(cmds, m.saveKeybinds(msg.Keybinds))

		case events.RunKeybind:
			cmds = append(cmds, commands.RunKeybindCmd(msg.Keybind, msg.Dir))

		case events.OpenRepoSettings:
			m.repoSettings = reposettings.New(m.cfg, msg.Key, msg.Tags, msg.TagsErr)
			cmds = append(cmds, commands.SetState(domain.StateRepoSettings))

		case events.RepoSettingsChanged:
			cmds = append(cmds, m.saveConfig("repo settings", func(c *config.Config) {
				if msg.Link.IsEmpty() {
					delete(c.Repos, msg.Key)
				} else {
					c.Repos[msg.Key] = msg.Link
				}
			}))
			if msg.Link.Origin != "" {
				cmds = append(cmds, m.renderGitConfig(msg.Key, msg.Link))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.OpenTagRelease:
			link := m.cfg.Repos[msg.Key]
			format := semver.Format{Prefix: link.TagPrefix, Suffix: link.TagSuffix}
			m.tagRelease = tagrelease.New(msg.Key, msg.Dir, format, msg.Tags)
			cmds = append(cmds, commands.SetState(domain.StateTagRelease))

		case events.TagReleased:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, msg.Err.Error()))
			} else {
				cmds = append(cmds, m.toastCmd(events.ToastInfo, "pushed "+msg.Tag))
			}

		case events.RepoDeleted:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("delete failed: %v", msg.Err)))
			} else {
				cmds = append(cmds,
					m.saveConfig("config", func(c *config.Config) { delete(c.Repos, msg.Key) }),
					m.toastCmd(events.ToastInfo, "repo deleted"),
				)
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.PullAllReposComplete:
			cmds = append(cmds,
				commands.RefreshReposCmd(),
				m.toastCmd(events.ToastInfo, msg.Summary()),
			)

		case events.ReposRefreshed:
			cmds = append(cmds, m.main.UpdateRepoList(msg.RepoList))

		case events.SortModeChanged:
			setMode := func(c *config.Config) { c.UI.SortMode = string(msg.Mode) }
			if cmd := m.saveConfig("sort order", setMode); cmd != nil {
				cmds = append(cmds, cmd)
			} else {
				cmds = append(cmds, m.toastCmd(events.ToastInfo, "sorted by "+msg.Mode.Label()))
			}

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
		case events.SettingsChanged:
			cmds = append(cmds, m.applySettings(msg.Config))
		case events.ReposDirChosen:
			cmds = append(cmds, m.applyReposDir(msg.Dir))
		}
	}

	var cmd tea.Cmd
	if isPullProgress(msg) {
		_, cmd = m.main.Update(msg)
	} else {
		m.active, cmd = m.active.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// isPullProgress reports whether msg drives the repo list's pull-all progress.
// Those always go to the repo list, so a pull keeps going while another
// screen is open.
func isPullProgress(msg tea.Msg) bool {
	switch msg.(type) {
	case events.PullAllStarted, events.PullResult, events.PullAllDrained, spinner.TickMsg:
		return true
	}
	return false
}

func (m *ModelManager) View() tea.View {
	active := m.active.View()
	body := styles.DocStyle.Render(active.Content)
	// The footer region is a single reserved line (FooterReservedLines). A clone
	// batch in flight owns it — showing a live gradient bar between the per-repo
	// terminal handovers — otherwise it's the toast line.
	footer := m.renderToast()
	if line := m.renderCloneProgress(); line != "" {
		footer = line
	}
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, body, footer))
	v.AltScreen = true
	if active.Cursor != nil {
		v.Cursor = active.Cursor
		v.Cursor.X += styles.DocStyle.GetMarginLeft()
		v.Cursor.Y += styles.DocStyle.GetMarginTop()
	}
	return v
}

func (m *ModelManager) saveKeybinds(keybinds []config.Keybind) tea.Cmd {
	m.main.SetKeybinds(keybinds)
	setKeybinds := func(c *config.Config) { c.Keybinds = keybinds }
	if cmd := m.saveConfig("keybinds", setKeybinds); cmd != nil {
		return cmd
	}
	return m.toastCmd(events.ToastInfo, "keybinds saved")
}

// saveConfig applies change to the config file, re-read first so edits another
// process made meanwhile survive, and adopts the result as the app's config. If
// the write fails, change is still applied in memory for this session and the
// returned command toasts the error, naming what couldn't be saved. It returns
// nil on success.
func (m *ModelManager) saveConfig(what string, change func(*config.Config)) tea.Cmd {
	saved, err := config.Update(change)
	if err != nil {
		change(&m.cfg)
		commands.SetDeps(m.cfg)
		return m.toastCmd(events.ToastError, fmt.Sprintf("couldn't save %s: %v", what, err))
	}
	m.cfg = saved
	commands.SetDeps(m.cfg)
	return nil
}

// renderGitConfig rewrites a repo's remotes to match link, returning an error
// toast command if that fails, or nil.
func (m *ModelManager) renderGitConfig(key string, link config.RepoLink) tea.Cmd {
	if err := repomgr.RenderGitConfig(m.cfg, key, link); err != nil {
		return m.toastCmd(events.ToastError, fmt.Sprintf("%s: %v", key, err))
	}
	return nil
}

// remoteHosts lists the host each of the link's forges resolves to in cfg,
// origin first, so a change to any of them shows up as a difference. A forge
// missing from the registry resolves to "".
func remoteHosts(cfg config.Config, link config.RepoLink) []string {
	names := append([]string{link.Origin}, link.Upstreams...)
	hosts := make([]string, len(names))
	for i, name := range names {
		if forge, ok := cfg.ForgeByName(name); ok {
			hosts[i] = forge.Host
		}
	}
	return hosts
}

// keybindClashes describes each custom keybind whose key the repo list already
// uses, or that another keybind took first. Such a keybind never runs, which
// the keybinds screen prevents but a hand-edited config doesn't.
func (m *ModelManager) keybindClashes() []string {
	var clashes []string
	reserved := m.main.ReservedKeys()
	claimed := map[string]string{}
	for _, bind := range m.cfg.Keybinds {
		if used, ok := keybind.FindClash(bind.Keys, reserved); ok {
			clashes = append(clashes,
				fmt.Sprintf("keybind %q won't run: %s is a gitkeeper key", bind.Name, used))
		} else if first, ok := claimed[bind.Keys]; ok {
			clashes = append(clashes,
				fmt.Sprintf("keybind %q won't run: %s is already bound to %q",
					bind.Name, bind.Keys, first))
		} else {
			claimed[bind.Keys] = bind.Name
		}
	}
	return clashes
}

// renderCloneProgress draws a gradient bar while a clone batch is in flight.
// Empty once every repo is done, so the final summary toast can take the footer.
func (m *ModelManager) renderCloneProgress() string {
	if m.cloneTotal == 0 || m.cloneDone >= m.cloneTotal {
		return ""
	}
	m.cloneProgress.SetWidth(styles.ProgressBarWidth(constants.WindowSize.Width))
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
