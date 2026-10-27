//go:build !windows

package clean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A spec folder that becomes a symbolic link after the scan is skipped. The files behind it stay.
func TestSpecFolderLinkAfterScan(t *testing.T) {
	f := newFixture(t)
	beta := scan(t, f)[2].Items[2:3] // beta In Progress: one task
	outside := t.TempDir()
	kept := writeTask(t, outside, "beta", "four", "In Progress")
	betaDir := filepath.Join(f.data, "tasks", "beta")
	if err := os.RemoveAll(betaDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "tasks", "beta"), betaDir); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return beta, nil }
	if err := Run(&out, f.cache, "own", ui(pick, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(betaDir); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the link of the spec folder was removed: %v", err)
	}
	if !exists(kept) || !strings.Contains(out.String(), "Skipped 1 tasks that changed after the selection.\n") {
		t.Fatalf("kept exists = %v, out = %q", exists(kept), out.String())
	}
}
