package mcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bkenks/lazymux/internal/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// instructions tell the client what this server is for. Routing a request to
// the right repo is the whole point, so say so explicitly rather than leaving
// the model to infer it from three tool names.
const instructions = `lazymux tracks every git repository on this machine: where it lives on disk,
which git forge hosts it, and what it is for.

Use list_repositories or search_repositories to work out which repo a request
refers to before touching the filesystem, then use the returned absolute path.

When you learn what a repo is for and it has no purpose recorded, call
set_repository_purpose so future sessions can route without rediscovering it.`

// listInput is deliberately empty: listing takes no arguments. Filtering is
// search_repositories' job.
type listInput struct{}

type searchInput struct {
	Query string `json:"query" jsonschema:"what the user is trying to do, in their own words"`
}

type listOutput struct {
	BaseDir     string     `json:"baseDir" jsonschema:"root directory all repos live under"`
	Count       int        `json:"count" jsonschema:"number of repos returned"`
	Undescribed []string   `json:"undescribed,omitempty" jsonschema:"keys of returned repos with no purpose recorded yet"`
	Repos       []RepoInfo `json:"repos"`
}

type getInput struct {
	Key string `json:"key" jsonschema:"repo key as '<namespace>/<name>', e.g. 'bkenks/lazymux'"`
}

type setInput struct {
	Key     string `json:"key" jsonschema:"repo key as '<namespace>/<name>'"`
	Purpose string `json:"purpose,omitempty" jsonschema:"one-line summary of what the repo is for"`
	Context string `json:"context,omitempty" jsonschema:"longer notes: stack, conventions, when to use this repo"`
}

// NewServer builds the MCP server with the repo tools registered.
func NewServer(version string) *mcpsdk.Server {
	s := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "lazymux",
		Title:   "lazymux repo inventory",
		Version: version,
	}, &mcpsdk.ServerOptions{Instructions: instructions})

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:  "list_repositories",
		Title: "List repositories",
		Description: "List every repository lazymux manages, with its absolute path and " +
			"recorded purpose. Most-recently-opened repos come first. Use " +
			"search_repositories instead when you already know what you're looking for.",
	}, handleList)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:  "search_repositories",
		Title: "Search repositories",
		Description: "Find the repositories most relevant to a natural-language description " +
			"of a task. Returns only repos matching at least one term, best match first.",
	}, handleSearch)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:  "get_repository",
		Title: "Get repository",
		Description: "Look up one repository by its '<namespace>/<name>' key, returning its " +
			"absolute path, forge links, and recorded purpose and context.",
	}, handleGet)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:  "set_repository_purpose",
		Title: "Set repository purpose",
		Description: "Record what a repository is for. Writes `purpose` and/or `context` into " +
			"lazymux.json so later sessions can route to this repo without rediscovering it. " +
			"Omitted fields keep their current value.",
		Annotations: &mcpsdk.ToolAnnotations{IdempotentHint: true},
	}, handleSet)

	return s
}

func handleList(_ context.Context, _ *mcpsdk.CallToolRequest, _ listInput) (*mcpsdk.CallToolResult, listOutput, error) {
	cfg := config.Load()
	infos, err := inventory(cfg)
	if err != nil {
		return nil, listOutput{}, err
	}
	return nil, newListOutput(cfg, infos), nil
}

func handleSearch(_ context.Context, _ *mcpsdk.CallToolRequest, in searchInput) (*mcpsdk.CallToolResult, listOutput, error) {
	if in.Query == "" {
		return nil, listOutput{}, errors.New("query is required; use list_repositories to see everything")
	}
	cfg := config.Load()
	infos, err := inventory(cfg)
	if err != nil {
		return nil, listOutput{}, err
	}
	return nil, newListOutput(cfg, search(infos, in.Query)), nil
}

func handleGet(_ context.Context, _ *mcpsdk.CallToolRequest, in getInput) (*mcpsdk.CallToolResult, RepoInfo, error) {
	if in.Key == "" {
		return nil, RepoInfo{}, errors.New("key is required")
	}
	cfg := config.Load()
	infos, err := inventory(cfg)
	if err != nil {
		return nil, RepoInfo{}, err
	}
	info, err := find(infos, in.Key)
	return nil, info, err
}

func handleSet(_ context.Context, _ *mcpsdk.CallToolRequest, in setInput) (*mcpsdk.CallToolResult, RepoInfo, error) {
	if in.Key == "" {
		return nil, RepoInfo{}, errors.New("key is required")
	}
	if in.Purpose == "" && in.Context == "" {
		return nil, RepoInfo{}, errors.New("nothing to write: pass purpose, context, or both")
	}
	info, err := setDescription(in.Key, in.Purpose, in.Context)
	return nil, info, err
}

func newListOutput(cfg config.Config, infos []RepoInfo) listOutput {
	out := listOutput{BaseDir: cfg.BaseDir, Count: len(infos), Repos: infos}
	for _, r := range infos {
		if !r.Described() {
			out.Undescribed = append(out.Undescribed, r.Key)
		}
	}
	return out
}

// Serve runs the MCP server in the foreground until the context is cancelled
// or the process is signalled (SIGINT/SIGTERM), which is how `mcp stop` ends
// a detached server.
//
// onListen, if non-nil, runs once the listener is bound and before any request
// is served. A detached server uses it to publish its pidfile, which is what
// tells the parent the bind actually succeeded — inferring that from the port
// being connectable would mistake an unrelated process for our own.
func Serve(ctx context.Context, cfg config.Config, version string, onListen func() error) error {
	server := NewServer(version)
	handler := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return server }, nil)

	mux := http.NewServeMux()
	mux.Handle(cfg.MCP.Path, handler)
	mux.Handle(cfg.MCP.Path+"/", handler)

	ln, err := net.Listen("tcp", cfg.MCP.Addr())
	if err != nil {
		return fmt.Errorf("binding %s: %w", cfg.MCP.Addr(), err)
	}
	if onListen != nil {
		if err := onListen(); err != nil {
			ln.Close()
			return err
		}
	}

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		errc <- srv.Serve(ln)
	}()
	fmt.Fprintf(os.Stderr, "lazymux mcp: listening on %s\n", cfg.MCP.Endpoint())

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fmt.Fprintln(os.Stderr, "lazymux mcp: shutting down")
		return srv.Shutdown(shutdownCtx)
	}
}
