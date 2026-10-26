// Package session holds the cache lock (ADR-015). Each claude-dx process that uses the cache holds
// a shared lock, so "cache clean" can refuse to remove a cache in use, and a session start can
// remove the old plugin directories when no other process uses the cache.
package session

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// lockName is the cache lock. No process removes it, because a process that waits on a removed
// file gets a lock that other processes do not see (ADR-015).
const lockName = "cache.lock"

// errInUse means that another claude-dx process holds the shared cache lock.
var errInUse = errors.New("a claude-dx session uses the cache. Close it and try again")

// Use makes the cache directory, waits for a shared cache lock and returns the function that
// releases it.
func Use(cacheDir string) (release func() error, err error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	return filelock.RLock(filepath.Join(cacheDir, lockName))
}

// Start removes the plugin directories but own and calls alone if no other process uses the cache.
// Then it takes the shared cache lock and calls ensure to get the plugin directory, because "cache
// clean" can remove the plugin before the shared lock. end releases the lock: call it when claude ends.
func Start(cacheDir, own string, alone func(), ensure func() (string, error)) (pluginDir string, end func(), err error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", nil, err
	}
	if f, err := filelock.TryLock(filepath.Join(cacheDir, lockName)); err == nil {
		removeUnused(filepath.Join(cacheDir, "plugin"), own)
		alone()
		_ = f.Close()
	}
	release, err := Use(cacheDir)
	if err != nil {
		return "", nil, err
	}
	pluginDir, err = ensure()
	if err != nil {
		_ = release()
		return "", nil, err
	}
	return pluginDir, func() { _ = release() }, nil
}

// Clean removes the cache directory if no process uses it, but not the lock files (ADR-015). It
// returns one line for each path that it cannot remove.
func Clean(cacheDir string) ([]string, error) {
	f, err := filelock.TryLock(filepath.Join(cacheDir, lockName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if errors.Is(err, filelock.ErrLocked) {
		return nil, errInUse
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return cache.Clean(cacheDir, lockFiles(cacheDir))
}

// lockFiles returns the cache lock and the docs lock of each project.
func lockFiles(cacheDir string) []string {
	keep := []string{filepath.Join(cacheDir, lockName)}
	projects := filepath.Join(cacheDir, "projects")
	entries, _ := os.ReadDir(projects)
	for _, e := range entries {
		path := filepath.Join(projects, e.Name(), "docs.lock")
		if _, err := os.Stat(path); err == nil {
			keep = append(keep, path)
		}
	}
	return keep
}

// removeUnused removes each entry of root but own, also the *.tmp-* directories of extractions
// that stopped. It ignores errors: the next start tries again.
func removeUnused(root, own string) {
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if e.Name() != own {
			_ = os.RemoveAll(filepath.Join(root, e.Name()))
		}
	}
}
