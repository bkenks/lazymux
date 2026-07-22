# Building lazymux

Build/install tasks live in `mise.toml` at the repo root. Run them with
`mise run <task>` from anywhere in the repo; `mise tasks` lists them.

The Go toolchain itself is pinned in `mise.toml` (`[tools] go = "1.26"`), so
`mise install` provisions it — no separate Go install needed.

## Commands

| Command                | Output                        |
|------------------------|--------------------------------|
| `mise run build`       | `build/bin/lazymux`            |
| `mise run dev`         | `build/bin/lazymux-dev`        |
| `mise run install`     | installs `lazymux` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `mise run install-dev` | installs `lazymux-dev` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `mise run clean`       | removes `build/bin`            |

`install` and `install-dev` depend on `build` / `dev`, so they compile first.

## Regular build vs. dev build

Both tasks compile the same source. The only difference is a build-time flag
(`-ldflags -X`) that overrides `internal/config.dirName`, which controls the
directory under `$HOME` used for the config file and the default repo `BaseDir`:

- **`build`** — `dirName` stays `lazymux`, so the binary reads/writes
  `~/lazymux/.lazymux.json` and clones repos under `~/lazymux/` by default.
- **`dev`** — `dirName` is overridden to `lazymux-dev`, so `lazymux-dev` reads/writes
  `~/lazymux-dev/.lazymux.json` and clones repos under `~/lazymux-dev/` instead.
  This keeps local development fully sandboxed from your real repo tree — you can
  run `lazymux-dev` against throwaway clones without touching `~/lazymux`.

Both binaries also embed a version string via `-X main.buildVersion=...` (derived
from `git describe`); the dev build appends a `-dev` suffix to that version.
