package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bkenks/lazymux/internal/atomicfile"
	"github.com/bkenks/lazymux/internal/keybind"
)

// DefaultPlaceholderHost is the fake host stored in every managed repo's
// origin. A per-repo local git `insteadOf` rule rewrites it to the origin
// forge, so the stored remote never changes when the origin forge does.
const DefaultPlaceholderHost = "lazymux-placeholder"

// DefaultSortMode is the repo list ordering used when none is stored. It
// mirrors domain.SortRecent, which config can't import without a cycle.
const DefaultSortMode = "recent"

// Clone schemes a repo link or the default protocol can take.
const (
	SchemeHTTPS = "https"
	SchemeSSH   = "ssh"
)

// Defaults for the MCP server. It binds to loopback so the repo inventory
// isn't exposed to the network unless the user opts in via `mcp set-url`.
const (
	DefaultMCPHost = "127.0.0.1"
	DefaultMCPPort = 7777
	DefaultMCPPath = "/mcp"
)

type Tools struct {
	Editor string `json:"editor"`
	Shell  string `json:"shell"`
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
	// SortMode is the repo list ordering, one of the domain.SortMode values
	// ("recent", "name-asc", "name-desc", "namespace"). The list's `S` key
	// cycles it; this is the value restored on launch. An unknown value is
	// resolved to the default when the app reads it.
	SortMode string `json:"sortMode"`
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

// RepoLink records which forges host a managed repo. Upstreams are every forge
// the repo is pushed to; Origin is the single one the placeholder insteadOf
// rewrite resolves to, making it the fetch/pull URL. Origin is always one of
// Upstreams. Scheme is the URL scheme used for that repo. It also carries the
// human/LLM-facing description of the repo written by the MCP server (see
// internal/mcp).
type RepoLink struct {
	Upstreams []string `json:"upstreams"`
	Origin    string   `json:"origin"`
	Scheme    string   `json:"scheme"`

	// LegacyForges and LegacyPrimary hold the pre-upstream schema. normalize
	// folds them into Upstreams/Origin and clears them, so they disappear from
	// the file on the next Save.
	LegacyForges  []string `json:"forges,omitempty"`
	LegacyPrimary string   `json:"primary,omitempty"`

	// Purpose is a one-line summary of what the repo is for, used to route a
	// natural-language request to the right repo.
	Purpose string `json:"purpose,omitempty"`
	// Context is longer-form detail — stack, conventions, when to reach for
	// this repo over a sibling.
	Context string `json:"context,omitempty"`
}

// Clone returns a copy of the link that shares no slices with the original.
func (l RepoLink) Clone() RepoLink {
	l.Upstreams = slices.Clone(l.Upstreams)
	l.LegacyForges = slices.Clone(l.LegacyForges)
	return l
}

// HasUpstream reports whether name is one of the forges the repo is pushed to.
func (l RepoLink) HasUpstream(name string) bool {
	return slices.Contains(l.Upstreams, name)
}

// ToggleUpstream adds name as an upstream, or removes it if it already is one.
// The first upstream added becomes the origin.
func (l *RepoLink) ToggleUpstream(name string) {
	if l.HasUpstream(name) {
		l.RemoveUpstream(name)
		return
	}
	l.Upstreams = append(slices.Clip(l.Upstreams), name)
	if l.Origin == "" {
		l.Origin = name
	}
}

// RemoveUpstream drops name from the upstreams. When it was the origin, the
// first remaining upstream takes over, or the origin is cleared if none remain.
func (l *RepoLink) RemoveUpstream(name string) {
	isRemoved := func(upstream string) bool { return upstream == name }
	l.Upstreams = slices.DeleteFunc(slices.Clone(l.Upstreams), isRemoved)
	if l.Origin == name {
		l.Origin = ""
		if len(l.Upstreams) > 0 {
			l.Origin = l.Upstreams[0]
		}
	}
}

// SetOrigin makes name the forge the repo is fetched from, adding it to the
// upstreams if it isn't one already.
func (l *RepoLink) SetOrigin(name string) {
	if !l.HasUpstream(name) {
		l.Upstreams = append(slices.Clip(l.Upstreams), name)
	}
	l.Origin = name
}

// RenameUpstream rewrites forge oldName to newName in the upstreams and origin,
// dropping the duplicate if the repo already linked newName.
func (l *RepoLink) RenameUpstream(oldName, newName string) {
	renamed := make([]string, 0, len(l.Upstreams))
	for _, u := range l.Upstreams {
		if u == oldName {
			u = newName
		}
		if !slices.Contains(renamed, u) {
			renamed = append(renamed, u)
		}
	}
	l.Upstreams = renamed
	if l.Origin == oldName {
		l.Origin = newName
	}
}

// ToggleScheme switches the link between https and ssh.
func (l *RepoLink) ToggleScheme() {
	if NormalizeScheme(l.Scheme) == SchemeSSH {
		l.Scheme = SchemeHTTPS
	} else {
		l.Scheme = SchemeSSH
	}
}

// WithForgeLinks returns l with Upstreams, Origin and Scheme taken from src,
// keeping l's Purpose and Context.
func (l RepoLink) WithForgeLinks(src RepoLink) RepoLink {
	l.Upstreams = slices.Clone(src.Upstreams)
	l.Origin = src.Origin
	l.Scheme = src.Scheme
	return l
}

// IsEmpty reports whether the link records neither forges nor a description.
func (l RepoLink) IsEmpty() bool {
	return len(l.Upstreams) == 0 && l.Origin == "" && l.Purpose == "" && l.Context == ""
}

// NormalizeScheme maps any scheme string to SchemeSSH or SchemeHTTPS.
func NormalizeScheme(scheme string) string {
	if strings.EqualFold(scheme, SchemeSSH) {
		return SchemeSSH
	}
	return SchemeHTTPS
}

// MCP configures the MCP server that exposes the repo inventory to LLMs.
type MCP struct {
	// Host is the bind address ("127.0.0.1" to stay local, "0.0.0.0" to expose
	// the server on the network).
	Host string `json:"host"`
	Port int    `json:"port"`
	// Path is the HTTP path the streamable-HTTP endpoint is mounted at.
	Path string `json:"path"`
}

// Endpoint is the full URL clients connect to.
func (m MCP) Endpoint() string {
	return fmt.Sprintf("http://%s%s", m.Addr(), m.Path)
}

// Addr is the host:port pair passed to net.Listen.
func (m MCP) Addr() string {
	return net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
}

// Keybind binds a key combo on the repo list to a shell command that runs in
// the selected repo's directory. Keys holds the canonical keystroke produced by
// keybind.Parse (e.g. "ctrl+shift+k").
type Keybind struct {
	Name    string `json:"name"`
	Keys    string `json:"keys"`
	Command string `json:"command"`
}

type Config struct {
	// BaseDir is the root under which repos live as <namespace>/<repo>.
	BaseDir         string `json:"baseDir"`
	PlaceholderHost string `json:"placeholderHost"`

	Tools    Tools    `json:"tools"`
	UI       UI       `json:"ui"`
	Behavior Behavior `json:"behavior"`
	MCP      MCP      `json:"mcp"`

	Forges   []Forge   `json:"forges"`
	Keybinds []Keybind `json:"keybinds"`
	// Repos maps a repo key ("<namespace>/<repo>") to its forge links.
	Repos map[string]RepoLink `json:"repos"`

	// LoadFailed is set when the config file exists but couldn't be read or
	// parsed. The Config holds defaults, and Update refuses to write over the
	// file until it is fixed. Not persisted.
	LoadFailed bool `json:"-"`

	// Warnings lists recoverable issues found while loading, such as a keybind
	// that won't parse. Surfaced to the user at startup. Not persisted.
	Warnings []string `json:"-"`
}

func Default() Config {
	return Config{
		BaseDir:         defaultBaseDir(),
		PlaceholderHost: DefaultPlaceholderHost,
		Tools: Tools{
			Editor: "codium",
			Shell:  "",
		},
		UI: UI{
			Theme:        "default",
			ShowFullPath: false,
			ShowForge:    true,
			ShowStats:    true,
			SortMode:     DefaultSortMode,
		},
		Behavior: Behavior{
			DefaultProtocol: SchemeHTTPS,
			ConfirmDelete:   true,
		},
		MCP: MCP{
			Host: DefaultMCPHost,
			Port: DefaultMCPPort,
			Path: DefaultMCPPath,
		},
		Forges:   []Forge{},
		Keybinds: []Keybind{},
		Repos:    map[string]RepoLink{},
	}
}

// dirName is the name of the directory under $HOME that holds the config
// file and, by default, cloned repos. Overridden at build time via
// -ldflags "-X .../config.dirName=lazymux-dev" to build a dev binary that
// is fully sandboxed from the normal ~/lazymux tree.
var dirName = "lazymux"

// DirName is the per-build directory name ("lazymux", or "lazymux-dev" for
// the dev binary) that every piece of on-disk state should be keyed by.
func DirName() string { return dirName }

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
// writing a default file if none exists. If the file exists but can't be read
// or parsed, it returns defaults with LoadFailed set.
func Load() Config {
	path := Path()
	cfg, err := readFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg = Default()
		if migrated, ok := migrateLegacy(cfg); ok {
			cfg = migrated
		}
		if writeErr := Save(cfg); writeErr != nil {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("couldn't write config: %v", writeErr))
		}
		return cfg
	}
	if err != nil {
		cfg = Default()
		cfg.LoadFailed = true
		cfg.Warnings = []string{fmt.Sprintf("using defaults, changes won't be saved: %v", err)}
	}
	return cfg
}

