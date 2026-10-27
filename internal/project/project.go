// Package project finds the project root and key, and keeps paths inside the root (ADR-002, ADR-007).
package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Project is the project of the current session.
type Project struct {
	// Root is the git top level, or the start directory outside git. Symlinks are resolved.
	Root string
	// Key identifies the repository. All worktrees of one repository have the same key.
	Key string
	// Common is the git common directory, or empty outside git. All worktrees of one repository share it.
	Common string
}

// Open finds the project that contains dir.
func Open(dir string) (Project, error) {
	dir, err := realPath(dir)
	if err != nil {
		return Project{}, err
	}
	root := dir
	if top, ok := git(dir, "--show-toplevel"); ok {
		root = top
	}
	common, ok := git(dir, "--git-common-dir")
	if !ok {
		return Project{Root: root, Key: key(root)}, nil
	}
	return Project{Root: root, Key: key(common), Common: common}, nil
}

// Resolve returns the absolute path of rel, or an error if the path is outside the root.
// It resolves ".." and the symlinks of the parts of the path that exist.
func (p Project) Resolve(rel string) (string, error) {
	path := rel
	if !filepath.IsAbs(path) {
		path = filepath.Join(p.Root, path)
	}
	resolved, err := realPath(path)
	if err != nil {
		return "", err
	}
	if !inside(p.Root, resolved) {
		return "", fmt.Errorf("path %s is outside the project", rel)
	}
	return resolved, nil
}

// git runs "git rev-parse <flag>" in dir and returns an absolute, resolved path.
func git(dir, flag string) (string, bool) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", flag).Output()
	if err != nil {
		return "", false
	}
	path := filepath.FromSlash(strings.TrimSpace(string(out)))
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	resolved, err := realPath(path)
	return resolved, err == nil
}

// key returns the first 16 hex characters of the SHA-256 of path.
func key(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:8])
}

// realPath resolves the symlinks of the longest part of path that exists.
// The parts that do not exist yet stay as they are.
func realPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for existing := path; ; existing = filepath.Dir(existing) {
		if resolved, err := filepath.EvalSymlinks(existing); err == nil {
			rest, _ := filepath.Rel(existing, path)
			return filepath.Join(resolved, rest), nil
		}
		if filepath.Dir(existing) == existing {
			return path, nil
		}
	}
}

// inside reports whether path is root or a path under root.
func inside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
