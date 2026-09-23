package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bkenks/lazymux/internal/config"
)

type InteractionStore map[string]time.Time

// interactionsFilePath honors XDG_DATA_HOME, falling back to ~/.local/share.
// The directory follows config.DirName, so the dev build keeps its own
// recency history.
func interactionsFilePath() string {
	dir := config.DirName()
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, dir, "interactions.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", dir+"-interactions.json")
	}
	return filepath.Join(home, ".local", "share", dir, "interactions.json")
}

func LoadInteractions() InteractionStore {
	store := make(InteractionStore)
	data, err := os.ReadFile(interactionsFilePath())
	if err != nil {
		return store
	}
	// Decode failure is not fatal — return an empty store so the app keeps working.
	// A future pass could rename the bad file to preserve user data for inspection.
	_ = json.Unmarshal(data, &store)
	return store
}

// SaveInteraction records the timestamp this repo was last opened. Errors are
// non-fatal: a missing interaction store just means the list won't be sorted
// by recency, which is a graceful degradation.
func SaveInteraction(repoPath string) error {
	store := LoadInteractions()
	store[repoPath] = time.Now()

	path := interactionsFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
