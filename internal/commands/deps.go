package commands

import "github.com/bkenks/gitkeeper/internal/config"

// deps is a package-private container for runtime dependencies that
// commands need (config-driven tool paths, etc.). The app sets it via SetDeps
// whenever its config changes, so we don't have to thread cfg through every
// command signature. It's only touched from the Bubble Tea update loop:
// commands read it when they're built, never inside the returned closure.
var deps = config.Default()

func SetDeps(cfg config.Config) {
	deps = cfg
}

func cfg() config.Config {
	return deps
}
