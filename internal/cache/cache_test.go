package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "claude-dx")
	if err := os.MkdirAll(filepath.Join(dir, "plugin", "v1", ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "update.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	failed, err := Clean(dir, nil)
	if err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("the cache directory exists")
	}
	if failed, err := Clean(dir, nil); err != nil || failed != nil {
		t.Fatalf("Clean of a missing directory = %v, %v", failed, err)
	}
}

func TestHasFailedChild(t *testing.T) {
	dir := filepath.Join("a", "b")
	failed := []string{filepath.Join(dir, "c.exe") + ": in use"}
	if !hasFailedChild(failed, dir) || hasFailedChild(failed, filepath.Join("a", "bb")) {
		t.Fatal("hasFailedChild is wrong")
	}
}

func TestCleanKeeps(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "projects", "a", "docs.lock")
	for _, path := range []string{keep, filepath.Join(dir, "projects", "a", "counters.json"), filepath.Join(dir, "plugin", "v1", "rules.md")} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if failed, err := Clean(dir, []string{keep}); err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "projects", "a"))
	if len(entries) != 1 || entries[0].Name() != "docs.lock" {
		t.Fatalf("projects/a has %v, want only docs.lock", entries)
	}
	if _, err := os.Stat(filepath.Join(dir, "plugin")); err == nil {
		t.Fatal("the plugin directory exists")
	}
}
