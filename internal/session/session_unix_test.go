//go:build !windows

package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// TestExecHelper is the helper process of TestLockStaysAfterExec. It holds the lock of the file in
// DX_EXEC_LOCK as Start does, and runs exec of a shell that waits for a line on stdin.
func TestExecHelper(t *testing.T) {
	path := os.Getenv("DX_EXEC_LOCK")
	if path == "" {
		t.Skip("helper process")
	}
	if _, err := hold(path, "v1"); err != nil {
		t.Fatal(err)
	}
	err := syscall.Exec("/bin/sh", []string{"sh", "-c", "echo ready; read line || true"}, os.Environ())
	t.Fatal(err)
}

// TestLockHelper is a helper process: it holds the lock of the file in DX_HOLD_LOCK until stdin
// closes, as a child of claude that runs after its session ended.
func TestLockHelper(t *testing.T) {
	path := os.Getenv("DX_HOLD_LOCK")
	if path == "" {
		t.Skip("helper process")
	}
	f, err := filelock.TryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	fmt.Println("ready")
	_, _ = io.ReadAll(os.Stdin)
}

// holdInOther locks path in a helper process until the test ends.
func holdInOther(t *testing.T, path string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockHelper$")
	cmd.Env = append(os.Environ(), "DX_HOLD_LOCK="+path)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdin.Close(); _ = cmd.Wait() })
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper output = %q, %v", line, err)
	}
}

// TestStartOwnPIDHeld is AC-1 of TASK-050: an old session file with the PID of this process, which
// another process holds, does not stop Start.
func TestStartOwnPIDHeld(t *testing.T) {
	cacheDir := t.TempDir()
	path := fakeSession(t, cacheDir, strconv.Itoa(os.Getpid()), "old", false)
	holdInOther(t, path)
	end := start(t, cacheDir)
	defer end()
	if data, _ := os.ReadFile(path); string(data) != "new\n" {
		t.Fatalf("session file = %q, want new", data)
	}
}

// TestCleanDeadPIDHeld is AC-2 of TASK-050: a held session file with the PID of no process is of an
// ended session, so Clean removes the cache.
func TestCleanDeadPIDHeld(t *testing.T) {
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cacheDir := filepath.Join(t.TempDir(), "claude-dx")
	mkdirs(t, cacheDir, "plugin/v1")
	holdInOther(t, fakeSession(t, cacheDir, strconv.Itoa(cmd.Process.Pid), "v1", false))
	if failed, err := Clean(cacheDir); err != nil || len(failed) != 0 {
		t.Fatalf("Clean = %v, %v", failed, err)
	}
	if exists(cacheDir) {
		t.Fatal("the cache directory exists")
	}
}

// AC-5.
func TestLockStaysAfterExec(t *testing.T) {
	path := filepath.Join(t.TempDir(), "1.lock")
	cmd := exec.Command(os.Args[0], "-test.run=^TestExecHelper$")
	cmd.Env = append(os.Environ(), "DX_EXEC_LOCK="+path)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper output = %q, %v", line, err)
	}
	if _, err := filelock.TryLock(path); !errors.Is(err, filelock.ErrLocked) {
		t.Fatalf("TryLock after exec = %v, want ErrLocked", err)
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	f, err := filelock.TryLock(path)
	if err != nil {
		t.Fatalf("TryLock after the program ended = %v", err)
	}
	_ = f.Close()
}
