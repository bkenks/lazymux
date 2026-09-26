package commands

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/gitkeeper/internal/config"
	"github.com/bkenks/gitkeeper/internal/events"
	"github.com/bkenks/gitkeeper/internal/styles"
	"github.com/charmbracelet/x/ansi"
)

// RunKeybindCmd hands the whole terminal to the keybind's command in dir. When
// the command ends gitkeeper comes back right away if the keybind returns on
// exit, and otherwise once enter is pressed, so its output can be read.
func RunKeybindCmd(bind config.Keybind, dir string) tea.Cmd {
	handoff := &keybindHandoff{
		cmd:         ShellCommand(bind.Command, dir),
		shouldPause: !bind.ReturnOnExit,
	}
	return tea.Exec(handoff, func(err error) tea.Msg { return events.CmdComplete{Err: err} })
}

// keybindHandoff runs a keybind's command over a loading screen on the main
// screen, which is what shows while gitkeeper hands over and when a command that
// used the alternate screen leaves it on exit. Otherwise the shell gitkeeper was
// started from flashes between the two.
type keybindHandoff struct {
	cmd         *exec.Cmd
	shouldPause bool
}

func (h *keybindHandoff) SetStdin(r io.Reader)  { h.cmd.Stdin = r }
func (h *keybindHandoff) SetStdout(w io.Writer) { h.cmd.Stdout = w }
func (h *keybindHandoff) SetStderr(w io.Writer) { h.cmd.Stderr = w }

func (h *keybindHandoff) Run() error {
	h.clearScreen()
	_, _ = fmt.Fprint(h.cmd.Stdout, styles.Subtle("loading…")+"\r\n")
	err := h.cmd.Run()
	if h.shouldPause {
		_, _ = fmt.Fprint(h.cmd.Stdout, "\n[command ended · press enter to return to gitkeeper]")
		_, _ = bufio.NewReader(h.cmd.Stdin).ReadString('\n')
	}
	h.clearScreen()
	return err
}

func (h *keybindHandoff) clearScreen() {
	_, _ = fmt.Fprint(h.cmd.Stdout, ansi.EraseEntireScreen+ansi.CursorHomePosition)
}
