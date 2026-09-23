package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteCreatesParentsAndReplacesContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "file.json")

	if err := Write(path, []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, []byte("second")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second" {
		t.Errorf("contents = %q", data)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
}

func TestWriteFailsWhenTargetIsADirectory(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, []byte("x")); err == nil {
		t.Fatal("expected an error writing over a directory")
	}
	entries, err := os.ReadDir(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}
