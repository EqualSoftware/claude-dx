//go:build !windows

package launcher

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// TestMoveDataWarning checks project-data#R-2: a failed move writes the warning, and the data stays in the
// cache folder.
func TestMoveDataWarning(t *testing.T) {
	repo, _ := filepath.EvalSymlinks(t.TempDir())
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cacheDir := t.TempDir()
	// Outside git, the data folder of repo was <cache>/projects/<key of repo>.
	if err := os.RemoveAll(filepath.Join(repo, ".git")); err != nil {
		t.Fatal(err)
	}
	before, err := project.Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	old := before.DataDir(cacheDir)
	if err := os.MkdirAll(filepath.Join(old, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("x", filepath.Join(old, "tasks", "link")); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	var stderr bytes.Buffer
	moveData(cacheDir, repo, &stderr)
	after, _ := project.Open(repo)
	dx := filepath.Join(after.Common, "dx")
	want := "claude-dx: cannot move the project data from " + old + " to " + dx + ": " +
		filepath.Join(old, "tasks", "link") + " is not a regular file or a directory. The data stays in " + old + ".\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
	if got := after.DataDir(cacheDir); got != old {
		t.Fatalf("DataDir = %q, want %q", got, old)
	}
}
