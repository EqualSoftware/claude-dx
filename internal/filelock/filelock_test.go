package filelock

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestLockExcludes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	unlock, err := Lock(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	got := make(chan time.Time, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		second, err := Lock(path)
		if err != nil {
			t.Error(err)
			return
		}
		got <- time.Now()
		_ = second()
	}()
	time.Sleep(100 * time.Millisecond)
	released := time.Now()
	if err := unlock(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if (<-got).Before(released) {
		t.Fatal("second lock was taken before the first lock was released")
	}
}

func TestTryLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := TryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TryLock(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("TryLock of a held lock = %v, want ErrLocked", err)
	}
	// Other handles can read the file while the lock is held.
	if data, err := os.ReadFile(path); err != nil || string(data) != "data" {
		t.Fatalf("ReadFile with the lock = %q, %v", data, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := TryLock(path)
	if err != nil {
		t.Fatalf("TryLock after Close = %v", err)
	}
	_ = second.Close()
}

// TestHoldHelper is a helper process: it holds a lock of the file in DX_LOCK_PATH until stdin
// closes. The lock is exclusive if DX_LOCK_MODE is "exclusive", else shared.
func TestHoldHelper(t *testing.T) {
	path := os.Getenv("DX_LOCK_PATH")
	if path == "" {
		t.Skip("helper process")
	}
	try := TryRLock
	if os.Getenv("DX_LOCK_MODE") == "exclusive" {
		try = TryLock
	}
	f, err := try(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	fmt.Println("ready")
	_, _ = io.ReadAll(os.Stdin)
}

// holdInOther locks path in a helper process with mode. release ends the helper process, and
// with it the lock.
func holdInOther(t *testing.T, path, mode string) (release func()) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHoldHelper$")
	cmd.Env = append(os.Environ(), "DX_LOCK_PATH="+path, "DX_LOCK_MODE="+mode)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	release = func() { _ = stdin.Close(); _ = cmd.Wait() }
	t.Cleanup(release)
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper output = %q, %v", line, err)
	}
	return release
}

// TestSharedLock is A-1 of TASK-061: a shared lock of another process allows a shared lock and
// excludes an exclusive lock.
func TestSharedLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	release := holdInOther(t, path, "shared")
	f, err := TryRLock(path)
	if err != nil {
		t.Fatalf("TryRLock with a shared lock = %v", err)
	}
	_ = f.Close()
	if _, err := TryLock(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("TryLock with a shared lock = %v, want ErrLocked", err)
	}
	release()
	f, err = TryLock(path)
	if err != nil {
		t.Fatalf("TryLock after the release = %v", err)
	}
	_ = f.Close()
}

// TestExclusiveLockExcludesShared is A-1 of TASK-061: an exclusive lock of another process
// excludes a shared lock.
func TestExclusiveLockExcludesShared(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	release := holdInOther(t, path, "exclusive")
	if _, err := TryRLock(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("TryRLock with an exclusive lock = %v, want ErrLocked", err)
	}
	release()
	unlock, err := RLock(path)
	if err != nil {
		t.Fatalf("RLock after the release = %v", err)
	}
	_ = unlock()
}

func TestLockWaitsForShared(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	unlock, err := RLock(path)
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan time.Time, 1)
	go func() {
		second, err := Lock(path)
		if err != nil {
			t.Error(err)
			got <- time.Time{}
			return
		}
		got <- time.Now()
		_ = second()
	}()
	time.Sleep(100 * time.Millisecond)
	released := time.Now()
	if err := unlock(); err != nil {
		t.Fatal(err)
	}
	if (<-got).Before(released) {
		t.Fatal("the exclusive lock was taken before the shared lock was released")
	}
}
