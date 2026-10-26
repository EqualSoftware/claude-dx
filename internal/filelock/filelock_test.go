package filelock

import (
	"errors"
	"os"
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
