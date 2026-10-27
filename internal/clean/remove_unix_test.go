//go:build !windows

package clean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The tasks folder becomes a symbolic link after the scan: the group skips all files, and a new scan
// has no task items.
func TestTasksFolderLinkAfterScan(t *testing.T) {
	f := newFixture(t)
	todo := scan(t, f)[2].Items[:1] // alpha Todo: two tasks
	outside := t.TempDir()
	kept := writeTask(t, outside, "alpha", "one", "Todo")
	tasks := filepath.Join(f.data, "tasks")
	if err := os.RemoveAll(tasks); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "tasks"), tasks); err != nil {
		t.Fatal(err)
	}
	for _, label := range labels(scan(t, f)) {
		if strings.Contains(label, "alpha") {
			t.Fatalf("the scan followed the tasks link: %q", label)
		}
	}
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return todo, nil }
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !exists(kept) || !strings.Contains(out.String(), "Deleted 0 items (0 B).\nSkipped 1 tasks") {
		t.Fatalf("kept exists = %v, out = %q", exists(kept), out.String())
	}
}

// A failed remove gives the partial result with the error.
func TestRemoveAllPartialFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can remove files in a read-only folder")
	}
	root := t.TempDir()
	good := filepath.Join(root, "a", "x")
	bad := filepath.Join(root, "b", "y")
	write(t, good, "12345")
	write(t, bad, "123")
	if err := os.Chmod(filepath.Dir(bad), 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Dir(bad), 0o700) })
	res, err := removeAll([]string{good, bad})
	if err == nil || res.removed != 1 || res.bytes != 5 {
		t.Fatalf("removeAll = %+v, %v; want 1 path, 5 bytes and an error", res, err)
	}
}
