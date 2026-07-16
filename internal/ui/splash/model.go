// Package splash is the brief gradient intro shown at launch. It displays the
// lazymux wordmark and build version, then auto-dismisses to the repo list (or
// on any keypress) while the initial repo scan runs behind it.
package splash

import (
	"strings"
	"time"

	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const splashDuration = 1600 * time.Millisecond

// gradientFrom/To bound the wordmark's horizontal color sweep.
var (
	gradientFrom, _ = colorful.Hex("#AD58B4")
	gradientTo, _   = colorful.Hex("#EE6FF8")
)

type Model struct {
	version string
}

func New(version string) *Model { return &Model{version: version} }

func (m *Model) Init() tea.Cmd {
	return tea.Tick(splashDuration, func(time.Time) tea.Msg {
		return dismissMsg{}
	})
}

// dismissMsg ends the splash after its timer elapses.
type dismissMsg struct{}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg, dismissMsg:
		return m, commands.SetState(domain.StateMain)
	}
	return m, nil
}

func (m *Model) View() string {
	wordmark := gradient("lazymux", gradientFrom, gradientTo)
	inner := lipgloss.JoinVertical(
		lipgloss.Center,
		wordmark,
		"",
		styles.Subtle("a TUI git repo manager"),
		styles.Subtle(m.version),
	)
	box := styles.DialogStyle.Render(inner)

	return lipgloss.Place(
		constants.WindowSize.Width,
		constants.WindowSize.Height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// gradient renders s with a per-character color sweep from → to.
func gradient(s string, from, to colorful.Color) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		t := 0.0
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}
		c := from.BlendLab(to, t).Clamped()
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Hex())).
			Bold(true).
			Render(string(r)))
	}
	return b.String()
}
