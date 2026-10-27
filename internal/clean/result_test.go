package clean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// An item whose tasks are all skipped is not counted, and has no bytes.
func TestAllSkippedIsNotDeleted(t *testing.T) {
	f := newFixture(t)
	todo := scan(t, f)[2].Items[:1] // alpha Todo: two tasks
	writeTask(t, f.data, "alpha", "one", "Done")
	writeTask(t, f.data, "alpha", "two", "Done")
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return todo, nil }
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), "Deleted 0 items (0 B).\nSkipped 2 tasks that changed after the selection.\n") {
		t.Fatalf("out = %q", out.String())
	}
}

// With no item to select, Run does not open the selection.
func TestNoDataToDelete(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	write(t, filepath.Join(cacheDir, "plugin", "own", "a.txt"), "own")
	called := false
	pick := func([]Group) ([]Item, error) { called = true; return nil, nil }
	var out bytes.Buffer
	var prompts []string
	if err := Run(&out, cacheDir, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "No data to delete.\n" || called {
		t.Fatalf("out = %q, Select called = %v", out.String(), called)
	}
}

// Only disabled and list-only items give the same text.
func TestNoDataToDeleteOnlyDisabled(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	write(t, filepath.Join(cacheDir, "plugin", "old", "a.txt"), "old")
	write(t, filepath.Join(cacheDir, "projects", "cccccccccccccccc", "x.txt"), "x")
	release, err := filelock.RLock(filepath.Join(cacheDir, "cache.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = release() }()
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { t.Error("Select called"); return nil, nil }
	if err := Run(&out, cacheDir, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "No data to delete.\n" {
		t.Fatalf("out = %q", out.String())
	}
}

// A path that is already gone is not counted.
func TestRemoveAllGonePath(t *testing.T) {
	root := t.TempDir()
	kept := filepath.Join(root, "x")
	write(t, kept, "12345")
	res, err := removeAll([]string{filepath.Join(root, "gone"), kept})
	if err != nil || res.removed != 1 || res.bytes != 5 || exists(kept) {
		t.Fatalf("removeAll = %+v, %v", res, err)
	}
}

// An item whose paths are all gone before the delete is not counted.
func TestItemGoneBeforeDelete(t *testing.T) {
	f := newFixture(t)
	for _, name := range []string{"old1", "old2.tmp-1"} {
		if err := os.RemoveAll(filepath.Join(f.cache, "plugin", name)); err != nil {
			t.Fatal(err)
		}
	}
	// The item exists in the scan, and the paths are gone after it.
	write(t, filepath.Join(f.cache, "plugin", "old1", "a.txt"), "old")
	groups := scan(t, f)
	if err := os.RemoveAll(filepath.Join(f.cache, "plugin", "old1")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return groups[0].Items, nil }
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), "Deleted 0 items (0 B).\n") {
		t.Fatalf("out = %q", out.String())
	}
}
