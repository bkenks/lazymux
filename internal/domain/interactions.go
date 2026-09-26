package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bkenks/gitkeeper/internal/atomicfile"
	"github.com/bkenks/gitkeeper/internal/config"
)

type InteractionStore map[string]time.Time

// interactionsFilePath sits in config.DataDir, so the dev build keeps its own
// recency history.
func interactionsFilePath() string {
	return filepath.Join(config.DataDir(), "interactions.json")
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

// SaveInteraction records the timestamp this repo was last opened. A failure
// only costs the repo its place in the recency sort, so callers report it
// without stopping the action that triggered it.
func SaveInteraction(repoPath string) error {
	store := LoadInteractions()
	store[repoPath] = time.Now()

	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return atomicfile.Write(interactionsFilePath(), data)
}
