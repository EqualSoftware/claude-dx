// Package session holds the cache lock (ADR-006). Each claude-dx process that uses the cache holds
// a shared lock, so "clean" disables the global items while the cache is in use, and a session start can
// remove the old plugin directories when no other process uses the cache.
package session

import (
	"os"
	"path/filepath"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// lockName is the cache lock. No process removes it, because a process that waits on a removed
// file gets a lock that other processes do not see (ADR-006).
const lockName = "cache.lock"

// Use makes the cache directory, waits for a shared cache lock and returns the function that
// releases it.
func Use(cacheDir string) (release func() error, err error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	return filelock.RLock(filepath.Join(cacheDir, lockName))
}

// Start removes the plugin directories but own and calls alone if no other process uses the cache.
// Then it takes the shared cache lock and calls ensure to get the plugin directory, because "clean"
// can remove the plugin before the shared lock. end releases the lock: call it when claude ends.
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
