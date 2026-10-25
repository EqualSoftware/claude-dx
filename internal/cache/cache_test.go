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
	failed, err := Clean(dir)
	if err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("the cache directory exists")
	}
	if failed, err := Clean(dir); err != nil || failed != nil {
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
