# lazymux

> Every repo on your machine in one TUI. Clone, browse, release, and push to several git forges
> without leaving the terminal.

![Go](https://img.shields.io/badge/Go-1.25.8+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/bkenks/lazymux)
![Version](https://img.shields.io/github/v/tag/bkenks/lazymux?label=version)

## Contents

- [What it does](#what-it-does)
- [Install](#install)
- [Getting started](#getting-started)
- [Repo list keys](#repo-list-keys)
- [Forges and the placeholder remote](#forges-and-the-placeholder-remote)
- [Screens](#screens)
- [Configuration](#configuration)
- [How it works](#how-it-works)
- [License](#license)

## What it does

lazymux keeps your repos in one directory as `<repos>/<namespace>/<repo>` and gives you a
searchable list of them. From that list you can:

- clone repos, pull all of them at once, or delete one
- open a shell in a repo or copy its path
- bind any key to any command: lazygit, your editor, Claude Code, whatever
- tag and push the next major, minor or patch release
- push to several forges at once and pick which one you fetch from

It's plain `git` underneath, with no `ghq`. The only requirement is [git](https://git-scm.com/).

## Install

### Install script

macOS or Linux:

```bash
curl -fsSL https://fj.ktbcloud.com/bkenks/lazymux/raw/branch/main/install.sh | bash
```

It downloads the latest release for your OS and CPU, checks it against `SHA256SUMS`, and puts it
in `~/.local/bin/lazymux`. Run it again to update.

### mise

lazymux isn't in the mise registry, so point mise's `forgejo` backend at the repo:

```bash
mise use -g 'forgejo:bkenks/lazymux[api_url=https://fj.ktbcloud.com/api/v1,bin=lazymux,minimum_release_age=0h]'
```

Or in a `mise.toml`:

```toml
[tools]
"forgejo:bkenks/lazymux" = { version = "latest", api_url = "https://fj.ktbcloud.com/api/v1", bin = "lazymux", minimum_release_age = "0h" }
```

- `api_url` points mise at my Forgejo instance.
- `bin` installs the release binary as `lazymux`.
- `minimum_release_age = "0h"` gets you a new release as soon as it's out, even if your mise
  settings hold new releases back for a while.

Update it with `mise upgrade`.

### Prebuilt binary

Every [release](https://fj.ktbcloud.com/bkenks/lazymux/releases) has binaries for macOS, Linux
and Windows on amd64 and arm64, plus a `SHA256SUMS` file. This is the way to go on Windows.

```bash
sha256sum -c SHA256SUMS --ignore-missing
chmod +x lazymux-*-darwin-arm64
mv lazymux-*-darwin-arm64 ~/.local/bin/lazymux
```

### go install

```bash
go install github.com/bkenks/lazymux@latest
```

This installs to `$GOBIN` (or `$(go env GOPATH)/bin`). It goes through the GitHub mirror, so it
can lag a release, and `lazymux --version` prints `dev`.

### From source

```bash
git clone https://fj.ktbcloud.com/bkenks/lazymux.git
cd lazymux
mise run install
```

The other build tasks are in [building lazymux](.project/docs/build.md).

## Getting started

```bash
lazymux            # launch
lazymux --help     # keys and config location
lazymux --version
```

The first launch writes `~/.config/lazymux/config.json` and asks where your repos live, unless
`$LAZYMUX_REPOS` or `reposDir` already points at a directory. Then register your forges with
`F` and clone with `c`.

## Repo list keys

| Key | Action |
|---|---|
| `↑` / `↓` | Move through the list |
| `/` | Filter |
| `s` | Open a **shell** in the repo |
| `y` | **Copy** the repo's path |
| `r` | **Refresh** the list |
| `c` | **Clone** repos |
| `p` | **Pull** every repo (`git pull --ff-only`, skips anything that can't fast-forward) |
| `S` | Cycle the **sort**: recent, name a-z, name z-a, namespace |
| `g` | Show or hide the **forge label** |
| `t` | Show or hide **git stats** |
| `v` | **Tag and push** the repo's next version |
| `F` | Manage the **forge registry** |
| `d` | **Delete** the repo |
| `1` | **Settings** |
| `2` | **Custom keybinds** |
| `3` | **Repo settings**: tag format, upstreams, origin, scheme |
| `Esc` | Clear the filter. `Esc` is back on every screen and never quits |
| `q` / `Ctrl+C` | Quit |

Letters do things, numbers open settings screens. The sort you pick is saved.

## Forges and the placeholder remote

lazymux is built for repos that live on more than one host, like a self-hosted Forgejo that
mirrors to GitHub.

- Register each forge once, a name and a host (`github` → `github.com`), with `F`.
- Each repo has **upstreams**, every forge a push goes to, and one **origin**, the upstream fetch
  and pull read from. Cloning matches the URL's host to a forge for you.
- Every repo's `origin` URL points at a fake host, `lazymux-placeholder`. A local git
  [`insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf)
  rule sends it to the origin forge, and each upstream gets a `pushurl`, so one `git push` hits
  all of them:

  ```ini
  [remote "origin"]
      url = https://lazymux-placeholder/bkenks/myrepo.git    # never changes
      pushurl = https://github.com/bkenks/myrepo.git         # upstream
      pushurl = https://fj.example.com/bkenks/myrepo.git     # upstream
  [url "https://github.com/"]
      insteadOf = https://lazymux-placeholder/               # origin = github
  ```

- Forge down? Switch the origin in repo settings (`3`). Only the `insteadOf` rule changes; the
  stored URL never does. There's no automatic failover.
- A repo with one upstream gets no `pushurl`, so push just follows `origin`.

## Screens

### Clone

Paste one or more URLs, or press `Ctrl+T` to clone every repo in a namespace on one forge (`Tab`
picks the forge). `Ctrl+P` goes, `Esc` backs out. Then lazymux walks you through each repo's
forges:

| Key | Action |
|---|---|
| `Space` | Toggle the forge as an **upstream** |
| `o` | Make the forge the **origin** |
| `s` | Switch the **scheme** (https / ssh) |
| `a` | **Add a forge** from the repo's URL |
| `Enter` | Next repo |
| `Esc` | Cancel the whole clone |

### Forge registry (`F`)

| Key | Action |
|---|---|
| `a` | Add a forge |
| `e` / `Enter` | Edit it |
| `d` | Delete it |
| `Tab` | Switch between name and host while editing |
| `Esc` | Save and go back |

Each row shows how many repos use the forge. Renaming a forge updates those repos. Deleting one
drops it from them and promotes another upstream to origin if needed. lazymux re-renders their
remotes either way.

### Repo settings (`3`)

One form for the selected repo:

- **Tag prefix** and **Tag suffix**: what goes around `MAJOR.MINOR.PATCH` in its release tags.
  `v` gives `v1.2.3`, `mypkg/v` gives `mypkg/v1.2.3`, a `-mypkg` suffix gives `1.2.3-mypkg`. The
  suffix field previews the latest tag and the next ones.
- **Upstreams**: `Space` or `x` toggles a forge.
- **Origin** and **Scheme**.

`Enter` moves through the fields and saves on the last one; `Ctrl+s` saves from any field. `Esc`
leaves without saving.

### Tag version (`v`)

Pick patch, minor or major. Each option shows the tag it'll make: the repo's highest local tag in
its tag format, bumped, or bumped from `0.0.0` if there isn't one. Confirm, and lazymux makes an
annotated tag at `HEAD` and pushes just that tag to `origin`, which reaches every upstream.

- If the push fails, the tag stays local. Retry with `git push origin <tag>`.
- Tags are read locally, so `git fetch --tags` first if someone else might have released.

### Custom keybinds (`2`)

| Key | Action |
|---|---|
| `n` | New keybind |
| `e` | Edit it |
| `Ctrl+\` | Delete it |
| `Ctrl+s` | Save the form from any field |
| `Esc` | Back, or cancel the form |

A keybind is a name, a key, a command, and a **Return to lazymux on command end** toggle. Type
the key as text, like `ctrl + g` or `cmd + shift + r`. Key names: `ctrl`, `alt`, `cmd`, `shift`,
`tab`, `caps`, `return`, `esc`, `space`, `backspace`, `del`, arrows, `home`, `end`, `pgup`,
`pgdown`, `insert`, `f1`–`f12`, or any single character. Keys lazymux or another keybind already
uses are refused.

Pressing it on the repo list runs the command in your shell, in the repo's directory, with the
whole terminal. With the toggle on, lazymux comes back as soon as the command exits (quit lazygit
with `q`, Claude Code with `esc`). With it off, you press `Enter` first, so you can read output
like `git status`.

Want an editor key? Bind `o` to `code .` (or `zed .`, `nvim .`) with the toggle on.

`ctrl+shift` combos and `cmd` need a terminal that reports them: Ghostty, kitty, WezTerm, or
iTerm2 with CSI u on.

### Delete (`d`)

`←` / `→` (or `h` / `l`) picks yes or no, `Enter` confirms, `Ctrl+P` deletes right away, `Esc`
backs out. Turn the prompt off with **Confirm before deleting** in settings.

## Configuration

Everything is in one JSON file, `$XDG_CONFIG_HOME/lazymux/config.json`
(`~/.config/lazymux/config.json` by default, or `$LAZYMUX_CONFIG`). Edit it by hand or through
the app. Older `~/lazymux/.lazymux.json` and `config.toml` configs are moved over on first launch.

```json
{
  "reposDir": "/home/you/Development",
  "placeholderHost": "lazymux-placeholder",
  "tools": { "shell": "" },
  "ui": {
    "colors": {
      "dark": { "main": "", "accent": "", "gray": "" },
      "light": { "main": "", "accent": "", "gray": "" }
    },
    "showFullPath": false,
    "showForge": true,
    "showStats": true,
    "sortMode": "recent"
  },
  "behavior": { "defaultProtocol": "https", "confirmDelete": true },
  "forges": [
    { "name": "github", "host": "github.com" },
    { "name": "forgejo", "host": "fj.example.com" }
  ],
  "keybinds": [
    { "name": "lazygit", "keys": "ctrl+g", "command": "lazygit", "returnOnExit": true }
  ],
  "repos": {
    "bkenks/myrepo": {
      "upstreams": ["forgejo", "github"],
      "origin": "forgejo",
      "scheme": "https",
      "tagPrefix": "v"
    }
  }
}
```

- `reposDir`: where repos live. `$LAZYMUX_REPOS` overrides it.
- `placeholderHost`: the fake host in every repo's `origin`.
- `tools.shell`: the shell `s` and keybinds use. File only; restart after changing it.
- `ui.sortMode`: `recent`, `name-asc`, `name-desc`, or `namespace`.
- `ui.colors`: three hex colors per terminal background. `main` is title bars and buttons,
  `accent` the selected row and highlights, `gray` text and borders. lazymux builds every other
  shade from these, so whatever you pick stays matched. Empty keeps the default (`#5F5FD7`,
  `#EE6FF8`, `#777777`).
- `forges`, `keybinds`, `repos`: managed with `F`, `2` and `3`.

Settings (`1`) covers the protocol, delete prompt, row display, sort and colors. `Enter` moves
through the fields and saves on the last one; `Ctrl+s` saves from any field; `Esc` leaves without
saving.

Recent-use history for the sort is in `$XDG_DATA_HOME/lazymux/interactions.json`
(`~/.local/share/lazymux/interactions.json` by default).

## How it works

lazymux is Go on the [Charm](https://github.com/charmbracelet) stack: Bubbletea, Bubbles, Huh and
Lipgloss. On launch it walks the repo directory to build the list. Cloning runs `git clone` on
the real URL, then rewrites the repo to the placeholder remote. Keybinds hand the terminal to
their command until it exits. Deleting removes the directory and any namespace directories left
empty. Errors show up in the footer instead of crashing the app.

## License

[GPL-3.0](LICENSE)
