package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// escQuitOverlay is the lazygit config lazymux layers on top of the user's own
// when Tools.LazygitEscQuit is on: esc at the top level leaves lazygit, so it
// reads as "back" the same way it does everywhere in lazymux.
const escQuitOverlay = "quitOnTopLevelReturn: true\n"

func LazygitCmd(absPath string) tea.Cmd {
	bin := cfg().Tools.Lazygit
	args := []string{"-p", absPath}

	if cfg().Tools.LazygitEscQuit {
		if list, err := escQuitConfigList(bin); err == nil {
			args = append(args, "--use-config-file", list)
		}
	}

	return TeaCmdBuilder(bin, args...)
}

// escQuitConfigList returns the comma separated list handed to lazygit's
// --use-config-file: every config lazygit would have loaded on its own, with
// the lazymux overlay last so its keys win. The user's own config is never
// modified.
func escQuitConfigList(bin string) (string, error) {
	overlay, err := writeEscQuitOverlay()
	if err != nil {
		return "", err
	}
	return strings.Join(append(userLazygitConfigs(bin), overlay), ","), nil
}

func writeEscQuitOverlay() (string, error) {
	path := escQuitOverlayPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if current, err := os.ReadFile(path); err == nil && string(current) == escQuitOverlay {
		return path, nil
	}
	if err := os.WriteFile(path, []byte(escQuitOverlay), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// escQuitOverlayPath honors XDG_DATA_HOME, falling back to ~/.local/share.
func escQuitOverlayPath() string {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "lazymux", "lazygit-esc-quit.yml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "lazymux-lazygit-esc-quit.yml")
	}
	return filepath.Join(home, ".local", "share", "lazymux", "lazygit-esc-quit.yml")
}

// userLazygitConfigs lists the configs lazygit loads on its own: LG_CONFIG_FILE
// when set, otherwise config.yml in the directory lazygit reports. Paths that
// don't exist are dropped, since lazygit refuses a missing config file.
func userLazygitConfigs(bin string) []string {
	var candidates []string
	if env := os.Getenv("LG_CONFIG_FILE"); env != "" {
		candidates = strings.Split(env, ",")
	} else if dir, err := lazygitConfigDir(bin); err == nil {
		candidates = []string{filepath.Join(dir, "config.yml")}
	}

	configs := make([]string, 0, len(candidates))
	for _, path := range candidates {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			configs = append(configs, path)
		}
	}
	return configs
}

func lazygitConfigDir(bin string) (string, error) {
	out, err := exec.Command(bin, "--print-config-dir").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
