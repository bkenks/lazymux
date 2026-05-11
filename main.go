package main

import (
	"fmt"
	"os"

	"github.com/bkenks/lazymux/internal/app"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Println("lazymux", version())
			return
		case "-h", "--help":
			printHelp()
			return
		}
	}

	cfg := config.Load()
	styles.Apply(cfg.UI.Theme)

	tui := app.New(cfg)
	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazymux: fatal:", err)
		os.Exit(1)
	}
}

// version is replaced at build time via -ldflags "-X main.buildVersion=..."
var buildVersion = "dev"

func version() string { return buildVersion }

func printHelp() {
	fmt.Println(`lazymux — a TUI for ghq + lazygit + your editor

Usage: lazymux [flags]

Flags:
  -h, --help     show this help
  -v, --version  show version

Configuration:
  Edit ~/.config/lazymux/config.toml (or $XDG_CONFIG_HOME/lazymux/config.toml)
  to change the editor, default protocol, theme, and tool paths.

Keybindings (repo list):
  tab       open with lazygit
  ctrl+o    open in editor
  s         shell in repo dir
  y         copy absolute path
  r         refresh
  ctrl+n    clone new repos
  ctrl+\    delete selected repo
  ctrl+s    settings
  q         quit`)
}
