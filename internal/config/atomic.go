package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// writePrivateFile replaces a file only after its complete contents have been
// written and synced. The caller creates the parent directory. Concurrent saves
// remain last-writer-wins; atomic replacement does not lock read/modify/write cycles.
func writePrivateFile(path string, data []byte) error {
	// Preserve symlinks used to manage dotfiles instead of replacing the link.
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("resolve destination symlink: %w", err)
		}
	}

	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()

	// CreateTemp creates a private (0600) file on the same filesystem, which
	// allows Rename to atomically replace the destination on macOS and Linux.
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}
	// Persist the directory entry too. If this fails, the new file is already
	// visible, but its durability across a system crash cannot be confirmed.
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync directory after replacing file: %w", err)
	}
	return nil
}
