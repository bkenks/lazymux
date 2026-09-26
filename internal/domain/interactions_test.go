package domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bkenks/gitkeeper/internal/config"
)

func TestSaveInteractionWritesUnderTheBuildsDirName(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	if err := SaveInteraction("bkenks/gitkeeper"); err != nil {
		t.Fatalf("SaveInteraction: %v", err)
	}
	if err := SaveInteraction("acme/website"); err != nil {
		t.Fatalf("SaveInteraction: %v", err)
	}

	dir := filepath.Join(dataHome, config.DirName())
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	if len(entries) != 1 || entries[0].Name() != "interactions.json" {
		t.Errorf("%s holds %v, want only interactions.json with no temp files left", dir, entries)
	}

	store := LoadInteractions()
	if store["bkenks/gitkeeper"].IsZero() || store["acme/website"].IsZero() {
		t.Errorf("store = %v, want both interactions recorded", store)
	}
}
