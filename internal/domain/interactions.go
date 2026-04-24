package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type InteractionStore map[string]time.Time

func interactionsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "lazymux", "interactions.json")
}

func LoadInteractions() InteractionStore {
	store := make(InteractionStore)
	data, err := os.ReadFile(interactionsFilePath())
	if err != nil {
		return store
	}
	_ = json.Unmarshal(data, &store)
	return store
}

func SaveInteraction(repoPath string) {
	store := LoadInteractions()
	store[repoPath] = time.Now()
	path := interactionsFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(store)
	_ = os.WriteFile(path, data, 0644)
}
