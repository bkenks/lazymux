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

// Colors are the hex base colors ("#7D56F4") the UI palette is derived from.
// An empty one keeps lazymux's default.
type Colors struct {
	Main   string `json:"main"`
	Accent string `json:"accent"`
	Gray   string `json:"gray"`
}

// ColorModes holds separate base colors for dark and light terminal
// backgrounds.
type ColorModes struct {
	Dark  Colors `json:"dark"`
	Light Colors `json:"light"`
}

type UI struct {
	Colors       ColorModes `json:"colors"`
	ShowFullPath bool       `json:"showFullPath"`
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

// ValidatePort reports an error unless port is a usable TCP port.
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d is out of range (1-65535)", port)
	}
	return nil
}

// Addr is the host:port pair passed to net.Listen.
func (m MCP) Addr() string {
	return net.JoinHostPort(m.Host, strconv.Itoa(m.Port))
}

// Keybind binds a key combo on the repo list to a shell command that runs in
// the selected repo's directory. Keys holds the canonical keystroke produced by
// keybind.Parse (e.g. "ctrl+shift+k"). ReturnOnExit goes straight back to
// lazymux when the command ends instead of waiting for enter.
type Keybind struct {
	Name         string `json:"name"`
	Keys         string `json:"keys"`
	Command      string `json:"command"`
	ReturnOnExit bool   `json:"returnOnExit"`
}

type Config struct {
	// ReposDir is the configured directory repos live under as
	// <namespace>/<repo>. RepoRoot resolves the directory actually used.
	ReposDir string `json:"reposDir,omitempty"`
	// LegacyBaseDir holds the pre-reposDir key. normalize folds it into
	// ReposDir, and it is never written back.
	LegacyBaseDir   string `json:"baseDir,omitempty"`
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
		PlaceholderHost: DefaultPlaceholderHost,
		Tools: Tools{
			Editor: "codium",
			Shell:  "",
		},
		UI: UI{
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

// dirName is the name of the directory under each XDG base directory that
// holds lazymux's config and data. Overridden at build time via
// -ldflags "-X .../config.dirName=lazymux-dev" to build a dev binary that
// is fully sandboxed from the normal lazymux directories.
var dirName = "lazymux"

// DirName is the per-build directory name ("lazymux", or "lazymux-dev" for
// the dev binary) that every piece of on-disk state should be keyed by.
func DirName() string { return dirName }

// xdgDir returns the directory named by the XDG base directory variable
// envVar, or $HOME joined with fallback when it is unset or not absolute, as
// the XDG spec requires.
func xdgDir(envVar string, fallback ...string) string {
	if dir := os.Getenv(envVar); filepath.IsAbs(dir) {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(append([]string{home}, fallback...)...)
}

// DataDir is lazymux's directory under $XDG_DATA_HOME (~/.local/share).
func DataDir() string {
	return filepath.Join(xdgDir("XDG_DATA_HOME", ".local", "share"), dirName)
}

// ReposEnvVar names the environment variable that overrides reposDir.
const ReposEnvVar = "LAZYMUX_REPOS"

// RepoRoot is the directory repos live under: $LAZYMUX_REPOS, then the
// configured reposDir. It is empty when neither is set.
func (c Config) RepoRoot() string {
	if dir := os.Getenv(ReposEnvVar); dir != "" {
		return normalizeDir(dir)
	}
	return c.ReposDir
}

// ValidateRepoRoot reports why RepoRoot can't hold repos: it is unset, missing
// or not a directory.
func (c Config) ValidateRepoRoot() error {
	root := c.RepoRoot()
	if root == "" {
		return errors.New("no repo directory is set")
	}
	info, err := os.Stat(root)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("repo directory %s doesn't exist", root)
	}
	if err != nil {
		return fmt.Errorf("checking repo directory %s: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("repo directory %s isn't a directory", root)
	}
	return nil
}

// ParseReposDir turns a user-entered repo directory into the clean absolute
// path to store, expanding a leading ~. It rejects empty and relative input
// and an existing path that isn't a directory.
func ParseReposDir(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("enter a directory")
	}
	dir := normalizeDir(input)
	if !filepath.IsAbs(expandHome(input)) {
		return "", errors.New("use an absolute path or one starting with ~")
	}
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		return "", fmt.Errorf("%s isn't a directory", dir)
	}
	return dir, nil
}

// Path returns the resolved config file path: config.json under
// $XDG_CONFIG_HOME (~/.config), honoring $LAZYMUX_CONFIG for overrides.
func Path() string {
	if p := os.Getenv("LAZYMUX_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(xdgDir("XDG_CONFIG_HOME", ".config"), dirName, "config.json")
}

// legacyJSONPath is where configs lived before moving under XDG_CONFIG_HOME.
func legacyJSONPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, dirName, ".lazymux.json")
}

