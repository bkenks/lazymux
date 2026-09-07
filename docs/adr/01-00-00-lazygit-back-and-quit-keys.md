# 01-00-00 — lazygit back and quit keys

## Context

lazymux runs lazygit as a subprocess via `tea.ExecProcess`. The goal was one
consistent key model across both programs: `esc` always goes back, `q` always
quits.

`esc` is already back on every lazymux screen. Inside lazygit it pops panels but
does nothing at the top level, so the only way out of lazygit was `q` — a quit
key doing a back key's job.

## Decision

`esc` leaves lazygit, controlled by the `Tools.LazygitEscQuit` setting ("Esc
quits lazygit", default on).

lazymux writes its own one-line config containing `quitOnTopLevelReturn: true`
to `$XDG_DATA_HOME/lazymux/lazygit-esc-quit.yml` and passes
`--use-config-file "<the user's configs>,<the overlay>"`. lazygit merges that
list left to right, so the overlay wins without the user's own config being read
by lazymux, written to, or backed up. Turning the setting off stops passing the
flag; nothing needs restoring.

The user's configs come from `LG_CONFIG_FILE` when set, otherwise `config.yml`
in the directory `lazygit --print-config-dir` reports. Paths that do not exist
are dropped, because lazygit refuses to start when a config file in the list is
missing.

`q` quits lazygit and returns to the repo list. A second `q` quits lazymux —
`GlobalKeyMap.Quit` binds `q` and `ctrl+c` on every lazymux screen that is not
accepting text input.

## Alternatives rejected

**Editing the user's lazygit config, with a backup to restore.** Requires
backup/restore machinery, and leaves the user's config wrong if lazymux is
killed between write and restore. The overlay merge gets the same result with
nothing of theirs touched.

**Rebinding `keybinding.universal.quit` to `<esc>`.** Collides with
`keybinding.universal.return`, which is `<esc>` and is what makes esc pop panels.
`quitOnTopLevelReturn` is lazygit's own knob for exactly this and leaves the
panel behavior intact.

**A single `q` that quits lazygit and lazymux together.** Not implementable.
lazymux only observes lazygit's exit, and lazygit emits nothing that
distinguishes which key caused it:

- `quitOnTopLevelReturn` esc calls the same `Quit()` action as `q`
  (`pkg/gui/controllers/quit_actions.go`), so the two paths are identical from
  the outside.
- `LAZYGIT_NEW_DIR_FILE` looked like a discriminator, since `Quit()` writes
  `os.Getwd()` while `QuitWithoutChangingDirectory()` writes `gui.InitialDir`
  (`pkg/gui/gui.go`). But `InitialDir` comes from `os.Getwd()` in
  `pkg/app/app.go`, evaluated *after* `-p` has already chdir'd into the repo, so
  both write the same string.
- The exit code is 0 either way.

The only remaining route was a `customCommand` bound to `q` that writes a
sentinel file and kills lazygit's own process. lazygit installs no SIGTERM
handler (`pkg/gui/controllers/helpers/signal_handling.go` traps only SIGCONT),
so it would die without running its shutdown. Not worth it to save one keypress:
`q` `q` already quits both.

Checked against lazygit 0.64.1.
