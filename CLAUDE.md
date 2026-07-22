# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

lazymux is a Bubble Tea (charmbracelet) TUI git repo manager: clone repos, browse them in a filterable list sorted by recency, and jump into lazygit / your editor / a shell. It manages repo locations natively (no `ghq`), cloning into `~/lazymux/<namespace>/<repo>`, and links each repo to one or more registered git **forges** with a switchable primary.

Module path is `github.com/bkenks/lazymux` (kept GitHub-style so `go install` works), but the repo is hosted on a Forgejo instance at `fj.homektb.com`. Use `tea` (the Forgejo CLI) / `git` for VCS ops here, not `gh`.

## Commands

```bash
go build ./...            # compile
go test ./...             # run tests (config + repomgr only)
go vet ./...

mise run build            # → build/bin/lazymux
mise run dev              # → build/bin/lazymux-dev, sandboxed to ~/lazymux-dev
mise run install          # install lazymux to $GOBIN
mise run install-dev      # install lazymux-dev to $GOBIN
mise run clean            # remove build/bin
```

Build/install tasks live in `mise.toml` at the repo root (see
`.project/docs/build.md`); the Go toolchain is pinned there too.

Run a single test: `go test ./internal/repomgr -run TestRenderGitConfig -v`. Tests only exist for `internal/config` (save/load round-trip + legacy TOML migration) and `internal/repomgr` (`ParseRepoURL` table test + `RenderGitConfig` insteadOf behavior against a real temp git repo). There is no UI/event/command test coverage.

Version is injected at build time: `-ldflags "-X main.buildVersion=..."`.

## Architecture

Standard Elm architecture (Model/Update/View) with a custom router and event layer on top.

**Root model & routing** — `internal/app/model.go` defines `ModelManager`, which holds one instance of every screen plus an `active tea.Model` pointer. `Update()` does two-level dispatch: a type switch on `tea.Msg`, with a nested switch on `events.Event`. The central case is `events.SetState`, which sets `m.state` (a `domain.SessionState` enum) and reassigns `m.active` to the matching screen — this is the router. It also re-broadcasts the current `WindowSize` on every state change so the newly-active screen lays out at the right dimensions (notably the repo list, built at size 0 behind the splash). **After** the event switch, every message is also forwarded to the active screen (`m.active, cmd = m.active.Update(msg)`), so screens still receive raw key/resize messages. `View()` renders `m.active.View()` plus a single footer line (`FooterReservedLines`) that shows a live clone-progress bar while a clone batch runs, otherwise the toast.

**Screens** live under `internal/ui/*` (splash, repolist, confirm, clonerepos, forgeregistry, forgeselect, repoforges) plus `pkg/settings`. Each is a self-contained Bubble Tea model unaware of its siblings. Most are constructed up front in `New()`; `forgeselect` and `repoforges` are built lazily (nil pointers until first opened). `splash` is the initial `active` screen — a gradient wordmark + build version shown on launch (state `StateSplash`) that auto-dismisses to the repo list after ~1.6s or on any key while the first repo scan runs behind it.

**Events & commands** — the messaging pattern:
- `internal/events/*` — typed structs implementing the `events.Event` marker interface (`isEvent()`). Just data + a marker so `model.go` can do one `case events.Event:`.
- `internal/commands/*` — `func() tea.Cmd` factories returning closures that produce those events.
- Flow: **keypress → screen's Update matches a `constants.*KeyMap` binding → calls a `commands.XCmd()` → runtime runs the `tea.Cmd` (often `tea.ExecProcess`) → closure returns an `events.X{}` msg → `ModelManager.Update` mutates state and/or emits more Cmds.**

`internal/domain/*` holds cross-cutting types unrelated to messaging: `SessionState` enum, `Repo` (implements bubbles `list.Item`), and interaction-recency persistence (`interactions.go`, stored at `$XDG_DATA_HOME/lazymux/interactions.json`).

## The forge system (core domain concept)

A **forge** is a named git host (`config.Forge{Name, Host}`, e.g. `github`→`github.com`). Repos can live on multiple forges (e.g. self-hosted Forgejo mirrored to GitHub) with one designated **primary**.

