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

// Run dispatches `lazymux mcp <subcommand> [args...]`. args excludes the
// "mcp" word itself. It returns an error for the caller to print and exit on.
func Run(args []string, version string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "start":
		return Start(config.Load())
	case "stop":
		return Stop()
	case "serve":
		return runServe(version)
	case "list", "status":
		return runList()
	case "set-url":
		return runSetURL(args[1:])
	case "set-port":
		return runSetPort(args[1:])
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown mcp subcommand %q (try: start, stop, serve, list, set-url, set-port)", args[0])
	}
}

// runServe runs the server in the foreground. `mcp start` re-execs the binary
// this way; run it directly to keep the server attached to a supervisor. The
// pidfile is written only once the port is bound, so its presence means the
// server is genuinely up.
func runServe(version string) error {
	cfg := config.Load()
	defer RemovePID()
	return Serve(context.Background(), cfg, version, func() error {
		if err := WritePID(); err != nil {
			return fmt.Errorf("writing %s: %w", PIDPath(), err)
		}
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

func printUsage() {
	fmt.Print(`lazymux mcp — serve the repo inventory to LLMs over MCP

Usage: lazymux mcp <command>

Commands:
  start          start the server in the background
  stop           stop the background server
  serve          run the server in the foreground (for supervisors / debugging)
  list           show configuration, endpoint, and running status
  set-url <v>    set the bind host — accepts a host, host:port, or full URL
  set-port <n>   set the port

The endpoint speaks streamable HTTP. Point an MCP client at it, e.g.:
  claude mcp add --transport http lazymux <endpoint>
`)
}
