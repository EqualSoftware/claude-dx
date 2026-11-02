// Package fs has the fs_* tools: move, copy, delete, mkdir and tree for the files of the project
// (ADR-002, ADR-016, filesystem-tools).
package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// Tools runs the fs_* tools in one project.
type Tools struct {
	Project project.Project
}

// path returns the absolute path of rel in the project. It acts on a symbolic link, not on its
// target: it resolves the parent directory and keeps the last name (filesystem-tools#R-9).
// A path outside the project root is refused (filesystem-tools#R-1).
func (t Tools) path(rel string) (string, error) {
	abs := rel
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(t.Project.Root, abs)
	}
	abs = filepath.Clean(abs)
	if abs == t.Project.Root {
		return abs, nil
	}
	parent, err := t.Project.Resolve(filepath.Dir(abs))
	if err != nil {
		return "", fmt.Errorf("path %s is outside the project", rel)
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}

// protected reports whether path is the project root, the .git directory or a path inside it
// (filesystem-tools#R-6, R-10).
func (t Tools) protected(path string) bool {
	if path == t.Project.Root {
		return true
	}
	return within(filepath.Join(t.Project.Root, ".git"), path)
}

// within reports whether path is dir or a path under dir.
func within(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// pair resolves and checks the source and the destination of fs_move and fs_copy.
func (t Tools) pair(source, destination string, overwrite bool) (src, dst string, err error) {
	if src, err = t.path(source); err != nil {
		return "", "", err
	}
	if dst, err = t.path(destination); err != nil {
		return "", "", err
	}
	if t.protected(src) || t.protected(dst) {
		return "", "", errors.New("the project root and the .git directory are not allowed as source or destination")
	}
	if _, err = os.Lstat(src); err != nil {
		return "", "", fmt.Errorf("%s does not exist", source)
	}
	if within(src, dst) || within(dst, src) {
		return "", "", errors.New("the source and the destination contain each other")
	}
	if _, err = os.Lstat(dst); err != nil {
		return src, dst, nil
	}
	if !overwrite {
		return "", "", fmt.Errorf("%s exists. Set overwrite to true to replace it", destination)
	}
	return src, dst, nil
}

// Move moves a file or a directory (filesystem-tools#R-7).
func (t Tools) Move(source, destination string, overwrite bool) (string, error) {
	src, dst, err := t.pair(source, destination, overwrite)
	if err != nil {
		return "", err
	}
	if err := removeIfExists(dst); err != nil {
		return "", err
	}
	if err := os.Rename(src, dst); err != nil {
		return "", err
	}
	return fmt.Sprintf("Moved %s to %s.", source, destination), nil
}

// Copy copies a file, or a directory with all its contents (filesystem-tools#R-7).
func (t Tools) Copy(source, destination string, overwrite bool) (string, error) {
	src, dst, err := t.pair(source, destination, overwrite)
	if err != nil {
		return "", err
	}
	if err := removeIfExists(dst); err != nil {
		return "", err
	}
	if err := copyAny(src, dst); err != nil {
		return "", err
	}
	return fmt.Sprintf("Copied %s to %s.", source, destination), nil
}

func removeIfExists(path string) error {
	if _, err := os.Lstat(path); err != nil {
		return nil
	}
	return os.RemoveAll(path)
}

func copyAny(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	case info.IsDir():
		return copyDir(src, dst, info.Mode().Perm())
	default:
		return copyFile(src, dst, info.Mode().Perm())
	}
}

func copyDir(src, dst string, perm os.FileMode) error {
	if err := os.Mkdir(dst, perm|0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyAny(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return os.Chmod(dst, perm)
}

func copyFile(src, dst string, perm os.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(out, in)
	return err
}

// Delete deletes files and directories (filesystem-tools#R-6, R-8). It checks all paths first.
func (t Tools) Delete(paths []string, recursive bool) (string, error) {
	if len(paths) == 0 {
		return "", errors.New("paths is empty")
	}
	abs := make([]string, len(paths))
	for i, p := range paths {
		a, err := t.deletable(p, recursive)
		if err != nil {
			return "", err
		}
		abs[i] = a
	}
	for _, a := range abs {
		if err := os.RemoveAll(a); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("Deleted %d %s.", len(abs), noun(len(abs))), nil
}

func (t Tools) deletable(rel string, recursive bool) (string, error) {
	abs, err := t.path(rel)
	if err != nil {
		return "", err
	}
	if t.protected(abs) {
		return "", errors.New("the project root and the .git directory cannot be deleted")
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("%s does not exist", rel)
	}
	if info.IsDir() && !recursive && !emptyDir(abs) {
		return "", fmt.Errorf("%s is a directory that is not empty. Set recursive to true to delete it", rel)
	}
	return abs, nil
}

func emptyDir(path string) bool {
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) == 0
}

// Mkdir makes directories and their parents. A directory that exists is not an error (filesystem-tools#R-8).
func (t Tools) Mkdir(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", errors.New("paths is empty")
	}
	abs := make([]string, len(paths))
	for i, p := range paths {
		a, err := t.path(p)
		if err != nil {
			return "", err
		}
		abs[i] = a
	}
	for _, a := range abs {
		if err := os.MkdirAll(a, 0o755); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("Made %d %s.", len(abs), noun(len(abs))), nil
}

func noun(n int) string {
	if n == 1 {
		return "path"
	}
	return "paths"
}
