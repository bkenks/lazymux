package repolist

import (
	"fmt"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/events"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	List     list.Model
	RepoList []list.Item

	// pull-all progress: streamed one PullResult at a time off pullCh while a
	// progress bar + spinner render below the list.
	pulling   bool
	pullCh    <-chan events.PullResult
	pullDone  int
	pullTotal int
	pulled    int
	skipped   []events.SkippedPull
	progress  progress.Model
	spinner   spinner.Model
}

// newDelegate builds the list row renderer, sized for whether the "forge:"
// line is currently shown: height 3 fits name + namespace + forge line, height
// 2 when the forge label is hidden.
func newDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	if domain.ShowForge {
		d.SetHeight(3)
	} else {
		d.SetHeight(2)
	}
	return d
}

func New() *Model {
	w, h := SizeBuffer()
	newList := list.New(
		[]list.Item{},
		newDelegate(),
		w, h,
	)
	newList.Title = "Repositories"
	newList.AdditionalShortHelpKeys = constants.RepoListKeyMap.HelpBinds(constants.Short)
	newList.AdditionalFullHelpKeys = constants.RepoListKeyMap.HelpBinds(constants.Full)

	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(styles.Purple)

	m := &Model{
		List:     newList,
		progress: progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
		spinner:  sp,
	}
	m.applySize()
	return m
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.applySize()

	case tea.KeyMsg:
		// Suppress repo-action keybinds while the list's filter input has focus —
		// otherwise typing "r" in a search filter triggers a refresh.
		if m.List.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, constants.RepoListKeyMap.Select):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.LazygitCmd(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.Clone):
			cmds = append(cmds, commands.SetState(domain.StateCloneRepo))

		case key.Matches(msg, constants.RepoListKeyMap.Delete):
			if ConvertToRepoType(m.List.SelectedItem()).AbsPath == "" {
				break
			}
			cmds = append(cmds, commands.SetState(domain.StateConfirmDelete))

		case key.Matches(msg, constants.RepoListKeyMap.VSCode):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.OpenInVSCode(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.Settings):
			cmds = append(cmds, commands.SetState(domain.StateSettings))

		case key.Matches(msg, constants.RepoListKeyMap.Refresh):
			cmds = append(cmds, commands.RefreshReposCmd())

		case key.Matches(msg, constants.RepoListKeyMap.CopyPath):
			cmds = append(cmds, commands.CopyPathCmd(AbsRepoPath(m.List.SelectedItem())))

		case key.Matches(msg, constants.RepoListKeyMap.Shell):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.AbsPath == "" {
				break
			}
			domain.SaveInteraction(repo.Path)
			cmds = append(cmds, commands.OpenShellCmd(repo.AbsPath))

		case key.Matches(msg, constants.RepoListKeyMap.PullAll):
			if m.pulling {
				break // already pulling; ignore repeat presses
			}
			m.pulling = true
			m.pullDone, m.pullTotal, m.pulled = 0, 0, 0
			m.skipped = nil
			m.applySize()
			cmds = append(cmds, commands.PullAllReposCmd(), m.spinner.Tick)

		case key.Matches(msg, constants.RepoListKeyMap.Forges):
			repo := ConvertToRepoType(m.List.SelectedItem())
			if repo.Path == "" {
				break
			}
			cmds = append(cmds, commands.OpenRepoForgesCmd(repo.Path))

		case key.Matches(msg, constants.RepoListKeyMap.Registry):
			cmds = append(cmds, commands.SetState(domain.StateForgeRegistry))

		case key.Matches(msg, constants.RepoListKeyMap.ToggleForge):
			domain.ShowForge = !domain.ShowForge
			m.List.SetDelegate(newDelegate())

		case key.Matches(msg, constants.RepoListKeyMap.ToggleStats):
			domain.ShowStats = !domain.ShowStats
		}

	case events.PullAllStarted:
		m.pullTotal = msg.Total
		m.pullCh = msg.Results
		cmds = append(cmds, commands.WaitForPullCmd(m.pullCh))

	case events.PullResult:
		m.pullDone++
		if msg.Reason == "" {
			m.pulled++
		} else {
			m.skipped = append(m.skipped, events.SkippedPull{RepoPath: msg.RepoPath, Reason: msg.Reason})
		}
		cmds = append(cmds, commands.WaitForPullCmd(m.pullCh))

	case events.PullAllDrained:
		pulled, skipped := m.pulled, m.skipped
		m.pulling = false
		m.pullCh = nil
		m.applySize()
		cmds = append(cmds, func() tea.Msg {
			return events.PullAllReposComplete{Pulled: pulled, Skipped: skipped}
		})

	case spinner.TickMsg:
		if m.pulling {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.pulling {
		return lipgloss.JoinVertical(lipgloss.Left, m.List.View(), m.pullView())
	}
	return m.List.View()
}

// pullView renders the live pull-all progress line: spinner, gradient bar, and
// a running count of repos processed.
func (m *Model) pullView() string {
	pct := 0.0
	if m.pullTotal > 0 {
		pct = float64(m.pullDone) / float64(m.pullTotal)
	}
	label := styles.Subtle(fmt.Sprintf(" %d/%d pulled", m.pullDone, m.pullTotal))
	return "  " + m.spinner.View() + " " + m.progress.ViewAs(pct) + label
}

// applySize lays out the list, leaving a row for the pull-progress line while a
// pull-all is running, and fits the progress bar to the available width.
func (m *Model) applySize() {
	w, h := SizeBuffer()
	if m.pulling {
		if h--; h < 1 {
			h = 1
		}
	}
	m.List.SetSize(w, h)

	bar := w - 24 // leave room for the spinner glyph and count label
	switch {
	case bar > 40:
		bar = 40
	case bar < 10:
		bar = 10
	}
	m.progress.Width = bar
}

// SyncForgeVisibility re-applies the row delegate so a change in
// domain.ShowForge (made outside the list, e.g. from settings) takes effect on
// the row height next render.
func (m *Model) SyncForgeVisibility() { m.List.SetDelegate(newDelegate()) }

// UpdateRepoList replaces the list's items. It returns the command from
// list.SetItems, which must be run so the filtered view is recomputed when a
// filter is applied — dropping it leaves the filter phrase set but the results
// empty ("0 results found") after a refresh (e.g. returning from lazygit).
func (m *Model) UpdateRepoList(repoList []list.Item) tea.Cmd {
	return m.List.SetItems(repoList)
}
