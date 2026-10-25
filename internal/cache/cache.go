// Package cache gives the paths of the claude-dx cache (ADR-001).
package cache

import (
	"os"
	"path/filepath"
)

// Dir returns the cache directory: os.UserCacheDir()/claude-dx.
func Dir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "claude-dx"), nil
}
