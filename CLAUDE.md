# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is lazymux

A TUI application combining `ghq` (repository manager) and `lazygit` (git TUI) into a unified GitHub client for the terminal. Users browse repos managed by `ghq`, then launch `lazygit` directly into the selected repo.

**Runtime dependencies:** `ghq` and `lazygit` must be installed and available in `$PATH`.

## Commands

```bash
# Run locally (builds to ./testing/ and installs to ~/.local/bin)
./devinstall.sh

# Build release binaries for all platforms (output to .builds/)
./build.sh

# Run directly without installing
go run .

# Build for current platform
go build -o ./testing/lazymux
```

There are no automated tests in this project.

## Architecture

The app uses the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework (elm-inspired MVU pattern). The entry point (`main.go`) creates a `ModelManager` and starts the Bubble Tea program.

### Key patterns

**ModelManager** (`internal/app/model.go`) — Top-level `tea.Model` that owns all sub-models and routes messages. It tracks `SessionState` and swaps the `active tea.Model` when state changes. All sub-model updates are delegated to `m.active.Update(msg)`.

**Events** (`internal/events/`) — Custom `tea.Msg` types that implement the `Event` interface. Commands return these as messages to trigger state changes, repo refreshes, etc. New events need both a struct in `events/` and handling in `app/model.go`.

**Commands** (`internal/commands/`) — Functions that return `tea.Cmd`. They wrap shell calls (`ghq list`, `ghq get`, `lazygit`, `code`/`codium`) using `tea.ExecProcess` or return event messages directly. `TeaCmdBuilder` is the generic helper for shelling out.

**State machine** — Three states defined in `internal/domain/sessionstate.go`: `StateMain` (repo list), `StateConfirmDelete` (delete dialog), `StateCloneRepo` (clone textarea). State transitions go through `commands.SetState()` → `events.SetState` → `ModelManager.Update`.

**Key bindings** (`internal/constants/bindings.go`) — Each view has its own `keyMap` struct with `HelpBinds()` for Bubble Tea's help component. Add new bindings here and reference them in the corresponding view's `Update`.

### Sub-models (views)

| Package | State | Description |
|---|---|---|
| `internal/ui/repolist` | `StateMain` | Filterable list of repos from `ghq list` |
| `internal/ui/confirm` | `StateConfirmDelete` | Confirmation dialog before `ghq remove` |
| `internal/ui/clonerepos` | `StateCloneRepo` | Textarea to paste repo URLs for `ghq get` |

### Adding a new feature

1. Define an event in `internal/events/`
2. Define a command in `internal/commands/` that emits the event
3. Handle the event in `internal/app/model.go`
4. If a new view is needed, add it under `internal/ui/`, add a new `SessionState`, and wire it into `ModelManager`
