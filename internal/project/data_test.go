package project

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile makes the parent directories of path and writes data.
func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func open(t *testing.T, dir string) Project {
	t.Helper()
	p, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// oldDir returns the cache folder of the top level of p.
func oldDir(cacheDir string, p Project) string {
	return filepath.Join(cacheDir, "projects", key(p.Root))
}

// AC-1 of TASK-068.
func TestDataDir(t *testing.T) {
	cacheDir := t.TempDir()
	repo := newRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	run(t, repo, "worktree", "add", "-q", "-b", "other", wt)
	want := filepath.Join(open(t, repo).Common, "dx")
	for _, dir := range []string{repo, wt} {
		if got := open(t, dir).DataDir(cacheDir); got != want {
			t.Fatalf("DataDir of %s = %q, want %q", dir, got, want)
		}
	}
	outside, _ := filepath.EvalSymlinks(t.TempDir())
	p := open(t, outside)
	if got := p.DataDir(cacheDir); got != filepath.Join(cacheDir, "projects", p.Key) {
		t.Fatalf("DataDir outside git = %q", got)
	}
}

// AC-2 of TASK-068.
func TestDataDirPendingMove(t *testing.T) {
	cacheDir := t.TempDir()
	p := open(t, newRepo(t))
	old := oldDir(cacheDir, p)
	writeFile(t, filepath.Join(old, "docs.lock"), "")
	if got := p.DataDir(cacheDir); got != filepath.Join(p.Common, "dx") {
		t.Fatalf("DataDir with only docs.lock = %q, want dx", got)
	}
	writeFile(t, filepath.Join(old, "project.json"), "{}")
	if got := p.DataDir(cacheDir); got != old {
		t.Fatalf("DataDir with data in the cache = %q, want %q", got, old)
	}
}

// AC-5 of TASK-068.
func TestMoveData(t *testing.T) {
	cacheDir := t.TempDir()
	p := open(t, newRepo(t))
	old := oldDir(cacheDir, p)
	writeFile(t, filepath.Join(old, "docs.lock"), "")
	writeFile(t, filepath.Join(old, "project.json"), `{"counters":{"adr":3}}`)
	writeFile(t, filepath.Join(old, "tasks", "general", "a.md"), "task")
	writeFile(t, filepath.Join(p.Common, "dx.tmp-old", "x"), "garbage")
	if err := p.MoveData(cacheDir); err != nil {
		t.Fatal(err)
	}
	dx := filepath.Join(p.Common, "dx")
	if p.DataDir(cacheDir) != dx {
		t.Fatal("DataDir is not dx after the move")
	}
	if data, _ := os.ReadFile(filepath.Join(dx, "tasks", "general", "a.md")); string(data) != "task" {
		t.Fatalf("moved task = %q", data)
	}
	if data, _ := os.ReadFile(filepath.Join(dx, "project.json")); string(data) != `{"counters":{"adr":3}}` {
		t.Fatalf("moved project.json = %q", data)
	}
	if entries, _ := os.ReadDir(old); len(entries) != 1 || entries[0].Name() != "docs.lock" {
		t.Fatalf("old folder has %v, want only docs.lock", entries)
	}
	if matches, _ := filepath.Glob(filepath.Join(p.Common, "dx.tmp-*")); len(matches) != 0 {
		t.Fatalf("temporary folders stay: %v", matches)
	}
	// A second call has nothing to move.
	if err := p.MoveData(cacheDir); err != nil {
		t.Fatal(err)
	}
}

func TestMoveDataOutsideGit(t *testing.T) {
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	if err := open(t, dir).MoveData(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
