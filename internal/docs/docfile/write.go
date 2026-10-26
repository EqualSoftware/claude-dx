package docfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// mu orders the document changes of one process. The OS lock of Locked orders the changes of processes.
var mu sync.Mutex

// Locked runs fn with the docs lock of the project (TASK-047), so two sessions do not lose an edit.
// lock is the path of the lock file; an empty lock takes only the lock of this process. The docs lock
// also protects the number counter (ADR-015).
func Locked(lock string, fn func() error) error {
	mu.Lock()
	defer mu.Unlock()
	if lock == "" {
		return fn()
	}
	if err := os.MkdirAll(filepath.Dir(lock), 0o755); err != nil {
		return err
	}
	unlock, err := filelock.Lock(lock)
	if err != nil {
		return err
	}
	defer func() { _ = unlock() }()
	return fn()
}

// Write writes data to a temporary file in the directory of path, then renames it to path, so a reader
// sees the old or the new content, never a partial file. A failure removes the temporary file.
func Write(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	err = errors.Join(err, f.Chmod(0o644), f.Close())
	if err == nil {
		err = rename(tmp, path)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}

// Move renames old to path with one rename, so a failure leaves one file (TASK-049). It refuses a path
// that exists, because os.Rename replaces a file on Unix.
func Move(old, path string) error {
	_, err := os.Lstat(path)
	if err == nil {
		return &fs.PathError{Op: "move", Path: path, Err: fs.ErrExist}
	}
	if !NotExist(err) {
		return err
	}
	return rename(old, path)
}

// rename renames tmp to path. Windows refuses to replace a file that a reader has open, because Go opens
// files with no FILE_SHARE_DELETE, so rename tries again for a short time.
func rename(tmp, path string) error {
	err := os.Rename(tmp, path)
	for i := 0; err != nil && runtime.GOOS == "windows" && i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		err = os.Rename(tmp, path)
	}
	return err
}
