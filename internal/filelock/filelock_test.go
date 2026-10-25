package filelock

import (
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
