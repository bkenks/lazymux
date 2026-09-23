package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bkenks/lazymux/internal/config"
)

// subcommand is one `lazymux mcp` command. The table below drives dispatch,
// the usage text, and the unknown-command error alike.
type subcommand struct {
	name    string
	alias   string
	args    string
	summary string
	run     func(args []string, version string) error
}

var subcommands = []subcommand{
	{
		name:    "start",
		summary: "start the server in the background",
		run:     func([]string, string) error { return Start(config.Load()) },
	},
	{
		name:    "stop",
		summary: "stop the background server",
		run:     func([]string, string) error { return Stop() },
	},
	{
		name:    "serve",
		summary: "run the server in the foreground (for supervisors / debugging)",
		run:     func(_ []string, version string) error { return runServe(version) },
	},
	{
		name:    "list",
		alias:   "status",
		summary: "show configuration, endpoint, and running status",
		run:     func([]string, string) error { return runList() },
	},
	{
		name:    "set-url",
		args:    "<v>",
		summary: "set the bind host — accepts a host, host:port, or full URL",
		run:     func(args []string, _ string) error { return runSetURL(args) },
	},
	{
		name:    "set-port",
		args:    "<n>",
		summary: "set the port",
		run:     func(args []string, _ string) error { return runSetPort(args) },
	},
}

// Run dispatches `lazymux mcp <subcommand> [args...]`. args excludes the
// "mcp" word itself. It returns an error for the caller to print and exit on.
func Run(args []string, version string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "-h", "--help", "help":
		printUsage()
		return nil
	}
	names := make([]string, 0, len(subcommands))
	for _, sub := range subcommands {
		if args[0] == sub.name || (sub.alias != "" && args[0] == sub.alias) {
			return sub.run(args[1:], version)
		}
		names = append(names, sub.name)
	}
	return fmt.Errorf("unknown mcp subcommand %q (try: %s)", args[0], strings.Join(names, ", "))
}

// runServe runs the server in the foreground. `mcp start` re-execs the binary
// this way; run it directly to keep the server attached to a supervisor. The
// pidfile is claimed only once the port is bound, so its presence means the
// server is genuinely up. A server that never bound leaves any existing
// pidfile alone, since it belongs to whichever server holds the port.
func runServe(version string) error {
	cfg := config.Load()
	release := func() {}
	defer func() { release() }()
	return Serve(context.Background(), cfg, version, func() error {
		claimed, err := claimPID()
		if err != nil {
			return err
		}
		release = claimed
		return nil
	})
}

func runList() error {
	cfg := config.Load()
	infos, err := inventory(cfg)
	if err != nil {
		return err
	}
	described := 0
	for _, r := range infos {
		if r.Described() {
			described++
		}
	}

	status := "stopped"
	if pid := Running(); pid != 0 {
		status = fmt.Sprintf("running (pid %d)", pid)
	}

	fmt.Println("lazymux MCP server")
	fmt.Println()
	fmt.Printf("  status    %s\n", status)
	fmt.Printf("  endpoint  %s\n", cfg.MCP.Endpoint())
	fmt.Printf("  host      %s\n", cfg.MCP.Host)
	fmt.Printf("  port      %d\n", cfg.MCP.Port)
	fmt.Printf("  path      %s\n", cfg.MCP.Path)
	fmt.Println()
	fmt.Printf("  config    %s\n", config.Path())
	fmt.Printf("  log       %s\n", LogPath())
	fmt.Printf("  repos     %d (%d with a recorded purpose)\n", len(infos), described)
	return nil
}

func runSetURL(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: lazymux mcp set-url <host|url>   (e.g. 127.0.0.1, 0.0.0.0:8080, http://0.0.0.0:8080/mcp)")
	}
	cfg := config.Load()
	updated, err := applyURL(cfg.MCP, args[0])
	if err != nil {
		return err
	}
	cfg.MCP = updated
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("mcp endpoint set to %s\n", cfg.MCP.Endpoint())
	warnIfRunning()
	return nil
}

func runSetPort(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: lazymux mcp set-port <port>")
	}
	port, err := parsePort(args[0])
	if err != nil {
		return err
	}
	cfg := config.Load()
	cfg.MCP.Port = port
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("mcp endpoint set to %s\n", cfg.MCP.Endpoint())
	warnIfRunning()
	return nil
}

// applyURL folds a user-supplied host, host:port, or full URL into the
// existing MCP settings. Components the user left out keep their old value.
func applyURL(current config.MCP, raw string) (config.MCP, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return current, errors.New("url is empty")
	}
	// url.Parse only recognizes the host when a scheme is present; a bare
	// "0.0.0.0:8080" would otherwise parse as scheme "0.0.0.0".
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return current, fmt.Errorf("parsing %q: %w", raw, err)
	}
	if u.Scheme != "http" {
		return current, fmt.Errorf("unsupported scheme %q: the server speaks plain http", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return current, fmt.Errorf("no host in %q", raw)
	}
	current.Host = host

	if p := u.Port(); p != "" {
		port, err := parsePort(p)
		if err != nil {
			return current, err
		}
		current.Port = port
	}
	if path := strings.TrimRight(u.Path, "/"); path != "" {
		current.Path = path
	}
	return current, nil
}

func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("port %q is not a number", s)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d is out of range (1-65535)", port)
	}
	return port, nil
}

func warnIfRunning() {
	if pid := Running(); pid != 0 {
		fmt.Printf("note: the server is still running on the old address (pid %d) — "+
			"run `lazymux mcp stop && lazymux mcp start` to apply\n", pid)
	}
}

func printUsage() { fmt.Print(composeUsage()) }

func composeUsage() string {
	var b strings.Builder
	b.WriteString("lazymux mcp — serve the repo inventory to LLMs over MCP\n\n")
	b.WriteString("Usage: lazymux mcp <command>\n\nCommands:\n")
	for _, sub := range subcommands {
		label := strings.TrimSpace(sub.name + " " + sub.args)
		if sub.alias != "" {
			label += ", " + sub.alias
		}
		fmt.Fprintf(&b, "  %-14s %s\n", label, sub.summary)
	}
	b.WriteString("\nThe endpoint speaks streamable HTTP. Point an MCP client at it, e.g.:\n")
	b.WriteString("  claude mcp add --transport http lazymux <endpoint>\n")
	return b.String()
}
