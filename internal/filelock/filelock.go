// Package filelock takes an exclusive OS lock on a file. The lock works between processes.
package filelock

import "os"

// Lock waits for an exclusive lock on path and returns the function that releases it.
// It makes the file if it does not exist.
func Lock(path string) (unlock func() error, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := lock(f); err != nil {
		_ = f.Close()
		return nil, err
	}
	// Close releases the OS lock.
	return f.Close, nil
}
