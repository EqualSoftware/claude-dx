package session

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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

// fakeSession writes the session file of pid with dirName. If held, the test holds its lock as a
// running session does.
func fakeSession(t *testing.T, cacheDir, pid, dirName string, held bool) string {
	t.Helper()
	mkdirs(t, cacheDir, "sessions")
	path := filepath.Join(cacheDir, "sessions", pid+".lock")
	if err := os.WriteFile(path, []byte(dirName+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if held {
		f, err := filelock.TryLock(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })
	}
	return path
}

// livePID is the PID of a process that runs for the full test: the parent of the test binary. A held
// session file must have the PID of a running process (TASK-050).
var livePID = strconv.Itoa(os.Getppid())

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// start runs Start with an ensure that makes <cacheDir>/plugin/new.
func start(t *testing.T, cacheDir string) func() {
	t.Helper()
	dir, end, err := Start(cacheDir, func() (string, error) {
		dir := filepath.Join(cacheDir, "plugin", "new")
		return dir, os.MkdirAll(dir, 0o755)
	})
	if err != nil || filepath.Base(dir) != "new" {
		t.Fatalf("Start = %q, %v", dir, err)
	}
	return end
}

// AC-1 and AC-4.
func TestStartRemovesUnused(t *testing.T) {
	cacheDir := t.TempDir()
	mkdirs(t, cacheDir, "plugin/old", "plugin/running", "plugin/v1.tmp-123")
	fakeSession(t, cacheDir, livePID, "running", true)
	ended := fakeSession(t, cacheDir, "200", "old", false)
	end := start(t, cacheDir)
	plugin := filepath.Join(cacheDir, "plugin")
	if exists(filepath.Join(plugin, "old")) || exists(filepath.Join(plugin, "v1.tmp-123")) {
		t.Fatal("Start did not remove the unused plugin directories")
	}
	if !exists(filepath.Join(plugin, "running")) || !exists(filepath.Join(plugin, "new")) {
		t.Fatal("Start removed a plugin directory in use")
	}
	if exists(ended) {
		t.Fatal("Start did not remove the session file of a session that ended")
	}
	own := filepath.Join(cacheDir, "sessions", strconv.Itoa(os.Getpid())+".lock")
	if data, err := os.ReadFile(own); err != nil || string(data) != "new\n" {
		t.Fatalf("session file = %q, %v", data, err)
	}
	end()
	if exists(own) {
		t.Fatal("end did not remove the session file")
	}
}

// AC-2, and AC-3 of TASK-050.
func TestCleanRefusesRunningSession(t *testing.T) {
	cacheDir := t.TempDir()
	mkdirs(t, cacheDir, "plugin/v1")
	fakeSession(t, cacheDir, livePID, "v1", true)
	_, err := Clean(cacheDir)
	if err == nil || err.Error() != "a claude-dx session uses the cache (PID "+livePID+"). Close it and try again" {
		t.Fatalf("Clean = %v", err)
	}
	if !exists(filepath.Join(cacheDir, "plugin", "v1")) {
		t.Fatal("Clean removed files of a running session")
	}
}

// AC-3.
func TestCleanRemovesCache(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	start(t, cacheDir)()
	fakeSession(t, cacheDir, "200", "new", false)
	if failed, err := Clean(cacheDir); err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	if exists(cacheDir) {
		t.Fatal("the cache directory exists")
	}
	if failed, err := Clean(cacheDir); err != nil || failed != nil {
		t.Fatalf("Clean of a missing directory = %v, %v", failed, err)
	}
}

func TestStartEnsureError(t *testing.T) {
	cacheDir := t.TempDir()
	_, _, err := Start(cacheDir, func() (string, error) { return "", os.ErrPermission })
	if err == nil || !strings.Contains(err.Error(), "permission") {
		t.Fatalf("Start = %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(cacheDir, "sessions")); len(entries) != 0 {
		t.Fatalf("Start wrote a session file after an error: %v", entries)
	}
}
