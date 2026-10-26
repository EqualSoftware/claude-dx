//go:build !windows

package project

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMoveDataFails is the project part of AC-7 of TASK-068: a symlink in tasks/ stops the move, no
// temporary folder stays, and the cache folder stays the data directory.
func TestMoveDataFails(t *testing.T) {
	cacheDir := t.TempDir()
	p := open(t, newRepo(t))
	old := oldDir(cacheDir, p)
	writeFile(t, filepath.Join(old, "tasks", "a.md"), "task")
	if err := os.Symlink("a.md", filepath.Join(old, "tasks", "b.md")); err != nil {
		t.Fatal(err)
	}
	if err := p.MoveData(cacheDir); err == nil {
		t.Fatal("MoveData with a symlink did not fail")
	}
	if matches, _ := filepath.Glob(filepath.Join(p.Common, "dx*")); len(matches) != 0 {
		t.Fatalf("folders stay after a failed move: %v", matches)
	}
	if got := p.DataDir(cacheDir); got != old {
		t.Fatalf("DataDir after a failed move = %q, want %q", got, old)
	}
}

// TestMoveDataKeepsMode is Q-15 of TASK-068.
func TestMoveDataKeepsMode(t *testing.T) {
	cacheDir := t.TempDir()
	p := open(t, newRepo(t))
	path := filepath.Join(oldDir(cacheDir, p), "tasks", "a.md")
	writeFile(t, path, "task")
	if err := os.Chmod(path, 0o664); err != nil {
		t.Fatal(err)
	}
	if err := p.MoveData(cacheDir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(p.Common, "dx", "tasks", "a.md"))
	if err != nil || info.Mode().Perm() != 0o664 {
		t.Fatalf("mode = %v, %v; want 0664", info.Mode().Perm(), err)
	}
}
