package app

import (
	"os"
	"path/filepath"
	"testing"
)

// stubEditorOnPath writes an executable named name into a temp dir and makes
// that dir the only entry on PATH for the duration of the test.
func stubEditorOnPath(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("writing stub editor: %v", err)
	}
	t.Setenv("PATH", dir)
	return path
}

func TestValidateEditorCommandAcceptsCommandOnPath(t *testing.T) {
	want := stubEditorOnPath(t, "zed")

	hint, err := validateEditorCommand("zed")

	if err != nil {
		t.Fatalf("validateEditorCommand(zed) = %v, want no error", err)
	}
	if hint != want {
		t.Errorf("hint = %q, want the resolved path %q", hint, want)
	}
}

func TestValidateEditorCommandRejectsMissingCommand(t *testing.T) {
	stubEditorOnPath(t, "zed")

	if _, err := validateEditorCommand("definitely-not-installed"); err == nil {
		t.Fatal("validateEditorCommand accepted a command that is not on PATH")
	}
}

func TestValidateEditorCommandRejectsEmpty(t *testing.T) {
	if _, err := validateEditorCommand(""); err == nil {
		t.Fatal("validateEditorCommand accepted an empty command")
	}
}

func TestValidateEditorCommandRejectsArguments(t *testing.T) {
	stubEditorOnPath(t, "zed")

	if _, err := validateEditorCommand("zed --wait"); err == nil {
		t.Fatal("validateEditorCommand accepted a command with arguments")
	}
}

func TestValidateEditorCommandAcceptsAbsolutePath(t *testing.T) {
	path := stubEditorOnPath(t, "zed")

	hint, err := validateEditorCommand(path)

	if err != nil {
		t.Fatalf("validateEditorCommand(%q) = %v, want no error", path, err)
	}
	if hint != path {
		t.Errorf("hint = %q, want %q", hint, path)
	}
}
