# Building lazymux

Build/install targets live in `build/Makefile`. Run them with
`make -C build <target>` (or `cd build && make <target>`).

## Commands

| Command                     | Output                        |
|------------------------------|--------------------------------|
| `make -C build build`        | `build/bin/lazymux`            |
| `make -C build dev`          | `build/bin/lazymux-dev`        |
| `make -C build install`      | installs `lazymux` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `make -C build install-dev`  | installs `lazymux-dev` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `make -C build clean`        | removes `build/bin`            |

## Regular build vs. dev build

Both targets compile the same source. The only difference is a build-time flag
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
