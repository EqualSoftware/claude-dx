// Package session holds a lock file for each running claude-dx session, so the launcher can remove
// the plugin directories that no session uses, and "cache clean" can refuse to remove a cache in use.
package session

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// lockName is the lock of the cache for session starts and "cache clean".
const lockName = "sessions.lock"

// Start holds the lock of the cache, calls ensure to get the plugin directory, writes and locks
// the session file <cacheDir>/sessions/<pid>.lock, and removes the plugin directories that no
// running session uses. end releases the session lock and removes the session file. On Unix,
// claude keeps the lock after exec, and the OS releases it when claude ends.
func Start(cacheDir string, ensure func() (string, error)) (pluginDir string, end func(), err error) {
	if err := os.MkdirAll(filepath.Join(cacheDir, "sessions"), 0o755); err != nil {
		return "", nil, err
	}
	unlock, err := filelock.Lock(filepath.Join(cacheDir, lockName))
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = unlock() }()
	pluginDir, err = ensure()
	if err != nil {
		return "", nil, err
	}
	f, err := hold(filepath.Join(cacheDir, "sessions", strconv.Itoa(os.Getpid())+".lock"), filepath.Base(pluginDir))
	if err != nil {
		return "", nil, err
	}
	removeUnused(filepath.Join(cacheDir, "plugin"), running(cacheDir))
	return pluginDir, func() { _ = f.Close(); _ = os.Remove(f.Name()) }, nil
}

// Clean removes the cache directory if no session runs. It returns one line for each path that it
// cannot remove. The lock file is removed last, after the lock is released, because Windows cannot
// remove an open file.
func Clean(cacheDir string) ([]string, error) {
	if _, err := os.Stat(cacheDir); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	lock := filepath.Join(cacheDir, lockName)
	unlock, err := filelock.Lock(lock)
	if err != nil {
		return nil, err
	}
	if pids := slices.Sorted(maps.Keys(running(cacheDir))); len(pids) > 0 {
		_ = unlock()
		return nil, fmt.Errorf("a claude-dx session uses the cache (PID %s). Close it and try again", pids[0])
	}
	failed, err := cache.Clean(cacheDir, lock)
	_ = unlock()
	if err != nil {
		return nil, err
	}
	for _, path := range []string{lock, cacheDir} {
		if err := os.Remove(path); err != nil && len(failed) == 0 {
			failed = append(failed, path+": "+err.Error())
		}
	}
	return failed, nil
}

// hold writes the plugin directory name to the session file and locks it. The descriptor stays
// open after exec, so the lock stays while the session runs. It first removes an old file of the
// same PID: two running processes do not have one PID, so that file is of an ended session, and a
// child of that session can still hold its lock (TASK-050). On Unix the new file has a free lock.
func hold(path, dirName string) (*os.File, error) {
	_ = os.Remove(path)
	if err := os.WriteFile(path, []byte(dirName+"\n"), 0o644); err != nil {
		return nil, err
	}
	f, err := filelock.TryLock(path)
	if err != nil {
		return nil, err
	}
	if err := inherit(f); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}

// running returns the plugin directory name of each running session, by PID. A session file with
// a free lock, or with the PID of no process, is of a session that ended, so running removes it. A
// file that it cannot lock for another reason counts as a running session, because a wrong removal
// is worse than a stale file.
func running(cacheDir string) map[string]string {
	dir := filepath.Join(cacheDir, "sessions")
	entries, _ := os.ReadDir(dir)
	sessions := map[string]string{}
	for _, e := range entries {
		pid, ok := strings.CutSuffix(e.Name(), ".lock")
		if !ok {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := filelock.TryLock(path)
		if err == nil {
			_ = f.Close()
			_ = os.Remove(path)
			continue
		}
		if !alive(pid) {
			_ = os.Remove(path)
			continue
		}
		data, _ := os.ReadFile(path)
		sessions[pid] = strings.TrimSpace(string(data))
	}
	return sessions
}

// removeUnused removes each entry of root that no running session uses, also the *.tmp-*
// directories of extractions that stopped. It ignores errors: the next start tries again.
func removeUnused(root string, sessions map[string]string) {
	entries, _ := os.ReadDir(root)
	used := slices.Collect(maps.Values(sessions))
	for _, e := range entries {
		if !slices.Contains(used, e.Name()) {
			_ = os.RemoveAll(filepath.Join(root, e.Name()))
		}
	}
}
