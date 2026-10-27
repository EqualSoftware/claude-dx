//go:build !windows

package clean

import (
	"os"
	"path/filepath"
	"testing"
)

// The scan skips symbolic links to spec folders and to task files.
func TestSymlinksSkipped(t *testing.T) {
	f := newFixture(t)
	outside := t.TempDir()
	linked := writeTask(t, outside, "gamma", "five", "Todo")
	tasks := filepath.Join(f.data, "tasks")
	if err := os.Symlink(filepath.Join(outside, "tasks", "gamma"), filepath.Join(tasks, "gamma")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(linked, filepath.Join(tasks, "alpha", "link.md")); err != nil {
		t.Fatal(err)
	}
	groups := scan(t, f)
	for _, label := range labels(groups) {
		if label == f.root+" | gamma Todo: 1 task" {
			t.Fatalf("the scan followed a spec link: %q", labels(groups))
		}
	}
	if groups[2].Items[0].Label != "alpha Todo: 2 tasks" {
		t.Fatalf("the scan counted a task link: %q", labels(groups))
	}
}
