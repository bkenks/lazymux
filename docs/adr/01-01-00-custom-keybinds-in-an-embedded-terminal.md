# 01-01-00 — custom keybinds in an embedded terminal

Supersedes 01-00-00. The built-in lazygit (`tab`) and Claude Code (`c`, `a`)
launchers are gone; lazygit and anything else is now a custom keybind the user
sets up.

## Context

Hard-coding each tool meant a new key, command, and setting per tool, and the
lazygit esc overlay in 01-00-00 existed only because lazymux handed the whole
terminal to lazygit with `tea.ExecProcess` and could not see its keys.

The ask: bind any key combo to any shell command, run it in the selected repo,
draw it inside a border with an "esc to return" hint, and keep it interactive.
Combos include `ctrl+shift+<key>` and `cmd`.

## Decision

**Keybinds are data.** `.lazymux.json` holds `keybinds: [{name, keys, command}]`,
edited on a screen opened with `2` (settings moved to `1`; the settings-style
screens sit on number keys). `keys` is typed as text
(`ctrl + g`) and stored in the canonical form bubbletea reports for the key
press (`ctrl+g`). A combo is refused when it doesn't parse, when the repo list
already uses it (derived from the live key maps, not a hand-kept list), or when
another custom keybind has it.

**Commands run in an embedded pseudo-terminal**, not `tea.ExecProcess`.
`charmbracelet/x/xpty` opens the pty (ConPTY on Windows) and
`charmbracelet/x/vt` emulates the screen, which lazymux renders inside a border.
lazymux sees every key first: `ctrl+]` returns, everything else is forwarded. On
Unix the child gets its own session with the pty as controlling terminal, since
xpty doesn't set that up and TUIs need it.

A command that exits stays on screen, marked exited, until `ctrl+]`. Otherwise
a short command such as `git status` would vanish before it could be read.

**`ctrl+]` returns, not `esc`.** lazygit and vim need `esc` themselves.
`ctrl+]` is the traditional escape key for a session (telnet), few TUIs bind
it, and every terminal sends it as one distinct byte. `ctrl+esc` was
considered: terminals without the kitty keyboard protocol send it as plain
`esc`, which would leave no way out of a command that doesn't exit.

**bubbletea v2.** v1 decodes legacy key sequences only: `ctrl+shift+k` arrives
as `ctrl+k` and `cmd` never arrives. v2 always asks the terminal for kitty
keyboard disambiguation, so both are reported in terminals that support it.
`x/vt` is also built on v2's key types.

## Consequences

- `ctrl+]` belongs to lazymux inside the pane, so a tool that binds it (vim's
  jump-to-tag) can't receive it.
- `ctrl+shift` and `cmd` combos only work in terminals with the kitty keyboard
  protocol.
- `x/vt` has no tagged release; it is pinned to a pseudo-version.
- The overlay 01-00-00 wrote to `$XDG_DATA_HOME/lazymux/lazygit-esc-quit.yml` is
  no longer read; lazymux leaves any existing copy in place.

## Rejected

**`tea.ExecProcess`, as before.** It gives the child the real terminal, so
lazymux can draw no border or hint and can't catch a return key.

**Staying on bubbletea v1.** It can't tell `ctrl+shift+k` from `ctrl+k`, so
custom keybinds couldn't use `ctrl+shift` or `cmd` combos.
