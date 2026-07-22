# lazymux

> A terminal UI that brings your entire repo workflow into one place — clone, organize, browse, and hack on repos across multiple git forges without ever leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/bkenks/lazymux)
![Version](https://img.shields.io/github/v/tag/bkenks/lazymux?label=version)

---

## What is lazymux?

**lazymux** is a TUI (Terminal User Interface) built with [Bubbletea](https://github.com/charmbracelet/bubbletea) that manages where your repositories live and unifies them with [lazygit](https://github.com/jesseduffield/lazygit) and your editor in a single workflow. It gives you a searchable list of all your repos, and from there you can jump into git history, clone new repos, delete old ones, copy a repo's path, drop into a shell, or open the project in your editor — all with a keystroke.

It manages repo locations natively (no `ghq` required): repos are cloned into `~/lazymux/<namespace>/<repo>`, and a **forge registry** lets you link each repo to one or more git hosts (GitHub, a self-hosted Forgejo/Gitea, GitLab, …) and pick which one is *primary*.

No more `cd`-ing around. No more remembering paths. Just launch `lazymux` and go.

---

## Forges & the placeholder remote

lazymux is built for repos that live on more than one host — for example a self-hosted Forgejo that mirrors to GitHub.

- You register your **forges** once (a name + host, e.g. `github` → `github.com`).
- When you clone, lazymux auto-matches the URL's host to a registered forge, and lets you link additional forges the repo is mirrored to. One linked forge is the **primary**.
- Under the hood, every managed repo's `origin` is rewritten to a stable placeholder host (`lazymux-placeholder`), and a per-repo local git [`insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf) rule resolves the placeholder to the **primary** forge:

  ```ini
  [remote "origin"]
      url = https://lazymux-placeholder/bkenks/myrepo.git   # never changes
  [url "https://github.com/"]
      insteadOf = https://lazymux-placeholder/               # primary = github
  ```

- If a forge goes down or you just want to point somewhere else, **switch the primary** (`f` on the repo) and lazymux re-renders that one rule. The stored `origin` never changes — only the host it resolves to. There's no automatic failover; you're always in control of which forge is live.

---

## Features

- **Native repo management** — clone into `~/lazymux/<namespace>/<repo>`, list, delete, and pull-all, all with plain `git` (no `ghq`)
- **Forge registry** — register git hosts and link repos to one or more of them, with a per-repo primary
- **Stable placeholder remotes** — switch a repo's forge without ever touching its `origin`
- **Browse all repos** in a clean, filterable list, sorted by most-recently used
- **Open with lazygit** to manage commits, branches, PRs, and more
- **Open in your editor** — codium, code, nvim, vim, helix, zed, idea, or whatever you configure
- **Drop into a shell** in the selected repo's directory
- **Copy the repo's absolute path** to your clipboard
- **Delete repos** with a confirmation prompt
- **Single JSON config** at `~/lazymux/.lazymux.json` — settings, forge registry, and per-repo links in one place
- **Status footer** surfaces errors and confirmations without crashing the TUI
- Reactive UI that adapts to your terminal size

---

## Requirements

| Tool | Purpose |
|---|---|
| [git](https://git-scm.com/) | Clone, pull, and the `insteadOf` remote rewriting lazymux relies on |
| [lazygit](https://github.com/jesseduffield/lazygit) | TUI for git — commits, PRs, branches, diffs, and more |

---

## Installation

```bash
go install github.com/bkenks/lazymux@latest
```

This builds and installs the `lazymux` binary into `$(go env GOPATH)/bin` — make sure that's on your `$PATH`.

---

## Usage

```bash
lazymux            # launch the TUI
lazymux --help     # show keybindings + config location
lazymux --version  # show the version
lazymux mcp start  # serve the repo inventory to LLMs (see "MCP server")
```

On first run, lazymux creates `~/lazymux/` and a `.lazymux.json` config (migrating an existing `~/.config/lazymux/config.toml` if present). It then lists any repos already under `~/lazymux/`. Register your forges (`F`), then clone (`Ctrl+N`) to start pulling repos in.

---

## Keybindings

### Repository List

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate the repository list |
| `/` | Filter / search repositories |
| `Tab` | Open selected repo in **lazygit** |
| `Ctrl+O` | Open selected repo in your **editor** |
| `s` | Open a **shell** in the repo's directory |
| `y` | **Copy** the absolute repo path to clipboard |
| `r` | **Refresh** the repo list |
| `Ctrl+N` | **Clone** new repositories |
| `Ctrl+P` | **Pull** every repo (`git pull --ff-only`, skips conflicts) |
| `f` | Edit the selected repo's **forge links** — forges, primary, scheme |
| `F` | Manage the **forge registry** |
| `Ctrl+\` | **Delete** the selected repository |
| `Ctrl+S` | Open **settings** |
| `q` / `Ctrl+C` | Quit |

### Clone → Forge Select

After entering one or more clone URLs, lazymux steps through each repo so you can confirm its forge links.

| Key | Action |
|---|---|
| `↑` / `↓` | Move the cursor |
| `Space` | Toggle the forge under the cursor on/off |
| `p` | Set the forge under the cursor as **primary** |
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

Each row shows how many repos link it. Deleting or renaming a forge cascades into the repos that use it: a rename updates their links, and a delete drops it — promoting another linked forge to primary, or leaving the repo unlinked if it was its only one. Repos whose primary changed have their remote re-rendered automatically.

### Repo Forges (`f`)

| Key | Action |
|---|---|
| `↑` / `↓` | Move the cursor |
| `Space` | Toggle the forge under the cursor |
| `p` | Set the forge under the cursor as **primary** |
| `s` | Toggle the URL **scheme** (https ↔ ssh) |
| `Esc` | Save & back (re-renders the repo's remote) |

### Confirm / Settings

| Key | Action |
|---|---|
| `Ctrl+P` | Proceed (confirm delete) |
| `←` / `h` · `→` / `l` / `Enter` / `Space` | Previous / next setting value |
| `Esc` | Cancel / back |

Changes save to disk immediately.

---

## MCP server

lazymux already knows where every repo on your machine lives. `lazymux mcp` hands that
inventory to an LLM over the [Model Context Protocol](https://modelcontextprotocol.io),
so an assistant can work out *which* repo a request is about — "fix the login bug on the
marketing site" — and get back an absolute path instead of guessing or globbing your home
directory.

```bash
lazymux mcp start            # start it in the background
lazymux mcp stop             # stop it
lazymux mcp list             # config, endpoint, and whether it's running
lazymux mcp set-port 8080    # change the port
lazymux mcp set-url 0.0.0.0  # change the bind host (accepts host, host:port, or a full URL)
lazymux mcp serve            # run in the foreground, for a supervisor or for debugging
```

Then point a client at the endpoint (`http://127.0.0.1:7777/mcp` by default):

```bash
claude mcp add --transport http lazymux http://127.0.0.1:7777/mcp
```

### Tools

| tool | what it does |
|---|---|
| `list_repositories` | every managed repo — path, forge links, recorded purpose. Recently-used first. |
| `search_repositories` | rank repos against a plain-English description of the task |
| `get_repository` | one repo by its `<namespace>/<name>` key |
| `set_repository_purpose` | write a `purpose` and/or `context` back into `.lazymux.json` |

That last one is what makes this improve over time. A repo starts out as just a path; once
an assistant works out what it's for it records a purpose, and every later session routes
straight there. `list_repositories` reports which repos still have nothing recorded, so a
model knows what's worth describing.

Purposes land in the same `repos` object as forge links:

```json
"repos": {
  "bkenks/myrepo": {
    "forges": ["forgejo", "github"],
    "primary": "forgejo",
    "scheme": "https",
    "purpose": "compose stacks for the homelab",
    "context": "One directory per stack. Deployed by Komodo; don't edit .env by hand."
  }
}
```

You can write these by hand too — the MCP server is just one way to fill them in.

The server binds to `127.0.0.1` by default, so the inventory isn't exposed to your network.
`set-url 0.0.0.0` opts into that; there's no authentication, so only do it on a network you
trust. Changing the host or port doesn't affect a server that's already running — stop and
start it to apply.

State lives next to the config: `.lazymux-mcp.pid` and `.lazymux-mcp.log` in the same
directory as `.lazymux.json`, so `$LAZYMUX_CONFIG` keeps a dev instance fully separate.

---

## Configuration

Everything lives in a single JSON file at `~/lazymux/.lazymux.json` (override the path with `$LAZYMUX_CONFIG`). It's created on first run — with a one-time migration from the legacy `~/.config/lazymux/config.toml` if that exists. Edit it directly or use the in-app screens.

```json
{
  "baseDir": "/home/you/lazymux",
  "placeholderHost": "lazymux-placeholder",
  "tools": {
    "lazygit": "lazygit",
    "editor": "codium",
    "shell": ""
  },
  "ui": {
    "theme": "default",
    "showFullPath": false
  },
  "behavior": {
    "defaultProtocol": "https",
    "confirmDelete": true
  },
  "mcp": {
    "host": "127.0.0.1",
    "port": 7777,
    "path": "/mcp"
  },
  "forges": [
    { "name": "github", "host": "github.com" },
    { "name": "forgejo", "host": "fj.example.com" }
  ],
  "repos": {
    "bkenks/myrepo": {
      "forges": ["forgejo", "github"],
      "primary": "forgejo",
      "scheme": "https",
      "purpose": "compose stacks for the homelab"
    }
  }
}
```

- `baseDir` — root under which repos live as `<namespace>/<repo>`.
- `placeholderHost` — the fake host stored in every managed repo's `origin`.
- `mcp` — where the MCP server binds (managed with `lazymux mcp set-url` / `set-port`).
- `forges` — the registry (managed in-app with `F`).
- `repos` — per-repo forge links, primary, and scheme (managed in-app with `f`), plus the
  `purpose`/`context` the MCP server reads and writes.

The in-app settings screen covers `editor`, `defaultProtocol`, `confirm_delete`, and `showFullPath`. Tool paths (`lazygit`, `shell`) and `theme` are file-only for now — edit and relaunch.

Repo interaction history (used for recency sorting) lives at `$XDG_DATA_HOME/lazymux/interactions.json` (fallback `~/.local/share/lazymux/interactions.json`).

---

## How It Works

lazymux is built using the [Charmbracelet](https://github.com/charmbracelet) stack:

- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — Elm-inspired TUI framework for Go
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — Pre-built TUI components (list, text input, key bindings)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling and layout

On startup, lazymux walks `~/lazymux/` to populate the repository list. Cloning runs `git clone` against the real URL, then rewrites the repo to a placeholder `origin` resolved to its primary forge; selecting a repo launches `lazygit`; deletion removes the local directory (and now-empty namespace parents). Errors surface in the status footer instead of crashing the TUI.

---

## License

[MIT](LICENSE)
