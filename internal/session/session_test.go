package session

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// mkdirs makes the directories in cacheDir.
func mkdirs(t *testing.T, cacheDir string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(cacheDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// start runs Start with an ensure that makes <cacheDir>/plugin/new.
func start(t *testing.T, cacheDir string) func() {
	t.Helper()
	dir, end, err := Start(cacheDir, "new", func() {}, func() (string, error) {
		dir := filepath.Join(cacheDir, "plugin", "new")
		return dir, os.MkdirAll(dir, 0o755)
	})
	if err != nil || filepath.Base(dir) != "new" {
		t.Fatalf("Start = %q, %v", dir, err)
	}
	return end
}

// use holds a shared cache lock, as another claude-dx process does, until the test ends.
func use(t *testing.T, cacheDir string) {
	t.Helper()
	release, err := Use(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = release() })
}

// AC-4: with no other process, only the plugin directory of this session stays.
func TestStartRemovesUnused(t *testing.T) {
	cacheDir := t.TempDir()
	mkdirs(t, cacheDir, "plugin/old", "plugin/new", "plugin/v1.tmp-123")
	defer start(t, cacheDir)()
	entries, _ := os.ReadDir(filepath.Join(cacheDir, "plugin"))
	if len(entries) != 1 || entries[0].Name() != "new" {
		t.Fatalf("plugin has %v, want only new", entries)
	}
}

// AC-4: while another process uses the cache, the old plugin directories stay.
func TestStartKeepsWhileInUse(t *testing.T) {
	cacheDir := t.TempDir()
	mkdirs(t, cacheDir, "plugin/old")
	use(t, cacheDir)
	defer start(t, cacheDir)()
	if !exists(filepath.Join(cacheDir, "plugin", "old")) {
		t.Fatal("Start removed a plugin directory while another process uses the cache")
	}
}

// AC-3: two sessions run at the same time.
func TestStartTwoSessions(t *testing.T) {
	cacheDir := t.TempDir()
	first := start(t, cacheDir)
	defer first()
	start(t, cacheDir)()
}

// AC-2: a start waits while "cache clean" holds the exclusive lock, then opens.
func TestStartWaitsForClean(t *testing.T) {
	cacheDir := t.TempDir()
	clean, err := filelock.TryLock(filepath.Join(cacheDir, lockName))
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan time.Time, 1)
	go func() {
		_, end, err := Start(cacheDir, "new", func() {}, func() (string, error) {
			got <- time.Now()
			return "new", nil
		})
		if err != nil {
			t.Error(err)
			return
		}
		end()
	}()
	time.Sleep(100 * time.Millisecond)
	released := time.Now()
	_ = clean.Close()
	if (<-got).Before(released) {
		t.Fatal("Start called ensure before the exclusive lock was released")
	}
}

// AC-1 and AC-6: Clean refuses a cache that another process uses, and removes nothing.
func TestCleanRefusesInUse(t *testing.T) {
	cacheDir := t.TempDir()
	mkdirs(t, cacheDir, "plugin/v1")
	use(t, cacheDir)
	if _, err := Clean(cacheDir); !errors.Is(err, errInUse) {
		t.Fatalf("Clean = %v, want %v", err, errInUse)
	}
	if !exists(filepath.Join(cacheDir, "plugin", "v1")) {
		t.Fatal("Clean removed files of a cache in use")
	}
}

// AC-5: after Clean, the cache has only the cache lock and the docs locks.
func TestCleanKeepsLockFiles(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	start(t, cacheDir)()
	mkdirs(t, cacheDir, "projects/a", "projects/b")
	for _, name := range []string{"projects/a/docs.lock", "projects/a/counters.json", "update.json"} {
		if err := os.WriteFile(filepath.Join(cacheDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if failed, err := Clean(cacheDir); err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	var files []string
	_ = filepath.WalkDir(cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(cacheDir, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return err
	})
	if want := []string{"cache.lock", "projects/a/docs.lock"}; !slices.Equal(files, want) {
		t.Fatalf("files after Clean = %v, want %v", files, want)
	}
	if exists(filepath.Join(cacheDir, "projects", "b")) {
		t.Fatal("Clean kept a project directory with no docs lock")
	}
}

func TestCleanMissingCache(t *testing.T) {
	if failed, err := Clean(filepath.Join(t.TempDir(), "claude-dx")); err != nil || failed != nil {
		t.Fatalf("Clean of a missing directory = %v, %v", failed, err)
	}
}

func TestStartEnsureError(t *testing.T) {
	cacheDir := t.TempDir()
	if _, _, err := Start(cacheDir, "new", func() {}, func() (string, error) { return "", os.ErrPermission }); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("Start = %v", err)
	}
	// Start released the shared lock, so the exclusive lock is free.
	f, err := filelock.TryLock(filepath.Join(cacheDir, lockName))
	if err != nil {
		t.Fatalf("TryLock after a Start error = %v", err)
	}
	_ = f.Close()
}

// TestStartCallsAlone is the session part of AC-5 and AC-6 of TASK-068: Start calls alone only when
// no other process uses the cache.
func TestStartCallsAlone(t *testing.T) {
	cacheDir := t.TempDir()
	for _, inUse := range []bool{false, true} {
		if inUse {
			use(t, cacheDir)
		}
		called := false
		_, end, err := Start(cacheDir, "new", func() { called = true }, func() (string, error) { return "new", nil })
		if err != nil {
			t.Fatal(err)
		}
		end()
		if called == inUse {
			t.Fatalf("with another process %v, alone called = %v", inUse, called)
		}
	}
}