// Update re-reads the config file, applies change to it and saves the result,
// so edits another process made since this one loaded survive. It refuses to
// write when the file exists but can't be read or parsed.
func Update(change func(*Config)) (Config, error) {
	path := Path()
	cfg, err := readFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg, err = Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("refusing to overwrite unreadable config: %w", err)
	}
	change(&cfg)
	if err := Save(cfg); err != nil {
		return Config{}, fmt.Errorf("writing %s: %w", path, err)
	}
	return cfg, nil
}

func readFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	loaded := Default()
	if err := json.Unmarshal(data, &loaded); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return normalize(loaded), nil
}

// normalize backfills fields an older/partial file may have left empty so the
// rest of the app can assume sane values.
func normalize(cfg Config) Config {
	d := Default()
	cfg.BaseDir = normalizeBaseDir(cfg.BaseDir, d.BaseDir)
	if cfg.PlaceholderHost == "" {
		cfg.PlaceholderHost = d.PlaceholderHost
	}
	if cfg.UI.SortMode == "" {
		cfg.UI.SortMode = d.UI.SortMode
	}
	cfg.Behavior.DefaultProtocol = strings.ToLower(cfg.Behavior.DefaultProtocol)
	switch cfg.Behavior.DefaultProtocol {
	case SchemeHTTPS, SchemeSSH:
	case "":
		cfg.Behavior.DefaultProtocol = d.Behavior.DefaultProtocol
	default:
		cfg.Warnings = append(cfg.Warnings,
			fmt.Sprintf("defaultProtocol %q isn't https or ssh, using %s",
				cfg.Behavior.DefaultProtocol, d.Behavior.DefaultProtocol))
		cfg.Behavior.DefaultProtocol = d.Behavior.DefaultProtocol
	}
	if cfg.MCP.Host == "" {
		cfg.MCP.Host = d.MCP.Host
	}
	if cfg.MCP.Port == 0 {
		cfg.MCP.Port = d.MCP.Port
	}
	// A hand-edited path like "mcp" or "/mcp/" would otherwise never match the
	// route the server registers.
	cfg.MCP.Path = strings.TrimRight(cfg.MCP.Path, "/")
	if cfg.MCP.Path == "" {
		cfg.MCP.Path = d.MCP.Path
	} else if !strings.HasPrefix(cfg.MCP.Path, "/") {
		cfg.MCP.Path = "/" + cfg.MCP.Path
	}
	if cfg.Repos == nil {
		cfg.Repos = map[string]RepoLink{}
	}
	for key, link := range cfg.Repos {
		cfg.Repos[key] = migrateRepoLink(link)
	}
	if cfg.Forges == nil {
		cfg.Forges = []Forge{}
	}
	if cfg.Keybinds == nil {
		cfg.Keybinds = []Keybind{}
	}
	for i, bind := range cfg.Keybinds {
		keys, err := keybind.Parse(bind.Keys)
		if err != nil {
			cfg.Warnings = append(cfg.Warnings,
				fmt.Sprintf("keybind %q won't run: %v", bind.Name, err))
			continue
		}
		cfg.Keybinds[i].Keys = keys
	}
	return cfg
}

