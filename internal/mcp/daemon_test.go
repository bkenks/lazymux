package mcp

import (
	"net"
	"os"
	"testing"

	"github.com/bkenks/lazymux/internal/config"
)

func writePIDFile(t *testing.T, contents string) {
	t.Helper()
	if err := os.WriteFile(PIDPath(), []byte(contents), 0o644); err != nil {
		t.Fatalf("writing pidfile: %v", err)
	}
}

func TestServeBindConflictKeepsExistingPIDFile(t *testing.T) {
	cfg := newTestWorkspace(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupying a port: %v", err)
	}
	defer ln.Close()
	cfg.MCP.Host = "127.0.0.1"
	cfg.MCP.Port = ln.Addr().(*net.TCPAddr).Port
	if err := config.Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}
	writePIDFile(t, "4242")

	if err := runServe("test"); err == nil {
		t.Fatal("runServe on an occupied port should fail")
	}

	data, err := os.ReadFile(PIDPath())
	if err != nil {
		t.Fatalf("pidfile was removed by a server that never bound: %v", err)
	}
	if string(data) != "4242" {
		t.Errorf("pidfile = %q, want the running server's %q untouched", data, "4242")
	}
}
