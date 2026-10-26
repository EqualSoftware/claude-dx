// Package cache gives the paths of the claude-dx cache (ADR-001).
package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Dir returns the cache directory: os.UserCacheDir()/claude-dx.
func Dir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "claude-dx"), nil
}

// Clean removes the cache directory dir, but not the file keep and its parent directories. It
// continues after an error (example: a file in use on Windows) and returns one line for each path
// that it cannot remove.
func Clean(dir, keep string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var failed []string
	// WalkDir gives a directory before its files, so the reverse order removes the files first.
	for i := len(paths) - 1; i >= 0; i-- {
		if paths[i] == keep || strings.HasPrefix(keep, paths[i]+string(filepath.Separator)) {
			continue
		}
		err := os.Remove(paths[i])
		if err == nil || hasFailedChild(failed, paths[i]) {
			continue
		}
		failed = append(failed, paths[i]+": "+err.Error())
	}
	return failed, nil
}

// hasFailedChild reports whether a path in failed is in dir. Then dir is not empty, and the
// report of the child is sufficient.
func hasFailedChild(failed []string, dir string) bool {
	for _, f := range failed {
		if strings.HasPrefix(f, dir+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
