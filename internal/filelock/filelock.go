// Package filelock takes an exclusive OS lock on a file. The lock works between processes.
package filelock

import (
	"errors"
	"os"
)

// ErrLocked means that another process or file handle holds the lock.
var ErrLocked = errors.New("the file is locked")

// Lock waits for an exclusive lock on path and returns the function that releases it.
// It makes the file if it does not exist.
func Lock(path string) (unlock func() error, err error) {
	f, err := open(path, true)
	if err != nil {
		return nil, err
	}
	// Close releases the OS lock.
	return f.Close, nil
}

// TryLock takes an exclusive lock on path with no wait. Close of the file releases the lock.
// It returns ErrLocked if another process holds the lock. It makes the file if it does not exist.
func TryLock(path string) (*os.File, error) {
	return open(path, false)
}

func open(path string, wait bool) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := lock(f, wait); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}
