package app

import (
	"fmt"
	"time"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/bkenks/lazymux/internal/ui/clonerepos"
	"github.com/bkenks/lazymux/internal/ui/confirm"
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

	active tea.Model

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
				}
				m.active = &m.confirmDelete

			case domain.StateCloneRepo:
				m.clonerepos = *clonerepos.New()
				m.active = &m.clonerepos

			case domain.StateSettings:
				m.active = &m.settingsModel
			}

		case events.StartRepoClone:
			m.clonerepos.RepoCounter = 0
			m.clonerepos.Failures = 0
			m.clonerepos.TotalRepos = len(msg.RepoUrls)
			cmds = append(cmds, commands.CloneReposExecCmd(msg.RepoUrls))

		case events.CloneRepoExec:
			if msg.Err != nil {
				m.clonerepos.Failures++
			}
			if m.clonerepos.RepoCounter < m.clonerepos.TotalRepos {
				m.clonerepos.RepoCounter++
			}
			if m.clonerepos.RepoCounter == m.clonerepos.TotalRepos {
				summary := fmt.Sprintf("cloned %d/%d", m.clonerepos.TotalRepos-m.clonerepos.Failures, m.clonerepos.TotalRepos)
				cmds = append(cmds,
					commands.RefreshReposCmd(),
					commands.SetState(domain.StateMain),
					m.toastCmd(events.ToastInfo, summary),
				)
			}

		case events.RepoDeleted:
			if msg.Err != nil {
				cmds = append(cmds, m.toastCmd(events.ToastError, fmt.Sprintf("delete failed: %v", msg.Err)))
			} else {
				cmds = append(cmds, m.toastCmd(events.ToastInfo, "repo deleted"))
			}
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.PullAllReposComplete:
			cmds = append(cmds, commands.RefreshReposCmd())

		case events.ReposRefreshed:
			m.main.UpdateRepoList(msg.RepoList)

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
