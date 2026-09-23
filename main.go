package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/bkenks/lazymux/internal/app"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/mcp"
	"github.com/bkenks/lazymux/internal/styles"
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
		case "mcp":
			if err := mcp.Run(os.Args[2:], version()); err != nil {
				fmt.Fprintln(os.Stderr, "lazymux mcp:", err)
				os.Exit(1)
			}
			return
		}
	}

	cfg := config.Load()
	styles.Apply(cfg.UI.Theme)

	tui := app.New(cfg, version())
	p := tea.NewProgram(tui)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazymux: fatal:", err)
		os.Exit(1)
	}
}

// version is replaced at build time via -ldflags "-X main.buildVersion=..."
var buildVersion = "dev"

func version() string { return buildVersion }

func printHelp() {
	fmt.Println(`lazymux — a TUI git repo manager (clone + your editor + custom keybinds)

Usage: lazymux [flags]
       lazymux mcp <command>

Flags:
  -h, --help     show this help
  -v, --version  show version

Commands:
  mcp            serve the repo inventory to LLMs over MCP
                 (start, stop, serve, list, set-url, set-port)

Configuration:
  All settings, the forge registry, and per-repo forge links live in a single
  ~/lazymux/.lazymux.json (override the path with $LAZYMUX_CONFIG). Repos are
  cloned into ~/lazymux/<namespace>/<repo>.

Keybindings (repo list):
  ctrl+shift+k  manage custom keybinds
  ctrl+o    open in editor
  s         shell in repo dir
  y         copy absolute path
  r         refresh
  ctrl+n    clone new repos
  f         edit selected repo's forge links
  F         manage the forge registry
  ctrl+\    delete selected repo
  ctrl+s    settings
  q         quit`)
}
