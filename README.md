# lazymux

> A terminal UI that brings your entire GitHub workflow into one place — browse, manage, clone, and hack on repos without ever leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/bkenks/lazymux)
![Version](https://img.shields.io/github/v/tag/bkenks/lazymux?label=version)

---

![Lazymux Demo](assets/Lazymux%20Demo.gif)

---

## What is lazymux?

**lazymux** is a TUI (Terminal User Interface) application built with [Bubbletea](https://github.com/charmbracelet/bubbletea) that combines `ghq` and `lazygit` into a single, unified workflow. It gives you a searchable list of all your repositories, and from there you can jump into git history, clone new repos, delete old ones, or open projects directly in VSCodium — all with a keystroke.

No more `cd`-ing around. No more remembering paths. Just launch `lazymux` and go.

---

## Features

- **Browse all repos** managed by `ghq` in a clean, filterable list
- **Open with lazygit** to manage commits, branches, PRs, and more
- **Clone repos** via `ghq` with a built-in input view
- **Delete repos** with a confirmation prompt so you don't accidentally nuke anything
- **Open in VSCodium** directly from the repo list
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

Just run:

```bash
lazymux
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
| `ctrl+o` | Open selected repo in **VSCodium** |
| `ctrl+n` | Clone a new repository |
| `ctrl+\` | Delete the selected repository |
| `q` | Quit |

### Clone / Confirm Dialogs

| Key | Action |
|---|---|
| `ctrl+p` | Proceed (confirm clone or delete) |
| `esc` | Cancel / go back |

---

## How It Works

lazymux is built using the [Charmbracelet](https://github.com/charmbracelet) stack:

- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — Elm-inspired TUI framework for Go
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — Pre-built TUI components (list, text input, key bindings)
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling and layout

On startup, lazymux calls `ghq list` to populate the repository list. From there, selecting a repo launches `lazygit` in that directory, cloning calls `ghq get`, and deletion removes the local directory managed by `ghq`.

---

## License

[MIT](LICENSE)
