package commands

import "github.com/bkenks/lazymux/internal/config"

// deps is a package-private container for runtime dependencies that
// commands need (config-driven tool paths, etc.). It's set once from
// app.New via SetDeps so we don't have to thread cfg through every
// command signature.
var deps = config.Default()

func SetDeps(cfg config.Config) {
	deps = cfg
}

func cfg() config.Config {
	return deps
}
