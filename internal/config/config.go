package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// DefaultPlaceholderHost is the fake host stored in every managed repo's
// origin. A per-repo local git `insteadOf` rule rewrites it to the primary
// forge, so the stored remote never changes when the primary forge does.
const DefaultPlaceholderHost = "lazymux-placeholder"

type Tools struct {
	Lazygit string `json:"lazygit"`
	Editor  string `json:"editor"`
	Shell   string `json:"shell"`
}

type UI struct {
	Theme        string `json:"theme"`
	ShowFullPath bool   `json:"showFullPath"`
	// ShowForge is the default visibility of the "forge:" line in the repo
	// list. The list's `g` key toggles it for the session; this is the value
	// restored on launch.
	ShowForge bool `json:"showForge"`
	// ShowStats is the default visibility of the git stats summary (branches,
	// unpushed commits, uncommitted files) on repo list rows. The list's `t`
	// key toggles it for the session; this is the value restored on launch.
	ShowStats bool `json:"showStats"`
}

type Behavior struct {
	// DefaultProtocol is the scheme ("https" | "ssh") used for a freshly
	// cloned repo when it isn't otherwise determined.
	DefaultProtocol string `json:"defaultProtocol"`
	ConfirmDelete   bool   `json:"confirmDelete"`
}

// Forge is a git host in the registry (e.g. {github, github.com}).
type Forge struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// RepoLink records which forges host a managed repo, which one is primary
// (drives the insteadOf rewrite), and the URL scheme used for that repo.
type RepoLink struct {
	Forges  []string `json:"forges"`
	Primary string   `json:"primary"`
	Scheme  string   `json:"scheme"`
}

type Config struct {
	// BaseDir is the root under which repos live as <namespace>/<repo>.
	BaseDir         string `json:"baseDir"`
	PlaceholderHost string `json:"placeholderHost"`

	Tools    Tools    `json:"tools"`
	UI       UI       `json:"ui"`
	Behavior Behavior `json:"behavior"`

	Forges []Forge `json:"forges"`
	// Repos maps a repo key ("<namespace>/<repo>") to its forge links.
	Repos map[string]RepoLink `json:"repos"`

	// LoadWarning is set when loading produced a recoverable issue
	// (e.g. parse failure → falling back to defaults). Surfaced to the
	// user as a startup toast. Not persisted.
	LoadWarning string `json:"-"`
}

func Default() Config {
	return Config{
		BaseDir:         defaultBaseDir(),
		PlaceholderHost: DefaultPlaceholderHost,
		Tools: Tools{
			Lazygit: "lazygit",
			Editor:  "codium",
			Shell:   "",
		},
		UI: UI{
			Theme:        "default",
			ShowFullPath: false,
			ShowForge:    true,
			ShowStats:    true,
		},
		Behavior: Behavior{
			DefaultProtocol: "https",
			ConfirmDelete:   true,
		},
		Forges: []Forge{},
		Repos:  map[string]RepoLink{},
	}
}

// dirName is the name of the directory under $HOME that holds the config
// file and, by default, cloned repos. Overridden at build time via
// -ldflags "-X .../config.dirName=lazymux-dev" to build a dev binary that
// is fully sandboxed from the normal ~/lazymux tree.
var dirName = "lazymux"

func defaultBaseDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", dirName)
	}
	return filepath.Join(home, dirName)
}

// Path returns the resolved config file path. Everything lives in a single
// .lazymux.json at the base dir root, honoring $LAZYMUX_CONFIG for overrides.
func Path() string {
	if p := os.Getenv("LAZYMUX_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(defaultBaseDir(), ".lazymux.json")
}

// Load reads .lazymux.json, migrating a legacy TOML config on first run and
// writing a default file if none exists. On parse errors it returns defaults
// with a non-empty LoadWarning so the caller can surface the issue.
func Load() Config {
	cfg := Default()
	path := Path()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// First run: migrate a legacy config.toml if present, else defaults.
		if migrated, ok := migrateLegacy(cfg); ok {
			cfg = migrated
		}
		if writeErr := Save(cfg); writeErr != nil {
			cfg.LoadWarning = fmt.Sprintf("couldn't write config: %v", writeErr)
		}
		return cfg
	}
	if err != nil {
		cfg.LoadWarning = fmt.Sprintf("couldn't read %s: %v", path, err)
		return cfg
	}

	loaded := Default()
	if err := json.Unmarshal(data, &loaded); err != nil {
		cfg.LoadWarning = fmt.Sprintf("config invalid, using defaults: %v", err)
		return cfg
	}
	return normalize(loaded)
}

// normalize backfills fields an older/partial file may have left empty so the
// rest of the app can assume sane values.
func normalize(cfg Config) Config {
	d := Default()
	if cfg.BaseDir == "" {
		cfg.BaseDir = d.BaseDir
	}
	if cfg.PlaceholderHost == "" {
		cfg.PlaceholderHost = d.PlaceholderHost
	}
	if cfg.Behavior.DefaultProtocol == "" {
		cfg.Behavior.DefaultProtocol = d.Behavior.DefaultProtocol
	}
	if cfg.Repos == nil {
		cfg.Repos = map[string]RepoLink{}
	}
	if cfg.Forges == nil {
		cfg.Forges = []Forge{}
	}
	return cfg
}

// legacyConfig mirrors the old TOML schema for one-time migration.
type legacyConfig struct {
	Tools struct {
		Lazygit string `toml:"lazygit"`
		Editor  string `toml:"editor"`
		Shell   string `toml:"shell"`
	} `toml:"tools"`
	UI struct {
		Theme        string `toml:"theme"`
		ShowFullPath bool   `toml:"show_full_path"`
	} `toml:"ui"`
	Behavior struct {
		DefaultProtocol string `toml:"default_protocol"`
		ConfirmDelete   bool   `toml:"confirm_delete"`
	} `toml:"behavior"`
}

func legacyPath() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "lazymux", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lazymux", "config.toml")
}

// migrateLegacy folds a legacy config.toml into the new Config, preserving the
// user's editor/theme/behavior choices. Returns (cfg, true) only on success.
func migrateLegacy(base Config) (Config, bool) {
	p := legacyPath()
	if p == "" {
		return base, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return base, false
	}
	var old legacyConfig
	if _, err := toml.Decode(string(data), &old); err != nil {
		return base, false
	}
	if old.Tools.Lazygit != "" {
		base.Tools.Lazygit = old.Tools.Lazygit
	}
	if old.Tools.Editor != "" {
		base.Tools.Editor = old.Tools.Editor
	}
	base.Tools.Shell = old.Tools.Shell
	if old.UI.Theme != "" {
		base.UI.Theme = old.UI.Theme
	}
	base.UI.ShowFullPath = old.UI.ShowFullPath
	if old.Behavior.DefaultProtocol != "" {
		base.Behavior.DefaultProtocol = old.Behavior.DefaultProtocol
	}
	base.Behavior.ConfirmDelete = old.Behavior.ConfirmDelete
	return base, true
}

// Save serializes cfg to Path() as indented JSON, creating parents as needed.
func Save(cfg Config) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ForgeByName returns the registry forge with the given name.
func (c Config) ForgeByName(name string) (Forge, bool) {
	for _, f := range c.Forges {
		if f.Name == name {
			return f, true
		}
	}
	return Forge{}, false
}

// ForgeByHost returns the registry forge whose host matches (case-insensitive).
func (c Config) ForgeByHost(host string) (Forge, bool) {
	for _, f := range c.Forges {
		if eqFold(f.Host, host) {
			return f, true
		}
	}
	return Forge{}, false
}

func eqFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