The key trick (`internal/repomgr/git.go` → `RenderGitConfig`): every managed repo's `origin` is rewritten to a stable placeholder host (`placeholderHost`, default `lazymux-placeholder`) that **never changes**. A per-repo local git `insteadOf` rule resolves that placeholder to the primary forge. Switching primary re-renders only that one rule — the stored `origin` is never touched. `clearManagedInsteadOf` idempotently strips stale rules first. No automatic failover; the user always picks the live forge.

Three forge screens, each emitting a distinct event (`internal/events/forge.go`):
- `forgeregistry` (`F`) — global CRUD of all forges; blocks deleting a forge still in use.
- `forgeselect` — shown once per clone batch, one page per pending URL; check forges, pick primary + scheme.
- `repoforges` (`f`) — same UI for a single existing repo.

Per-repo links persist in `config.Repos map[string]RepoLink` (keyed by `namespace/name`); any primary/scheme change re-runs `RenderGitConfig`.

## internal/repomgr (native ghq replacement)

- `url.go` — `ParseRepoURL` handles https/http/ssh/scp-like (`git@host:ns/repo`) forms; `Key()` = `namespace/name`.
- `pending.go` — `PendingClone`: a parsed URL + auto-matched forge, built before the forge-select screen.
- `git.go` — `RepoDir` (path builder), `Clone`, `RenderGitConfig` (the placeholder/insteadOf logic), `List` (walks `BaseDir`, stops at first `.git`, annotates each `Repo` with forge link + last-interacted time), `Remove` (rm -rf + prune empty namespace dirs).

Actual clone **execution** is in `commands/clonerepoexec.go` via `tea.ExecProcess` (not inside repomgr) so git can prompt for credentials interactively.

## Config

`internal/config/config.go`. Single JSON file `~/lazymux/.lazymux.json` (override with `$LAZYMUX_CONFIG`); `Path()` resolves it, `BaseDir` defaults to `$HOME/lazymux`. `Load()` auto-migrates a legacy `~/.config/lazymux/config.toml` on first run and writes defaults otherwise; parse errors fall back to defaults with a `LoadWarning` surfaced as a startup toast. Schema: `Config{BaseDir, PlaceholderHost, Tools{Lazygit,Editor,Shell}, UI{Theme,ShowFullPath}, Behavior{DefaultProtocol,ConfirmDelete}, Forges, Repos}`.

## External tools

All go through `tea.ExecProcess` (suspends the renderer, hands over the terminal, resumes on exit), wrapped in a Cmd that emits a completion event carrying any error:
- `commands/teacmdbuilder.go` — generic `TeaCmdBuilder(name, args...)` → `events.CmdComplete`; used by lazygit.
- `commands/openinvscode.go` (editor), `commands/openshell.go` (shell; resolves `Tools.Shell` → `$SHELL` → `/bin/sh`), `commands/clonerepoexec.go` (git clone).
- Exception: `commands/pullallrepos.go` runs `git pull --ff-only` **headless** (no terminal) via an 8-way worker-pool semaphore, but *streams* results: `PullAllReposCmd` returns immediately with a `PullAllStarted{Total, Results chan}`; the repo list drains the channel via `WaitForPullCmd` (one `PullResult` at a time) to drive a live `bubbles/progress` bar + spinner, emitting the terminal `PullAllReposComplete` (→ refresh + summary toast) once drained.

## Conventions

- `internal/constants/globalvars.go` holds package-level mutable `WindowSize` (set once in `app/model.go`, read by each screen's own size helper) and `FooterReservedLines`.
- `internal/styles/{styles,themes}.go` use mutable package-level lipgloss vars; `themes.Apply(name)` (called once in `main.go`) overwrites the color vars and calls `rebuildStyles()`. Theming is global mutation, not per-render. `styles.Help` is a shared `bubbles/help` model (rebuilt on theme swap) that `domain.FormatBindingsInline` renders through, so the non-list screens' key hints match the help the list screens draw internally.
- `pkg/settings` is a **generic** reusable settings-list widget (no lazymux imports); `internal/app/settings.go` is the lazymux-specific glue mapping `SettingChanged` back onto `config.Config`.
- Errors and status surface in the footer via toast events rather than crashing the TUI. Toasts fade in → hold ~4s → fade out via a self-scheduling `events.ToastAnim` tick loop (seq-guarded against superseded toasts); `styles.RenderToast` blends the text color toward the terminal background per frame (`go-colorful`).
