//go:build !windows

package session

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// TestCleanAfterChildOfClaude is AC-7 of TASK-061: a child of claude that runs after the session
// ended does not hold the cache lock, because Go opens files with close-on-exec.
func TestCleanAfterChildOfClaude(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	end := start(t, cacheDir)
	// The shell is claude: it starts a child that stays after the shell ends.
	out, err := exec.Command("/bin/sh", "-c", "sleep 30 >/dev/null 2>&1 & echo $!").Output()
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	end()
	if failed, err := Clean(cacheDir); err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
}
