# 01-02-00 — custom keybinds take the whole terminal

Supersedes the embedded terminal in 01-01-00. Keybinds as data, `ctrl+shift`
and `cmd` combos, and bubbletea v2 stand.

## Context

The embedded pane re-emulated the command's screen with `x/vt` and drew it
inside a border. Scrolling didn't work in Claude Code run that way, and every
TUI behaves best in the terminal it was written for, not an emulation of one.

## Decision

**A keybind's command gets the real terminal** through `tea.Exec`, the same
path `s` uses for a shell. lazymux draws nothing while it runs and sees no keys,
so there is no border, no header and no `ctrl+]`.

**A loading screen covers the handoff.** bubbletea leaves the alternate screen
before handing over, and a TUI command such as lazygit leaves its own on exit
before the process ends, so the main screen shows at both ends of the command.
lazymux clears the main screen and writes "loading…" there before starting the
command, then clears it again once the command ends. Drawing it on the
alternate screen instead didn't help on the way back: the command itself
switches away from it.

## Consequences

- A command that doesn't exit can't be left from lazymux; the user quits it the
  way they would in any terminal.
- `x/vt` and `xpty` are no longer dependencies.
- `ctrl+]` reaches the command again.
- The first keybind run clears what the shell had on screen before lazymux
  started, so it isn't there after quitting lazymux.
