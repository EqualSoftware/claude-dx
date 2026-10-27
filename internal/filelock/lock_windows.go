//go:build windows

package filelock

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// lock locks one byte at offset 1<<62. A Windows lock also stops reads of the locked bytes from
// other handles, so a lock after the end of the file lets them read the data.
func lock(f *os.File, exclusive, wait bool) error {
	var flags uint32
	if exclusive {
		flags |= windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	if !wait {
		flags |= windows.LOCKFILE_FAIL_IMMEDIATELY
	}
	at := &windows.Overlapped{OffsetHigh: 1 << 30}
	err := windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, at)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return ErrLocked
	}
	return err
}
