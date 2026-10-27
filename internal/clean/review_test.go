package clean

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// An unknown folder (no path in project.json) is only in the list. A person cannot select it.
func TestUnknownFolderListOnly(t *testing.T) {
	f := newFixture(t)
	other := filepath.Join(f.cache, "projects", "cccccccccccccccc", "x.txt")
	write(t, other, "12345")
	groups := scan(t, f)
	var out bytes.Buffer
	List(&out, groups)
	if !strings.Contains(out.String(), "cccccccccccccccc (unknown)\n  All data  5 B\n") {
		t.Fatalf("list:\n%s", out.String())
	}
	var offered []Group
	pick := func(g []Group) ([]Item, error) { offered = g; return all(g) }
	var prompts []string
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	for _, g := range offered {
		if strings.Contains(g.Name, "unknown") {
			t.Fatalf("Select got the unknown folder: %v", labels(offered))
		}
	}
	if !exists(other) || !strings.HasPrefix(prompts[0], "Delete 5 items") {
		t.Fatalf("an unknown folder was deleted or selected, prompts = %q", prompts)
	}
}

// A task that changes its status after the selection stays, and the output says so.
func TestTaskChangedAfterSelection(t *testing.T) {
	f := newFixture(t)
	groups := scan(t, f)
	todo := groups[2].Items[:1] // alpha Todo: two tasks
	writeTask(t, f.data, "alpha", "one", "Done")
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return todo, nil }
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), "Skipped 1 tasks that changed after the selection.\n") {
		t.Fatalf("out = %q", out.String())
	}
	if !exists(f.tasks[0]) || exists(f.tasks[1]) {
		t.Fatalf("one exists = %v, two exists = %v", exists(f.tasks[0]), exists(f.tasks[1]))
	}
}

// A project delete waits while a process holds the exclusive cache lock (MoveData does).
func TestProjectDeleteWaitsForCacheLock(t *testing.T) {
	f := newFixture(t)
	hold, err := filelock.TryLock(filepath.Join(f.cache, "cache.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hold.Close() }()
	pick := func(groups []Group) ([]Item, error) { return groups[1].Items, nil } // the missing project
	done := make(chan error, 1)
	var out bytes.Buffer
	var prompts []string
	go func() { done <- Run(&out, f.cache, "own", ui(pick, true, &prompts)) }()
	select {
	case err := <-done:
		t.Fatalf("Run did not wait for the cache lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	_ = hold.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if exists(filepath.Join(f.cache, "projects", "aaaaaaaaaaaaaaaa", "other.txt")) {
		t.Fatal("the project data exists after the lock was released")
	}
}

// gitRepo makes a repository and returns its project.
func gitRepo(t *testing.T) project.Project {
	t.Helper()
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	p, err := project.Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// In a repository, the data directory is <Common>/dx. Its docs.lock is not an item, and a delete keeps it.
func TestRepositoryDocsLockStays(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	p := gitRepo(t)
	dx := p.DataDir(cacheDir)
	if dx != filepath.Join(p.Common, "dx") {
		t.Fatalf("data dir = %q", dx)
	}
	writeInfo(t, cacheDir, p.Key, p.Root, "2026-01-01T00:00:00Z")
	path := writeTask(t, dx, "alpha", "one", "Todo")
	write(t, filepath.Join(dx, "docs.lock"), "")
	write(t, filepath.Join(dx, "project.json"), `{"counters":{"adr":1}}`)
	var out bytes.Buffer
	var prompts []string
	if err := Run(&out, cacheDir, "own", ui(all, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if exists(path) || !exists(filepath.Join(dx, "docs.lock")) || !exists(filepath.Join(dx, "project.json")) {
		t.Fatalf("task exists = %v, docs.lock exists = %v, counters exist = %v", exists(path),
			exists(filepath.Join(dx, "docs.lock")), exists(filepath.Join(dx, "project.json")))
	}
}

// While a move to <Common>/dx is pending, the data directory is a cache folder with no path. It is the
// task group of the project and not an unknown folder.
func TestPendingMoveIsNotUnknown(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	p := gitRepo(t)
	sum := sha256.Sum256([]byte(p.Root))
	old := filepath.Join(cacheDir, "projects", hex.EncodeToString(sum[:8]))
	writeTask(t, old, "alpha", "one", "Todo")
	write(t, filepath.Join(old, "project.json"), `{"counters":{"adr":1}}`)
	writeInfo(t, cacheDir, p.Key, p.Root, "2026-01-01T00:00:00Z")
	if got := p.DataDir(cacheDir); got != old {
		t.Fatalf("data dir = %q, want the pending cache folder %q", got, old)
	}
	groups, err := Scan(cacheDir, "own")
	if err != nil {
		t.Fatal(err)
	}
	if got := labels(groups); len(got) != 1 || got[0] != p.Root+" | alpha Todo: 1 task" {
		t.Fatalf("items = %q", got)
	}
	var out bytes.Buffer
	List(&out, groups)
	if strings.Contains(out.String(), "unknown") {
		t.Fatalf("list:\n%s", out.String())
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatal(err)
	}
}
