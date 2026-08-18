package ui_plugins

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFileAtomic replaces path with data by writing a temp file in the same
// directory and renaming it into place. When path already exists, the
// replacement keeps the original permission bits; otherwise 0644 is used.
func writeFileAtomic(path string, data []byte) error {
	perm, err := filePermissions(path, 0644)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", filepath.Base(path), err)
	}

	tmpPath := tmp.Name()
	succeeded := false
	defer func() {
		if !succeeded {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set permissions on temp file for %s: %w", filepath.Base(path), err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file for %s: %w", filepath.Base(path), err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp file for %s: %w", filepath.Base(path), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", filepath.Base(path), err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}

	succeeded = true
	return nil
}

func filePermissions(path string, defaultPerm os.FileMode) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultPerm, nil
		}
		return 0, fmt.Errorf("stat %s: %w", filepath.Base(path), err)
	}
	return info.Mode().Perm(), nil
}
