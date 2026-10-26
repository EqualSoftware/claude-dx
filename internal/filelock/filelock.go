// Package filelock takes an OS lock on a file. The lock works between processes. Many processes
// can hold a shared lock at the same time. An exclusive lock excludes all other locks.
package filelock

import (
	"errors"
	"os"
)

// ErrLocked means that another process or file handle holds a lock that excludes this lock.
var ErrLocked = errors.New("the file is locked")

// Lock waits for an exclusive lock on path and returns the function that releases it.
// It makes the file if it does not exist.
func Lock(path string) (unlock func() error, err error) {
	return wait(path, true)
}

// RLock waits for a shared lock on path and returns the function that releases it.
// It makes the file if it does not exist.
func RLock(path string) (unlock func() error, err error) {
	return wait(path, false)
}

// TryLock takes an exclusive lock on path with no wait. Close of the file releases the lock.
// It returns ErrLocked if another process holds a lock. It makes the file if it does not exist.
func TryLock(path string) (*os.File, error) {
	return open(path, true, false)
}

// TryRLock takes a shared lock on path with no wait. Close of the file releases the lock.
// It returns ErrLocked if another process holds the exclusive lock. It makes the file if it does
// not exist.
func TryRLock(path string) (*os.File, error) {
	return open(path, false, false)
}

func wait(path string, exclusive bool) (func() error, error) {
	f, err := open(path, exclusive, true)
	if err != nil {
		return nil, err
	}
	// Close releases the OS lock.
	return f.Close, nil
}

func open(path string, exclusive, wait bool) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := lock(f, exclusive, wait); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}
