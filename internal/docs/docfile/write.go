package docfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// mu orders the document reads and changes of one process. The OS lock orders them between processes.
var mu sync.RWMutex

// errNoLock means that a store has no docs lock path. Without it, two processes could change one
// document at the same time with no sign.
var errNoLock = errors.New("the docs lock path is empty")

// Locked runs fn with the exclusive docs lock of the project (docs-common#R-10, R-11), so two sessions do not lose
// an edit. lock is the path of the lock file. The docs lock also protects the number counter (ADR-006).
func Locked(lock string, fn func() error) error {
	mu.Lock()
	defer mu.Unlock()
	return withLock(lock, filelock.Lock, fn)
}

// RLocked runs fn with a shared docs lock, so a read never sees half of a change of more than one
// file. Reads do not wait for other reads (ADR-006).
func RLocked(lock string, fn func() error) error {
	mu.RLock()
	defer mu.RUnlock()
	return withLock(lock, filelock.RLock, fn)
}

// withLock takes the OS lock of lock with take, runs fn and releases the lock.
func withLock(lock string, take func(string) (func() error, error), fn func() error) error {
	if lock == "" {
		return errNoLock
	}
	if err := os.MkdirAll(filepath.Dir(lock), 0o755); err != nil {
		return err
	}
	unlock, err := take(lock)
	if err != nil {
		return err
	}
	defer func() { _ = unlock() }()
	return fn()
}

// modeOf returns the permission bits of the file at path, or 0644 if it does not exist (docs-common#R-8).
func modeOf(path string) (fs.FileMode, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0o644, nil
	}
	if err != nil {
		return 0, err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return 0, fmt.Errorf("%s is a symbolic link. Replace it with a regular file", path)
	}
	return info.Mode().Perm(), nil
}

// Write writes data to a temporary file in the directory of path, then renames it to path, so a reader
// sees the old or the new content, never a partial file. The file keeps its mode. A failure removes the
// temporary file.
func Write(path string, data []byte) error {
	mode, err := modeOf(path)
	if err != nil {
		return err
	}
	cleanStale(filepath.Dir(path))
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	err = errors.Join(err, f.Chmod(mode), f.Sync(), f.Close())
	if err == nil {
		err = rename(tmp, path)
	}
	if err != nil {
		_ = os.Remove(tmp)
	}
	return err
}

// staleAge is the age after which a temporary file of a crashed write is removed.
const staleAge = time.Minute

// cleanStale removes the temporary files ".<name>.tmp-*" that are older than staleAge in dir. A crash
// between create and rename leaves them. The caller holds the docs lock, so no live write owns an old file.
func cleanStale(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, ".") || !strings.Contains(name, ".tmp-") {
			continue
		}
		if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > staleAge {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

// Move renames old to path with one rename, so a failure leaves one file (docs-common#R-9). It refuses a path
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

// MoveAside renames the bad file at path to path.bad, or to path.bad.<n> with the first free n if that
// name exists, so an old bad file stays.
func MoveAside(path string) error {
	name := path + ".bad"
	for n := 1; ; n++ {
		if _, err := os.Lstat(name); NotExist(err) {
			break
		}
		name = fmt.Sprintf("%s.bad.%d", path, n)
	}
	return Move(path, name)
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
