//go:build !windows

package docfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteKeepsMode is AC-1 and AC-2 of TASK-058.
func TestWriteKeepsMode(t *testing.T) {
	dir := t.TempDir()
	old, path := filepath.Join(dir, "old.md"), filepath.Join(dir, "new.md")
	must(t, os.WriteFile(old, []byte("a"), 0o600))
	must(t, os.Chmod(old, 0o664)) // Chmod does not use the umask.
	for file, want := range map[string]fs.FileMode{old: 0o664, path: 0o644} {
		must(t, Write(file, []byte("b")))
		if info, err := os.Stat(file); err != nil || info.Mode().Perm() != want {
			t.Fatalf("mode of %s = %v, %v; want %v", filepath.Base(file), info.Mode().Perm(), err, want)
		}
	}
}

// TestWriteStatFails is AC-3 of TASK-058: a directory with no search permission makes Stat fail.
func TestWriteStatFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores the directory permission")
	}
	dir := filepath.Join(t.TempDir(), "docs")
	must(t, os.Mkdir(dir, 0o755))
	must(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("a"), 0o644))
	must(t, os.Chmod(dir, 0o600))
	err := Write(filepath.Join(dir, "a.md"), []byte("b"))
	must(t, os.Chmod(dir, 0o755))
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("Write = %v, want a permission error", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("directory has %v, want only a.md", entries)
	}
}

// TestWriteStatLoop checks FR-2 of TASK-058 in a writable directory: a symlink loop makes Stat fail,
// and Write makes no temporary file.
func TestWriteStatLoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.md")
	must(t, os.Symlink(path, path))
	if err := Write(path, []byte("b")); err == nil {
		t.Fatal("Write through a symlink loop did not fail")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("directory has %v, want only a.md", entries)
	}
}
