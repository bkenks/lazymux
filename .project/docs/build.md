# Building lazymux

Build, install and release tasks are uv/Python scripts in `.mise/tasks/`. Run them
with `mise run <task>` from anywhere in the repo; `mise tasks` lists them.

`mise.toml` at the repo root pins the toolchain (Go and uv), so `mise install`
provisions everything. Each script carries its own PEP 723 header and runs under
`uv run --script` — there is no virtualenv to create and no dependencies to install.

## Commands

| Command                       | Output                        |
|-------------------------------|--------------------------------|
| `mise run build`              | `build/bin/lazymux` (this machine) |
| `mise run build --all`        | `build/dist/*` — all 6 platforms + `SHA256SUMS` |
| `mise run build --platform GOOS/GOARCH` | `build/dist/*` — one platform (repeatable) |
| `mise run build --version X.Y.Z` | stamps an explicit version instead of `git describe` |
| `mise run build --list`       | prints the release matrix       |
| `mise run dev`                | `build/bin/lazymux-dev`        |
| `mise run install`            | installs `lazymux` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `mise run install-dev`        | installs `lazymux-dev` to `$GOBIN` (or `$(go env GOPATH)/bin`) |
| `mise run clean`              | removes `build/bin` and `build/dist` |
| `mise run release <bump>`     | tests, tags and pushes; CI builds and publishes (see below) |

`install` and `install-dev` declare `#MISE depends=` on `build` / `dev`, so they
compile first.

## Cross-compilation

Plain `mise run build` is host-only, so the local edit-rebuild loop and
`mise run install` stay fast (~1s). `--all` cross-compiles the release matrix
into `build/dist/` (~20s):

| GOOS    | GOARCH        |
|---------|---------------|
| darwin  | amd64, arm64  |
| linux   | amd64, arm64  |
| windows | amd64, arm64  |

lazymux is pure Go — there is no cgo anywhere in its dependency graph — so these
need nothing but `GOOS`/`GOARCH`; no C cross-toolchain is involved. Matrix builds
set `CGO_ENABLED=0`, so the Linux binaries are static. Windows artifacts get a
`.exe` suffix. Artifacts are named `lazymux-<version>-<goos>-<goarch>`, and a
`SHA256SUMS` file is written alongside them in the standard `sha256sum -c` format.

The matrix lives in `PLATFORMS` in `.mise/tasks/_lib.py`; `--platform` only accepts
pairs from that list.

## Task layout

```
.mise/tasks/
  _lib.py       shared helpers — not executable, so mise ignores it
  build.py      \
  dev.py         |
  install.py     |  executable PEP 723 scripts, one per task
  install-dev.py |
  clean.py       |
  release.py    /
```

Tasks are file tasks, so mise passes arguments straight through to the script;
`argparse` handles them. `mise run release --help` prints the script's own help.
Lint with `uvx ruff check .mise/tasks/` and `uvx ty check .mise/tasks/`.

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

## Cutting a release

```bash
mise run release patch        # 1.0.2 -> 1.0.3
mise run release minor        # 1.0.2 -> 1.1.0
mise run release major        # 1.0.2 -> 2.0.0
mise run release 1.4.0        # explicit version
mise run release patch --dry-run   # print the plan, change nothing
```

Tags are bare `X.Y.Z` — no `v` prefix. The new version is derived from the
highest existing tag, including any left over from the old `vX.Y.Z` scheme, and
is always written unprefixed. Flags:
`--dry-run` and `--yes` (skip the confirmation prompt; required when stdin is not
a TTY).

**Preflight** — the release aborts before touching anything if the working tree is
dirty, the current branch is not `main`, `main` has diverged from `origin/main`,
the target tag already exists, or the new version is not greater than the latest tag.

**Order of operations** — `go vet` and `go test` run, then the tag is created and
pushed. A tree that fails its tests never gets tagged. If `git push` of the tag
fails, the local tag is deleted so a retry starts clean.

**Publishing** — pushing the tag is the whole job. Woodpecker builds the binaries
and creates the Forgejo release; see [CI](#ci) below. Nothing is built locally, so
the task finishes in seconds.

## CI

`.woodpecker.yml` runs on a pushed `X.Y.Z` tag and nothing else — no branch or
PR builds. Three steps, in order:

1. **test** — `mise install`, then `go vet ./...` and `go test ./...`.
2. **build** — `mise run build --all --version $CI_COMMIT_TAG`, producing the six
   binaries plus `SHA256SUMS` in `build/dist/`.
3. **release** — [`woodpeckerci/plugin-release`](https://woodpecker-ci.org/plugins/release)
   creates the Forgejo release for the tag and attaches everything in `build/dist/`.

The version comes from `$CI_COMMIT_TAG` rather than `git describe`: Woodpecker
clones shallowly, so the tag history `describe` needs may not be there. That is
what `mise run build --version` exists for.

The first two steps run in the `jdxcode/mise` image and provision the toolchain
from `mise.toml`, so CI compiles with the same Go version as a local build.

**Setup** — the pipeline needs one Woodpecker secret, `forgejo_token`: a Forgejo
access token with the `write:repository` and `read:misc` scopes. The forge URL is
taken from Woodpecker's own `$CI_FORGE_URL`, so nothing else is hardcoded.

lazymux is not distributed through Homebrew. Releases ship prebuilt binaries; install
one from the releases page or use `go install`.