// normalizeBaseDir expands a leading ~ and makes dir absolute and clean, so
// path comparisons against it hold. An empty dir falls back to fallback.
func normalizeBaseDir(dir, fallback string) string {
	if dir == "" {
		return fallback
	}
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, strings.TrimPrefix(dir, "~"))
		}
	}
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return filepath.Clean(dir)
}

// migrateRepoLink folds the pre-upstream forges/primary fields of a repo link
// into Upstreams/Origin, leaving an already-migrated link untouched.
func migrateRepoLink(link RepoLink) RepoLink {
	if len(link.Upstreams) == 0 {
		link.Upstreams = link.LegacyForges
	}
	if link.Origin == "" {
		link.Origin = link.LegacyPrimary
	}
	link.LegacyForges = nil
	link.LegacyPrimary = ""
	return link
}

// legacyConfig mirrors the old TOML schema for one-time migration.
type legacyConfig struct {
	Tools struct {
		Editor string `toml:"editor"`
		Shell  string `toml:"shell"`
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
// The write goes to a temp file in the same directory and is renamed into
// place, so a crash (or the MCP server and the TUI writing at once) can't
// leave a half-written config behind.
func Save(cfg Config) error {
	if cfg.LoadFailed {
		return errors.New("config file couldn't be read at startup; " +
			"fix it and restart before saving")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.Write(Path(), data)
}

// Clone returns a copy of the config that shares no maps or slices with c, so
// it can be read from another goroutine while c is changed.
func (c Config) Clone() Config {
	c.Forges = slices.Clone(c.Forges)
	c.Keybinds = slices.Clone(c.Keybinds)
	c.Warnings = slices.Clone(c.Warnings)
	repos := make(map[string]RepoLink, len(c.Repos))
	for key, link := range c.Repos {
		repos[key] = link.Clone()
	}
	c.Repos = repos
	return c
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
		if strings.EqualFold(f.Host, host) {
			return f, true
		}
	}
	return Forge{}, false
}
