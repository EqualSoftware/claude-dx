package docfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// must stops the test if err is not nil.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

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

// TestWriteFails checks docs-common#R-10, R-11: a rename over a directory fails on all platforms.
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

// TestWriteOpenFile checks docs-common#R-10, R-11: Write replaces a file that a reader has open. On Windows the
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

// TestMoveExists checks docs-common#R-9: Move refuses a path that exists and changes no file.
func TestMoveExists(t *testing.T) {
	dir := t.TempDir()
	old, path := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	must(t, os.WriteFile(old, []byte("a"), 0o644))
	must(t, os.WriteFile(path, []byte("b"), 0o644))
	if err := Move(old, path); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("Move = %v, want fs.ErrExist", err)
	}
	a, _ := os.ReadFile(old)
	b, _ := os.ReadFile(path)
	if string(a) != "a" || string(b) != "b" {
		t.Fatalf("Move changed the files: %q, %q", a, b)
	}
}

// TestMoveOpenFile checks docs-common#R-9: Move renames a file that a reader has open. On Windows the
// rename waits until the reader closes the file.
func TestMoveOpenFile(t *testing.T) {
	dir := t.TempDir()
	old, path := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	must(t, os.WriteFile(old, []byte("a"), 0o644))
	f, err := os.Open(old)
	must(t, err)
	done := make(chan error)
	go func() { done <- Move(old, path) }()
	time.Sleep(100 * time.Millisecond)
	_ = f.Close()
	if err := <-done; err != nil {
		t.Fatalf("Move = %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("the old file exists: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "a" {
		t.Fatalf("new file = %q, want a", data)
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

// TestNoLockPath checks docs-common#R-10.
func TestNoLockPath(t *testing.T) {
	for name, run := range map[string]func(string, func() error) error{"Locked": Locked, "RLocked": RLocked} {
		ran := false
		if err := run("", func() error { ran = true; return nil }); !errors.Is(err, errNoLock) || ran {
			t.Fatalf("%s with no path = %v, ran %v; want %v", name, err, ran, errNoLock)
		}
	}
}

func TestRLockedShares(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "docs.lock")
	err := RLocked(lock, func() error {
		// A second read in another goroutine does not wait for this read.
		done := make(chan error, 1)
		go func() { done <- RLocked(lock, func() error { return nil }) }()
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			return errors.New("a read waited for another read")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
