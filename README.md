# lazymux

> A terminal UI that brings your entire GitHub workflow into one place — browse, manage, clone, and hack on repos without ever leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/bkenks/lazymux)
![Version](https://img.shields.io/github/v/tag/bkenks/lazymux?label=version)

---

## What is lazymux?

**lazymux** is a TUI (Terminal User Interface) application built with [Bubbletea](https://github.com/charmbracelet/bubbletea) that combines `ghq` and `lazygit` into a single, unified workflow. It gives you a searchable list of all your repositories, and from there you can jump into git history, clone new repos, delete old ones, copy a repo's path, drop into a shell in the repo directory, or open the project in your editor — all with a keystroke.

No more `cd`-ing around. No more remembering paths. Just launch `lazymux` and go.

---

## Features

- **Browse all repos** managed by `ghq` in a clean, filterable list, sorted by most-recently used
- **Open with lazygit** to manage commits, branches, PRs, and more
- **Open in your editor** — codium, code, nvim, vim, helix, zed, idea, or whatever you configure
- **Drop into a shell** in the selected repo's directory
- **Copy the repo's absolute path** to your clipboard
- **Clone repos** via `ghq` with live progress and a failure count
- **Delete repos** with a confirmation prompt
- **Persistent settings** at `$XDG_CONFIG_HOME/lazymux/config.toml`
- **Status footer** surfaces errors and confirmations without crashing the TUI
- Reactive UI that adapts to your terminal size

---

## Requirements

lazymux wraps two tools — make sure both are installed before proceeding.

| Tool | Purpose |
|---|---|
| [lazygit](https://github.com/jesseduffield/lazygit) | TUI for git — commits, PRs, branches, diffs, and more |
| [ghq](https://github.com/x-motemen/ghq) | Repository manager — clones and organizes repos in a predictable folder structure |

> lazymux relies on `ghq`'s structured directory layout to find and display your repositories. Without it, the repo list will be empty.

---

## Installation

### Option 1 — Installer Script

```bash
curl -fsSL https://raw.githubusercontent.com/bkenks/lazymux/main/installer.sh | sh
```

> **macOS note:** You may see a warning that the binary is not verified or signed. This is because the binary is not signed with an Apple Developer certificate. To bypass it: dismiss the prompt, go to **Settings → Privacy & Security**, scroll to the bottom and click **Allow Lazymux**, then run the program again. You will only be prompted once. To avoid this entirely, use the Go installation below.

### Option 2 — Go Install

```bash
go install github.com/bkenks/lazymux@latest
```

---

## Usage

```bash
lazymux            # launch the TUI
lazymux --help     # show keybindings + config location
lazymux --version  # show the version
```

lazymux opens with your full list of `ghq`-managed repositories. Use the keybindings below to navigate and take action.

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
| `Ctrl+\` | **Delete** the selected repository |
| `Ctrl+S` | Open **settings** |
| `q` / `Ctrl+C` | Quit |

### Clone / Confirm Dialogs

| Key | Action |
|---|---|
| `Ctrl+P` | Proceed (confirm clone or delete) |
| `Esc` | Cancel / go back |

### Settings

| Key | Action |
|---|---|
| `←` / `h` | Previous value |
| `→` / `l` / `Enter` / `Space` | Next value |
| `Esc` | Back to main |

Changes save to disk immediately.

---

## Configuration

Settings live at `$XDG_CONFIG_HOME/lazymux/config.toml` (defaulting to `~/.config/lazymux/config.toml`). The file is created on first run with sensible defaults; edit it directly or use the in-app settings screen for the most common options.

```toml
[tools]
ghq     = "ghq"      # path or name of the ghq binary
lazygit = "lazygit"  # path or name of the lazygit binary
editor  = "codium"   # codium, code, nvim, vim, hx, zed, idea, or any command on $PATH
shell   = ""         # leave empty to use $SHELL (then /bin/sh fallback)

[ui]
theme          = "default"  # "default" or "mono"
show_full_path = false

[behavior]
default_protocol = "https"  # "https" or "ssh"
confirm_delete   = true
```

In-app settings cover `editor`, `default_protocol`, `confirm_delete`, and `show_full_path`. Tool paths (`ghq`, `lazygit`, `shell`) and `theme` are TOML-only for now — edit the file directly and relaunch.

Repo interaction history (used for recency sorting) lives at `$XDG_DATA_HOME/lazymux/interactions.json` (fallback `~/.local/share/lazymux/interactions.json`).

---

## How It Works

lazymux is built using the [Charmbracelet](https://github.com/charmbracelet) stack:

- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — Elm-inspired TUI framework for Go
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — Pre-built TUI components (list, text input, key bindings)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling and layout

On startup, lazymux calls `ghq list` to populate the repository list. Selecting a repo launches `lazygit` in that directory; cloning calls `ghq get`; deletion removes the local directory managed by `ghq`. Errors from any of these surface in the status footer at the bottom of the screen instead of crashing the TUI.

---

## License

[MIT](LICENSE)
