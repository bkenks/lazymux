package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Tools struct {
	Ghq     string `toml:"ghq"`
	Lazygit string `toml:"lazygit"`
	Editor  string `toml:"editor"`
	Shell   string `toml:"shell"`
}

type UI struct {
	Theme        string `toml:"theme"`
	ShowFullPath bool   `toml:"show_full_path"`
}

type Behavior struct {
	DefaultProtocol string `toml:"default_protocol"`
	ConfirmDelete   bool   `toml:"confirm_delete"`
}

type Config struct {
	Tools    Tools    `toml:"tools"`
	UI       UI       `toml:"ui"`
	Behavior Behavior `toml:"behavior"`

	// LoadWarning is set when loading produced a recoverable issue
	// (e.g. parse failure → falling back to defaults). Surfaced to the
	// user as a startup toast. Not persisted.
	LoadWarning string `toml:"-"`
}

func Default() Config {
	return Config{
		Tools: Tools{
			Ghq:     "ghq",
			Lazygit: "lazygit",
			Editor:  "codium",
			Shell:   "",
		},
		UI: UI{
			Theme:        "default",
			ShowFullPath: false,
		},
		Behavior: Behavior{
			DefaultProtocol: "https",
			ConfirmDelete:   true,
		},
	}
}

// Path returns the resolved config file path, honoring $XDG_CONFIG_HOME.
func Path() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "lazymux", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "lazymux", "config.toml")
	}
	return filepath.Join(home, ".config", "lazymux", "config.toml")
}

// Load reads the config file, writing a default file on first run.
// On parse errors it returns defaults with a non-empty LoadWarning so the
// caller can surface the issue without crashing.
func Load() Config {
	cfg := Default()
	path := Path()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// First run: try to write a default file but don't block startup if we can't.
		if writeErr := Save(cfg); writeErr != nil {
			cfg.LoadWarning = fmt.Sprintf("couldn't write default config: %v", writeErr)
		}
		return cfg
	}
	if err != nil {
		cfg.LoadWarning = fmt.Sprintf("couldn't read %s: %v", path, err)
		return cfg
	}

	loaded := Default()
	if _, err := toml.Decode(string(data), &loaded); err != nil {
		cfg.LoadWarning = fmt.Sprintf("config invalid, using defaults: %v", err)
		return cfg
	}
	return loaded
}

// Save serializes cfg to disk at Path(), creating parent directories as needed.
func Save(cfg Config) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
