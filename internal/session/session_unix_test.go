//go:build !windows

package session

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
