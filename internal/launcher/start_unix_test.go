//go:build !windows

package launcher

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// readyFile sets DX_READY to a new path and returns it.
func readyFile(t *testing.T) string {
	t.Helper()
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("DX_READY", ready)
	return ready
}

// waitReady waits until the fake claude makes the file ready. It can run in another goroutine.
func waitReady(t *testing.T, ready string) bool {
	t.Helper()
	for range 500 {
		if _, err := os.Stat(ready); err == nil {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("the fake claude did not start")
	return false
}

// signalSelf sends sig to this process when the fake claude is ready, while start waits.
func signalSelf(t *testing.T, ready string, sigs ...syscall.Signal) {
	t.Helper()
	go func() {
		if !waitReady(t, ready) {
			return
		}
		for _, sig := range sigs {
			_ = syscall.Kill(syscall.Getpid(), sig)
		}
	}()
}

// TestStartForwardsSIGTERM is AC-2 of TASK-060.
func TestStartForwardsSIGTERM(t *testing.T) {
	signalSelf(t, readyFile(t), syscall.SIGTERM)
	if code, err := fakeClaude(t, "wait"); err != nil || code != 7 {
		t.Fatalf("start = %d, %v; want 7", code, err)
	}
}

// TestStartIgnoresSIGINT is AC-3 of TASK-060: after SIGINT, this process and the fake claude still run,
// so the SIGTERM after it ends the fake claude with its own exit code. The fake claude also gets SIGINT
// with the default action (FR-2).
func TestStartIgnoresSIGINT(t *testing.T) {
	signalSelf(t, readyFile(t), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	if code, err := fakeClaude(t, "wait"); err != nil || code != 7 {
		t.Fatalf("start = %d, %v; want 7", code, err)
	}
}

// TestStartSignalExitCode is FR-4 of TASK-060: a claude that SIGKILL ended gives 128 + 9.
func TestStartSignalExitCode(t *testing.T) {
	if code, err := fakeClaude(t, "kill"); err != nil || code != 137 {
		t.Fatalf("start = %d, %v; want 137", code, err)
	}
}
