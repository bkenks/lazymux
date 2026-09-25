package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bkenks/lazymux/internal/app"
	"github.com/bkenks/lazymux/internal/config"
	"github.com/bkenks/lazymux/internal/constants"
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
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	if err := app.ApplyColors(cfg.UI.Colors, isDark); err != nil {
		cfg.Warnings = append(cfg.Warnings, err.Error())
	}

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
	fmt.Println(`lazymux — a TUI git repo manager (clone + custom keybinds)

Usage: lazymux [flags]

Flags:
  -h, --help     show this help
  -v, --version  show version

Configuration:
  All settings, the forge registry, and per-repo forge links live in a single
  $XDG_CONFIG_HOME/lazymux/config.json, by default ~/.config/lazymux/config.json
  (override the path with $LAZYMUX_CONFIG). Repos are cloned into
  <repos>/<namespace>/<repo>, where <repos> is $LAZYMUX_REPOS or else the
  config's reposDir. If neither names an existing directory, lazymux asks for
  one at startup.

Keybindings (repo list):`)
	fmt.Print(repoListKeysHelp())
}

// repoListKeysHelp lists the repo-list keys for --help, from the same table
// the in-app help uses.
func repoListKeysHelp() string {
	var b strings.Builder
	row := func(keys, desc string) { fmt.Fprintf(&b, "  %-13s %s\n", keys, desc) }
	row("/", "filter repos")
	for _, c := range constants.RepoListKeyMap.Commands() {
		row(c.Binding.Help().Key, c.Full)
	}
	row("?", "full help")
	row("esc", "back (never quits)")
	return b.String()
}
