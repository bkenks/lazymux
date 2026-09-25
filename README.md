# lazymux

> A terminal UI that brings your entire repo workflow into one place — clone, organize, browse, and hack on repos across multiple git forges without ever leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.25.8+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/bkenks/lazymux)
![Version](https://img.shields.io/github/v/tag/bkenks/lazymux?label=version)

---

## What is lazymux?

**lazymux** is a TUI (Terminal User Interface) built with [Bubbletea](https://github.com/charmbracelet/bubbletea) that manages where your repositories live and unifies them with your editor and the terminal tools you use in a single workflow. It gives you a searchable list of all your repos, and from there you can clone new repos, delete old ones, copy a repo's path, drop into a shell, open the project in your editor, or run any command you've bound to a key — all with a keystroke.

It manages repo locations natively (no `ghq` required): repos are cloned into `<repos>/<namespace>/<repo>` (see [Repo directory](#repo-directory)), and a **forge registry** lets you link each repo to one or more git hosts (GitHub, a self-hosted Forgejo/Gitea, GitLab, …) as **upstreams**, with one of them set as the **origin** you fetch from.

No more `cd`-ing around. No more remembering paths. Just launch `lazymux` and go.

---

## Forges & the placeholder remote

lazymux is built for repos that live on more than one host — for example a self-hosted Forgejo that mirrors to GitHub.

- You register your **forges** once (a name + host, e.g. `github` → `github.com`).
- When you clone, lazymux auto-matches the URL's host to a registered forge, and lets you check off every forge the repo is pushed to — its **upstreams**. Exactly one upstream is the **origin**: the host fetch and pull read from.
- Under the hood, every managed repo's `origin` is rewritten to a stable placeholder host (`lazymux-placeholder`), and a per-repo local git [`insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf) rule resolves the placeholder to the **origin** forge. With more than one upstream, each gets a `pushurl`, so a single `git push` fans out to all of them:

  ```ini
  [remote "origin"]
      url = https://lazymux-placeholder/bkenks/myrepo.git    # never changes
      pushurl = https://github.com/bkenks/myrepo.git         # upstream
      pushurl = https://fj.example.com/bkenks/myrepo.git     # upstream
  [url "https://github.com/"]
      insteadOf = https://lazymux-placeholder/               # origin = github
  ```

- If a forge goes down or you just want to fetch from somewhere else, **switch the origin** (`3` on the repo) and lazymux re-renders that one rule. The stored `origin` URL never changes — only the host it resolves to. There's no automatic failover; you're always in control of which forge is live.
- A repo with a single upstream gets no `pushurl` at all: push follows the placeholder origin, exactly as before.

---

## Features

- **Native repo management** — clone into `<repos>/<namespace>/<repo>`, list, delete, and pull-all, all with plain `git` (no `ghq`)
- **Forge registry** — register git hosts and link repos to one or more of them as upstreams, with a per-repo origin
- **Stable placeholder remotes** — switch a repo's forge without ever touching its `origin`
- **Browse all repos** in a clean, filterable list, sorted by most-recently used
- **Release tags** — tag & push a repo's next major, minor or patch version, in a per-repo format like `v1.2.3`, `mypkg/v1.2.3` or `1.2.3-mypkg`
- **Custom keybinds** — bind a key to any shell command (e.g. `lazygit`); it runs in the selected repo with the whole terminal
- **Open in your editor** — any command on your `PATH`; the settings screen checks it resolves before saving
- **Drop into a shell** in the selected repo's directory
- **Copy the repo's absolute path** to your clipboard
- **Delete repos** with a confirmation prompt
- **Single JSON config** at `~/.config/lazymux/config.json` — settings, forge registry, and per-repo links in one place
- **Status footer** surfaces errors and confirmations without crashing the TUI
- Reactive UI that adapts to your terminal size

---

## Requirements

| Tool | Purpose |
|---|---|
| [git](https://git-scm.com/) | Clone, pull, and the `insteadOf` remote rewriting lazymux relies on |

---

## Installation

### Prebuilt binary

Every release carries binaries for macOS, Linux and Windows on both amd64 and arm64,
plus a `SHA256SUMS` file. Grab one from the
[releases page](https://fj.ktbcloud.com/bkenks/lazymux/releases), verify it, and drop it
on your `$PATH`:

```bash
sha256sum -c SHA256SUMS --ignore-missing
chmod +x lazymux-*-darwin-arm64
mv lazymux-*-darwin-arm64 ~/.local/bin/lazymux
```

### go install

```bash
go install github.com/bkenks/lazymux@latest
```

This builds and installs the `lazymux` binary into `$GOBIN` (or `$(go env GOPATH)/bin`) —
make sure that's on your `$PATH`. It resolves through the GitHub mirror, so it lags a
release until that mirror has the tag.

### From source

```bash
git clone https://fj.ktbcloud.com/bkenks/lazymux.git
cd lazymux
mise run install
```

Same destination, built from your checkout. See [building
lazymux](.project/docs/build.md) for the other build tasks.

---

## Usage

```bash
lazymux            # launch the TUI
lazymux --help     # show keybindings + config location
lazymux --version  # show the version
```

On first run, lazymux creates `~/.config/lazymux/config.json` (moving an existing `~/lazymux/.lazymux.json` there, or migrating a `~/.config/lazymux/config.toml`, if present). If no [repo directory](#repo-directory) is set yet, it asks you for one, then lists any repos already in it. Register your forges (`F`), then clone (`n`) to start pulling repos in.

---

## Keybindings

### Repository List

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate the repository list |
| `/` | Filter / search repositories |
| `o` | Open selected repo in your **editor** |
| `s` | Open a **shell** in the repo's directory |
| `y` | **Copy** the absolute repo path to clipboard |
| `r` | **Refresh** the repo list |
| `n` | **Clone** new repositories |
| `p` | **Pull** every repo (`git pull --ff-only`, skips conflicts) |
| `S` | Cycle the **sort order** (recent → name a-z → name z-a → namespace) |
| `g` | Show/hide the **forge label** on rows |
| `t` | Show/hide the **git stats** on rows |
| `v` | **Tag & push** the selected repo's next major, minor or patch version |
| `F` | Manage the **forge registry** |
| `d` | **Delete** the selected repository |
| `1` | Open **settings** |
| `2` | Manage **custom keybinds** |
| `3` | Open the selected repo's **settings** — tag format, upstreams, origin, scheme |
| `Esc` | Clear the filter — `Esc` is back on every screen and never quits |
| `q` / `Ctrl+C` | Quit |

Repo-list keys are unmodified letters, with the settings-style screens on numbers.
Screens with a text field (clone, add-forge) keep their `Ctrl` shortcuts, since plain
letters go into the input there.

The active sort shows in the list title, and the order you pick is saved — the list comes
back in it next launch. It's also in the settings screen as **Sort repos by**.

### Clone → Forge Select

After entering one or more clone URLs, lazymux steps through each repo so you can confirm its forge links.

| Key | Action |
|---|---|
| `↑` / `↓` | Move the cursor |
| `Space` | Toggle the forge under the cursor as an **upstream** |
| `o` | Set the forge under the cursor as the **origin** (fetch/pull) |
| `s` | Toggle the URL **scheme** (https ↔ ssh) for this repo |
| `a` | **Add a new forge** from this repo's clone URL |
| `Enter` | Confirm this repo (advance to the next) |
| `Esc` | Cancel the whole clone |

### Forge Registry (`F`)

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate forges |
| `a` | Add a forge |
| `e` / `Enter` | Edit the selected forge |
| `d` | Delete the selected forge |
| `Tab` | Switch between name / host fields (while editing) |
| `Esc` | Save & back |

Each row shows how many repos link it. Deleting or renaming a forge cascades into the repos that use it: a rename updates their links, and a delete drops it — promoting another upstream to origin, or leaving the repo unlinked if it was its only one. Repos whose links changed have their remotes re-rendered automatically.

### Repo Settings (`3`)

One form for the selected repo:

- **Tag prefix** / **Tag suffix** — the text its version tags wrap around `MAJOR.MINOR.PATCH`,
  so `v` gives `v1.2.3`, `mypkg/v` gives `mypkg/v1.2.3`, and a `-mypkg` suffix gives
  `1.2.3-mypkg`. Both start empty (`1.2.3`). The suffix field previews the repo's latest tag in
  that format and the next patch / minor / major tags.
- **Upstreams** — every forge a push goes to (`space` / `x` toggles one).
- **Origin** — the upstream fetch and pull read from.
- **Scheme** — https or ssh.

`enter` moves to the next field and saves on the last one, re-rendering the repo's remote;
`esc` leaves without saving.

### Tag Version (`v`)

Pick **patch**, **minor** or **major**. Each option shows the tag it creates: the repo's highest
local tag in its [tag format](#repo-settings-3), bumped, or a bump from `0.0.0` when there is none
yet. Then confirm to create an annotated tag at `HEAD` and push just that tag to `origin`, which
reaches every upstream. If the push fails, the tag stays in the local repo and the error says
so; push it yourself with `git push origin <tag>`. Tags are read locally, so run `git fetch
--tags` first if someone else may have released.

### Custom Keybinds (`2`)

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate keybinds |
| `n` | **New** keybind |
| `e` | **Edit** the selected keybind |
| `Ctrl+\` | **Delete** the selected keybind (asks Yes / No) |
| `Esc` | Back (cancels the form when one is open) |

Each keybind has a **Name**, a **Keybind**, a **Command** and a **Return to lazymux on
command end** toggle. Type the keybind as
text, e.g. `ctrl + g` or `cmd + shift + r`. Key names: `ctrl`, `alt`, `cmd`,
`shift`, `tab`, `caps`, `return`, `esc`, `space`, `backspace`, `del`, arrows,
`home`, `end`, `pgup`, `pgdown`, `insert`, `f1`–`f12`, plus any single character.
A keybind lazymux already uses, or another custom keybind already has, is refused.

Pressing a keybind on the repo list runs its command with your shell (`sh -c`
style) in the selected repo's directory and hands it the whole terminal, the way
`s` opens a shell. lazymux comes back when the command ends: straight away when
**Return to lazymux on command end** is on (quitting lazygit with `q` or Claude
Code with `esc`), otherwise after you press `Enter`, so a short command's output
such as `git status` can be read first.

`ctrl+shift` combos and `cmd` need a terminal that reports them (kitty keyboard
protocol — e.g. Ghostty, kitty, WezTerm, or iTerm2 with CSI u enabled).

### Confirm / Settings

| Key | Action |
|---|---|
| `Ctrl+P` | Proceed (confirm delete) |
| `←` / `h` · `→` / `l` / `Enter` / `Space` | Previous / next setting value |
| `Esc` | Cancel / back |

Changes save to disk immediately.

---

## Configuration

Everything lives in a single JSON file at `$XDG_CONFIG_HOME/lazymux/config.json`, by default `~/.config/lazymux/config.json` (override the path with `$LAZYMUX_CONFIG`). It's created on first run — moving a config from the old `~/lazymux/.lazymux.json` location, or migrating the legacy `~/.config/lazymux/config.toml`, if either exists. Edit it directly or use the in-app screens.

```json
{
  "reposDir": "/home/you/Development",
  "placeholderHost": "lazymux-placeholder",
  "tools": {
    "editor": "codium",
    "shell": ""
  },
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
  "behavior": {
    "defaultProtocol": "https",
    "confirmDelete": true
  },
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

- `reposDir` — directory repos live under as `<namespace>/<repo>`; see [Repo directory](#repo-directory).
- `placeholderHost` — the fake host stored in every managed repo's `origin`.
- `forges` — the registry (managed in-app with `F`).
- `keybinds` — custom keybinds (managed in-app with `2`).
- `repos` — per-repo upstreams, origin, scheme, and `tagPrefix`/`tagSuffix` (managed in-app
  with `3`).

- `ui.sortMode` — repo list order: `recent`, `name-asc`, `name-desc`, or `namespace` (cycled in-app with `S`).
- `ui.colors` — three hex base colors (`#7D56F4` or `#75F`) that every color in the UI comes
  from, set separately for `dark` and `light` terminal backgrounds. `main` colors title bars and
  buttons, `accent` the selected row and highlights, and `gray` text, borders and hints. lazymux
  stretches each one into a scale of lighter and darker shades and picks from those, so the
  pieces stay matched whatever you choose. lazymux asks the terminal for its background at
  launch and uses that mode's colors, falling back to dark if the terminal doesn't answer. An
  empty color keeps the default (`#5F5FD7`, `#EE6FF8`, `#777777`).

The in-app settings screen is one form covering `editor`, `defaultProtocol`, `confirmDelete`, `showFullPath`, `showForge`, `showStats`, `sortMode`, and the dark and light mode `colors`. `enter` moves to the next field and saves on the last one; `esc` leaves without saving. The editor field resolves the command on `PATH` and won't let the form save one it cannot find; the color fields only take a hex value. `shell` (the shell keybind commands and `s` use) is file-only for now — edit and relaunch.

### Repo directory

Repos live under `$LAZYMUX_REPOS` if it is set, otherwise under `reposDir` in the config. lazymux has no default: at startup, if neither names an existing directory, it asks for one before opening the repo list, creating the directory if needed and saving it as `reposDir`. The prompt only closes once a directory is set (`ctrl+c` quits).

A config that still has the older `baseDir` key is read as `reposDir`. A config moved from `~/lazymux/.lazymux.json` keeps pointing at `~/lazymux`, so existing repos stay where they are; move them and change `reposDir` to relocate them.

Repo interaction history (used for recency sorting) lives at `$XDG_DATA_HOME/lazymux/interactions.json` (fallback `~/.local/share/lazymux/interactions.json`).

---

## How It Works

lazymux is built using the [Charmbracelet](https://github.com/charmbracelet) stack:

- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — Elm-inspired TUI framework for Go
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — Pre-built TUI components (list, text input, key bindings)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling and layout

On startup, lazymux walks the repo directory to populate the repository list. Cloning runs `git clone` against the real URL, then rewrites the repo to a placeholder `origin` resolved to its origin forge (plus a push URL per upstream); a custom keybind hands the terminal to its command until it exits; deletion removes the local directory (and now-empty namespace parents). Errors surface in the status footer instead of crashing the TUI.

---

## License

[MIT](LICENSE)
