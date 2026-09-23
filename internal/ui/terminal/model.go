// Package terminal is the screen that runs a custom keybind's command in a
// pseudo-terminal and draws it inside a border, forwarding keys to it until
// the process exits or esc returns to the repo list.
package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/commands"
	"github.com/bkenks/lazymux/internal/constants"
	"github.com/bkenks/lazymux/internal/domain"
	"github.com/bkenks/lazymux/internal/styles"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

const (
	returnKey    = "esc"
	borderSize   = 2
	headerHeight = 1
)

type outputMsg struct{ model *Model }

type exitedMsg struct {
	model *Model
	err   error
}

type Model struct {
	title           string
	emulator        *vt.SafeEmulator
	pty             xpty.Pty
	cmd             *exec.Cmd
	output          chan struct{}
	exited          chan error
	isCursorVisible atomic.Bool
	hasExited       bool
	exitErr         error
}

// New starts cmd attached to a new pseudo-terminal sized to the screen.
func New(title string, cmd *exec.Cmd) (*Model, error) {
	width, height := innerSize()
	pty, err := xpty.NewPty(width, height)
	if err != nil {
		return nil, fmt.Errorf("open terminal for %s: %w", title, err)
	}
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	setSessionLeader(cmd)
	if err := pty.Start(cmd); err != nil {
		_ = pty.Close()
		return nil, fmt.Errorf("start %s: %w", title, err)
	}

	m := &Model{
		title:    title,
		emulator: vt.NewSafeEmulator(width, height),
		pty:      pty,
		cmd:      cmd,
		output:   make(chan struct{}, 1),
		exited:   make(chan error, 1),
	}
	m.isCursorVisible.Store(true)
	m.emulator.SetCallbacks(vt.Callbacks{CursorVisibility: m.isCursorVisible.Store})

	go m.copyOutput()
	go m.forwardInput()
	go func() { m.exited <- xpty.WaitProcess(context.Background(), cmd) }()
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.waitForOutput(), m.waitForExit())
}

// copyOutput feeds process output to the emulator until the pty closes, then
// closes the emulator's input pipe to end forwardInput. The pipe is closed
// rather than the emulator because Emulator.Close isn't safe alongside Read.
func (m *Model) copyOutput() {
	defer func() {
		if pipe, ok := m.emulator.InputPipe().(io.Closer); ok {
			_ = pipe.Close()
		}
	}()
	buf := make([]byte, 32*1024)
	for {
		n, err := m.pty.Read(buf)
		if n > 0 {
			_, _ = m.emulator.Write(buf[:n])
			select {
			case m.output <- struct{}{}:
			default:
			}
		}
		if err != nil {
			return
		}
	}
}

func (m *Model) forwardInput() {
	_, _ = io.Copy(m.pty, m.emulator)
}

func (m *Model) waitForOutput() tea.Cmd {
	return func() tea.Msg {
		<-m.output
		return outputMsg{model: m}
	}
}

func (m *Model) waitForExit() tea.Cmd {
	return func() tea.Msg {
		return exitedMsg{model: m, err: <-m.exited}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width, height := innerSize()
		m.emulator.Resize(width, height)
		_ = m.pty.Resize(width, height)

	case outputMsg:
		if msg.model == m {
			return m, m.waitForOutput()
		}

	case exitedMsg:
		if msg.model == m {
			m.hasExited = true
			m.exitErr = msg.err
			m.close()
		}

	case tea.KeyPressMsg:
		if msg.String() == returnKey {
			m.close()
			return m, returnToRepoList()
		}
		if !m.hasExited {
			m.emulator.SendKey(vt.KeyPressEvent(msg))
		}

	case tea.PasteMsg:
		if !m.hasExited {
			m.emulator.Paste(msg.Content)
		}
	}
	return m, nil
}

// close ends the process and releases the pty, which also ends copyOutput and
// with it the emulator. Closing the pty hangs up the process group; the kill
// covers a child that ignores SIGHUP.
func (m *Model) close() {
	_ = m.pty.Close()
	if m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
}

func returnToRepoList() tea.Cmd {
	return tea.Batch(
		commands.SetState(domain.StateMain),
		commands.RefreshReposCmd(),
	)
}

func (m *Model) View() tea.View {
	width, height := innerSize()
	frame := styles.TerminalFrameStyle.
		Width(width + borderSize).
		Height(height + borderSize).
		Render(m.emulator.Render())

	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, m.header(), frame))
	if !m.hasExited && m.isCursorVisible.Load() {
		position := m.emulator.CursorPosition()
		view.Cursor = tea.NewCursor(position.X+borderSize/2, position.Y+headerHeight+borderSize/2)
	}
	return view
}

func (m *Model) header() string {
	parts := []string{styles.Subtle(m.title)}
	switch {
	case m.exitErr != nil:
		parts = append(parts, styles.ToastErrorStyle.Render("exited: "+m.exitErr.Error()))
	case m.hasExited:
		parts = append(parts, styles.Subtle("exited"))
	}
	parts = append(parts, styles.Subtle(returnKey+" to return"))
	return strings.Join(parts, styles.Subtle(" · "))
}

// innerSize is the area inside the border available to the process.
func innerSize() (width, height int) {
	x, y := styles.DocStyle.GetFrameSize()
	width = max(constants.WindowSize.Width-x-borderSize, 1)
	reservedHeight := y + constants.FooterReservedLines + borderSize + headerHeight
	height = max(constants.WindowSize.Height-reservedHeight, 1)
	return width, height
}