// Load reads the config file, moving a legacy ~/lazymux/.lazymux.json into
// place or migrating a legacy TOML config on first run, and writing a default
// file if none exists. If the file exists but can't be read or parsed, it
// returns defaults with LoadFailed set.
func Load() Config {
	path := Path()
	cfg, err := readFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if moved, ok := moveLegacyJSON(path); ok {
			return moved
		}
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
		if moved, ok := moveLegacyJSON(path); ok {
			cfg = moved
		}
	}
	if err != nil {
		return Config{}, fmt.Errorf("refusing to overwrite unreadable config: %w", err)
	}
	if cfg.LoadFailed {
		return Config{}, fmt.Errorf("refusing to overwrite unreadable config: %s", cfg.Warnings[0])
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
	if cfg.ReposDir == "" {
		cfg.ReposDir = cfg.LegacyBaseDir
	}
	cfg.LegacyBaseDir = ""
	if cfg.ReposDir != "" {
		cfg.ReposDir = normalizeDir(cfg.ReposDir)
	}
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
	} else if err := ValidatePort(cfg.MCP.Port); err != nil {
		cfg.Warnings = append(cfg.Warnings,
			fmt.Sprintf("mcp %v, using %d", err, d.MCP.Port))
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

// expandHome replaces a leading ~ in dir with the home directory.
func expandHome(dir string) string {
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir, "~"))
		}
	}
	return dir
}

// normalizeDir expands a leading ~ and makes dir absolute and clean, so path
// comparisons against it hold.
func normalizeDir(dir string) string {
	dir = expandHome(dir)
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
		ShowFullPath bool `toml:"show_full_path"`
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

// moveLegacyJSON moves a config from legacyJSONPath to path, keeping its repos
// where they are. It reports false when there is nothing to move or
// $LAZYMUX_CONFIG picks the path. A legacy file that can't be read yields
// defaults with LoadFailed set, so a fresh config never buries it.
func moveLegacyJSON(path string) (Config, bool) {
	legacy := legacyJSONPath()
	if legacy == "" || os.Getenv("LAZYMUX_CONFIG") != "" {
		return Config{}, false
	}
	cfg, err := readFile(legacy)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, false
	}
	if err != nil {
		cfg = Default()
		cfg.LoadFailed = true
		cfg.Warnings = []string{fmt.Sprintf(
			"using defaults, changes won't be saved: %v (fix it to move it to %s)", err, path)}
		return cfg, true
	}
	if cfg.ReposDir == "" {
		cfg.ReposDir = filepath.Dir(legacy)
	}
	if err := Save(cfg); err != nil {
		cfg.Warnings = append(cfg.Warnings,
			fmt.Sprintf("couldn't move config from %s to %s: %v", legacy, path, err))
		return cfg, true
	}
	if err := os.Remove(legacy); err != nil {
		cfg.Warnings = append(cfg.Warnings,
			fmt.Sprintf("config moved to %s, but couldn't remove %s: %v", path, legacy, err))
	}
	return cfg, true
}

// migrateLegacy folds a legacy config.toml into the new Config, preserving the
// user's editor/UI/behavior choices. Returns (cfg, true) only on success.
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
