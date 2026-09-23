// Package atomicfile replaces a file's contents so a reader never sees a
// partial write.
package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
)

// Write replaces path with data, creating its parent directories as needed.
// The data goes to a temp file in the same directory, which is then renamed
// over path, so a crash or a concurrent writer can't leave it truncated.
func Write(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		return errors.Join(err, tmp.Close(), os.Remove(tmpName))
	}
	if err := tmp.Close(); err != nil {
		return errors.Join(err, os.Remove(tmpName))
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return errors.Join(err, os.Remove(tmpName))
	}
	if err := os.Rename(tmpName, path); err != nil {
		return errors.Join(err, os.Remove(tmpName))
	}
	return nil
}
