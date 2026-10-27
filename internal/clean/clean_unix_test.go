//go:build !windows

package clean

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/session"
)

// TestGlobalItemAfterChildOfClaude checks session#R-5: a child of claude that runs after
// the session ended does not hold the cache lock, because Go opens files with close-on-exec.
func TestGlobalItemAfterChildOfClaude(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	_, end, err := session.Start(cacheDir, "new", func() {}, func() (string, error) {
		dir := filepath.Join(cacheDir, "plugin", "new")
		return dir, os.MkdirAll(dir, 0o755)
	})
	if err != nil {
		t.Fatal(err)
	}
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
	if err := os.MkdirAll(filepath.Join(cacheDir, "plugin", "old"), 0o755); err != nil {
		t.Fatal(err)
	}
	groups, err := Scan(cacheDir, "new")
	if err != nil || groups[0].Items[0].Reason == "" {
		t.Fatalf("the global item is enabled while the session runs: %v, %v", groups, err)
	}
	end()
	groups, err = Scan(cacheDir, "new")
	if err != nil || len(groups) != 1 || groups[0].Items[0].Reason != "" {
		t.Fatalf("Scan after the session = %v, %v", groups, err)
	}
}
