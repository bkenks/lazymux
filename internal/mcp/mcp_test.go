package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bkenks/lazymux/internal/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestApplyURL(t *testing.T) {
	base := config.MCP{Host: "127.0.0.1", Port: 7777, Path: "/mcp"}

	tests := []struct {
		name    string
		raw     string
		want    config.MCP
		wantErr bool
	}{
		{
			name: "bare host keeps port and path",
			raw:  "0.0.0.0",
			want: config.MCP{Host: "0.0.0.0", Port: 7777, Path: "/mcp"},
		},
		{
			name: "host and port",
			raw:  "0.0.0.0:8080",
			want: config.MCP{Host: "0.0.0.0", Port: 8080, Path: "/mcp"},
		},
		{
			name: "full url overrides everything",
			raw:  "http://192.168.1.5:9000/lazymux",
			want: config.MCP{Host: "192.168.1.5", Port: 9000, Path: "/lazymux"},
		},
		{
			name: "trailing slash is trimmed",
			raw:  "http://localhost:9000/mcp/",
			want: config.MCP{Host: "localhost", Port: 9000, Path: "/mcp"},
		},
		{
			name: "bare host with root path keeps existing path",
			raw:  "http://localhost/",
			want: config.MCP{Host: "localhost", Port: 7777, Path: "/mcp"},
		},
		{name: "https is rejected", raw: "https://localhost:9000", wantErr: true},
		{name: "empty is rejected", raw: "   ", wantErr: true},
		{name: "out of range port is rejected", raw: "localhost:99999", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := applyURL(base, tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("applyURL(%q) = %+v, want error", tc.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("applyURL(%q): %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("applyURL(%q) = %+v, want %+v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestEndpoint(t *testing.T) {
	m := config.MCP{Host: "127.0.0.1", Port: 8080, Path: "/mcp"}
	if got, want := m.Endpoint(), "http://127.0.0.1:8080/mcp"; got != want {
		t.Errorf("Endpoint() = %q, want %q", got, want)
	}
}

func TestSearchRanksNameOverContext(t *testing.T) {
	infos := []RepoInfo{
		{Key: "bkenks/notes", Name: "notes", Context: "sometimes mentions the deploy scripts repo"},
		{Key: "bkenks/deploy", Name: "deploy", Purpose: "compose stacks for the homelab"},
		{Key: "bkenks/unrelated", Name: "unrelated"},
	}
	got := search(infos, "deploy")
	if len(got) != 2 {
		t.Fatalf("search returned %d repos, want 2 (the non-matching one must be dropped): %+v", len(got), got)
	}
	if got[0].Key != "bkenks/deploy" {
		t.Errorf("best match = %q, want %q (a name hit must outrank a context mention)", got[0].Key, "bkenks/deploy")
	}
}

func TestSearchMultiTermPrefersRepoMatchingBoth(t *testing.T) {
	infos := []RepoInfo{
		{Key: "bkenks/api", Name: "api", Purpose: "public REST api"},
		{Key: "bkenks/docs", Name: "docs", Purpose: "documentation for the public REST api"},
	}
	got := search(infos, "docs api")
	if len(got) == 0 {
		t.Fatal("search returned nothing")
	}
	if got[0].Key != "bkenks/docs" {
		t.Errorf("best match = %q, want %q", got[0].Key, "bkenks/docs")
	}
}

func TestSearchEmptyQueryReturnsEverything(t *testing.T) {
	infos := []RepoInfo{{Key: "a/b", Name: "b"}, {Key: "c/d", Name: "d"}}
	if got := search(infos, "   "); len(got) != 2 {
		t.Errorf("search with a blank query returned %d repos, want 2", len(got))
	}
}

func TestFindSuggestsNearMisses(t *testing.T) {
	infos := []RepoInfo{{Key: "bkenks/lazymux", Name: "lazymux"}}

	if _, err := find(infos, "lazymux"); err == nil {
		t.Fatal("find with a bare name should fail — keys are namespaced")
	} else if got := err.Error(); !strings.Contains(got, "bkenks/lazymux") {
		t.Errorf("error %q should suggest the real key", got)
	}

	if _, err := find(infos, "totally/absent"); err == nil {
		t.Fatal("find with an unknown key should fail")
	}

	if got, err := find(infos, "bkenks/lazymux"); err != nil || got.Name != "lazymux" {
		t.Errorf("find(exact) = (%+v, %v), want the lazymux repo", got, err)
	}
}

// newTestWorkspace points $LAZYMUX_CONFIG at a temp dir and creates the named
// repos under it, so tests exercise the real config and repo-scanning paths.
func newTestWorkspace(t *testing.T, keys ...string) config.Config {
	t.Helper()
	base := t.TempDir()
	t.Setenv("LAZYMUX_CONFIG", filepath.Join(base, ".lazymux.json"))

	for _, key := range keys {
		dir := filepath.Join(base, filepath.FromSlash(key))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("creating %s: %v", dir, err)
		}
		cmd := exec.Command("git", "init", "--quiet", dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init %s: %v: %s", dir, err, out)
		}
	}

	cfg := config.Default()
	cfg.BaseDir = base
	if err := config.Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}
	return cfg
}

func TestInventoryFindsReposAndPurposes(t *testing.T) {
	cfg := newTestWorkspace(t, "bkenks/lazymux", "acme/website")
	cfg.Repos["bkenks/lazymux"] = config.RepoLink{Purpose: "the TUI repo manager"}

	infos, err := inventory(cfg)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("inventory found %d repos, want 2: %+v", len(infos), infos)
	}

	got, err := find(infos, "bkenks/lazymux")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Purpose != "the TUI repo manager" {
		t.Errorf("Purpose = %q, want the value from config", got.Purpose)
	}
	if got.Namespace != "bkenks" {
		t.Errorf("Namespace = %q, want %q", got.Namespace, "bkenks")
	}
	if got.Path != filepath.Join(cfg.BaseDir, "bkenks", "lazymux") {
		t.Errorf("Path = %q, want the absolute on-disk path", got.Path)
	}
}

func TestSetDescriptionPersistsAndMerges(t *testing.T) {
	cfg := newTestWorkspace(t, "bkenks/lazymux")
	// A pre-existing forge link must survive a purpose write.
	cfg.Repos["bkenks/lazymux"] = config.RepoLink{
		Forges:  []string{"github"},
		Primary: "github",
		Scheme:  "https",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	if _, err := setDescription("bkenks/lazymux", "the TUI repo manager", "Go, bubbletea"); err != nil {
		t.Fatalf("setDescription: %v", err)
	}

	reloaded := config.Load()
	link := reloaded.Repos["bkenks/lazymux"]
	if link.Purpose != "the TUI repo manager" {
		t.Errorf("Purpose = %q, want it persisted", link.Purpose)
	}
	if link.Context != "Go, bubbletea" {
		t.Errorf("Context = %q, want it persisted", link.Context)
	}
	if link.Primary != "github" {
		t.Errorf("Primary = %q, want the pre-existing forge link preserved", link.Primary)
	}

	// Writing only a context must leave the purpose alone.
	if _, err := setDescription("bkenks/lazymux", "", "now with tests"); err != nil {
		t.Fatalf("setDescription (context only): %v", err)
	}
	link = config.Load().Repos["bkenks/lazymux"]
	if link.Purpose != "the TUI repo manager" {
		t.Errorf("Purpose = %q, want it untouched by a context-only write", link.Purpose)
	}
	if link.Context != "now with tests" {
		t.Errorf("Context = %q, want it updated", link.Context)
	}
}

func TestSetDescriptionRejectsUnknownRepo(t *testing.T) {
	newTestWorkspace(t, "bkenks/lazymux")
	if _, err := setDescription("bkenks/does-not-exist", "nope", ""); err == nil {
		t.Fatal("setDescription on an unmanaged key should fail rather than invent an entry")
	}
	if _, ok := config.Load().Repos["bkenks/does-not-exist"]; ok {
		t.Error("a rejected write must not leave an entry in the config")
	}
}

// TestEndToEnd drives the real MCP server over HTTP with the SDK client,
// exercising the same path a client like Claude Code takes.
func TestEndToEnd(t *testing.T) {
	newTestWorkspace(t, "bkenks/lazymux", "acme/website")

	handler := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return NewServer("test") }, nil)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: srv.URL}, nil)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("listing tools: %v", err)
	}
	wantTools := map[string]bool{
		"list_repositories":      false,
		"search_repositories":    false,
		"get_repository":         false,
		"set_repository_purpose": false,
	}
	for _, tool := range tools.Tools {
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("tool %q was not advertised", name)
		}
	}

	var listed listOutput
	callInto(t, ctx, session, "list_repositories", map[string]any{}, &listed)
	if listed.Count != 2 {
		t.Errorf("list_repositories returned %d repos, want 2", listed.Count)
	}
	if len(listed.Undescribed) != 2 {
		t.Errorf("undescribed = %v, want both repos flagged", listed.Undescribed)
	}

	var written RepoInfo
	callInto(t, ctx, session, "set_repository_purpose", map[string]any{
		"key":     "bkenks/lazymux",
		"purpose": "terminal UI for managing git repos",
	}, &written)
	if written.Purpose != "terminal UI for managing git repos" {
		t.Errorf("set_repository_purpose returned Purpose %q", written.Purpose)
	}

	var found listOutput
	callInto(t, ctx, session, "search_repositories", map[string]any{
		"query": "terminal ui for git",
	}, &found)
	if found.Count == 0 {
		t.Fatal("search_repositories found nothing after a purpose was recorded")
	}
	if found.Repos[0].Key != "bkenks/lazymux" {
		t.Errorf("best match = %q, want the repo whose purpose matches", found.Repos[0].Key)
	}

	var one RepoInfo
	callInto(t, ctx, session, "get_repository", map[string]any{"key": "acme/website"}, &one)
	if one.Name != "website" {
		t.Errorf("get_repository returned %q, want website", one.Name)
	}

	// A bad key must come back as a tool error, not a protocol failure.
	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "get_repository",
		Arguments: map[string]any{"key": "nope/nope"},
	})
	if err != nil {
		t.Fatalf("calling get_repository with a bad key errored at the protocol level: %v", err)
	}
	if !res.IsError {
		t.Error("get_repository with an unknown key should return IsError")
	}
}

func callInto(t *testing.T, ctx context.Context, session *mcpsdk.ClientSession, name string, args map[string]any, out any) {
	t.Helper()
	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("calling %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("calling %s: tool error: %+v", name, res.Content)
	}
	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("re-marshaling %s output: %v", name, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decoding %s output: %v", name, err)
	}
}
