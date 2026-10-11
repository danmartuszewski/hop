package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAtomicallyReplacesPrivateFile(t *testing.T) {
	cfg := &Config{Version: 1, Connections: []Connection{{ID: "new", Host: "new.example.com"}}}
	history := &History{}
	history.RecordUsage("new")
	for name, save := range map[string]func(string) error{"config": cfg.Save, "history": history.SaveToPath} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name+".yaml")
			const original = "original contents\n"
			if err := os.WriteFile(path, []byte(original), 0644); err != nil {
				t.Fatal(err)
			}
			// An open reader must retain the previous contents after replacement.
			// An in-place WriteFile would truncate and overwrite this same inode.
			reader, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			if err := save(path); err != nil {
				t.Fatal(err)
			}
			old, err := io.ReadAll(reader)
			if err != nil || string(old) != original {
				t.Fatalf("previous file was overwritten in place: %q, %v", old, err)
			}
			data, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(data), "new") {
				t.Fatalf("new file contents missing: %q, %v", data, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0600 {
				t.Errorf("permissions = %o, want 600", info.Mode().Perm())
			}
			assertNoTemporaryFiles(t, filepath.Dir(path))
		})
	}
}

func TestAtomicWritePreservesSymlink(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, "dotfiles")
	if err := os.Mkdir(targetDir, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(targetDir, "config.yaml")
	if err := os.WriteFile(target, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "config.yaml")
	if err := os.Symlink(filepath.Join("dotfiles", "config.yaml"), link); err != nil {
		t.Fatal(err)
	}
	if err := writePrivateFile(link, []byte("replacement")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Readlink(link); err != nil {
		t.Fatalf("symlink was replaced: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "replacement" {
		t.Fatalf("symlink target not updated: %q, %v", data, err)
	}
	assertNoTemporaryFiles(t, targetDir)
}

func TestAtomicWriteCleansUpAfterFailedRename(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "config.yaml")
	if err := os.Mkdir(destination, 0700); err != nil {
		t.Fatal(err)
	}
	if err := writePrivateFile(destination, []byte("replacement")); err == nil {
		t.Fatal("expected replacing a directory with a file to fail")
	}
	info, err := os.Stat(destination)
	if err != nil || !info.IsDir() {
		t.Fatalf("existing destination changed: %v", err)
	}
	assertNoTemporaryFiles(t, dir)
}

func TestAtomicWriteRejectsDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "config.yaml")
	if err := os.Symlink("missing.yaml", link); err != nil {
		t.Fatal(err)
	}
	if err := writePrivateFile(link, []byte("replacement")); err == nil {
		t.Fatal("expected an unresolved symlink to fail")
	}
	if _, err := os.Readlink(link); err != nil {
		t.Fatalf("dangling symlink was replaced: %v", err)
	}
	assertNoTemporaryFiles(t, dir)
}

func assertNoTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Errorf("temporary file left behind: %s", entry.Name())
		}
	}
}
