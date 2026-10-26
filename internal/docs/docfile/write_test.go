package docfile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.md")
	for _, text := range []string{"old", "new"} {
		if err := Write(path, []byte(text)); err != nil {
			t.Fatal(err)
		}
	}
	if data, _ := os.ReadFile(path); string(data) != "new" {
		t.Fatalf("file = %q, want new", data)
	}
}

// TestWriteFails is AC-2 of TASK-047: a rename over a directory fails on all platforms.
func TestWriteFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.md")
	if err := os.MkdirAll(filepath.Join(path, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, []byte("new")); err == nil {
		t.Fatal("Write over a directory did not fail")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != "a.md" || !entries[0].IsDir() {
		t.Fatalf("directory has %v, want only a.md", entries)
	}
}

// TestWriteOpenFile is AC-3 of TASK-047: Write replaces a file that a reader has open. On Windows the
// rename waits until the reader closes the file.
func TestWriteOpenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.md")
	if err := Write(path, []byte("old")); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error)
	go func() { done <- Write(path, []byte("new")) }()
	time.Sleep(100 * time.Millisecond)
	_ = f.Close()
	if err := <-done; err != nil {
		t.Fatalf("Write = %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "new" {
		t.Fatalf("file = %q, want new", data)
	}
}

func TestLocked(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "projects", "key", "docs.lock")
	ran := false
	if err := Locked(lock, func() error { ran = true; return nil }); err != nil || !ran {
		t.Fatalf("Locked = %v, ran %v", err, ran)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("Locked did not make the lock file: %v", err)
	}
}
