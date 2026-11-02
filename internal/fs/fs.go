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
	"syscall"

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
	if filepath.VolumeName(abs) != "" && !filepath.IsAbs(abs) {
		return "", fmt.Errorf("path %s has a drive and no root. Use a path from the project root", rel)
	}
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
// (filesystem-tools#R-6, R-10). It compares file identity, so a name with another case on a
// case-insensitive volume is also found. The string compare is the fallback.
func (t Tools) protected(path string) bool {
	root := t.Project.Root
	if path == root || within(filepath.Join(root, ".git"), path) {
		return true
	}
	if info, err := os.Lstat(path); err == nil {
		if rootInfo, err := os.Lstat(root); err == nil && os.SameFile(info, rootInfo) {
			return true
		}
	}
	gitInfo, err := os.Lstat(filepath.Join(root, ".git"))
	return err == nil && t.hasAncestor(gitInfo, path, true)
}

// hasAncestor reports whether target is path (when self is true) or a directory above path, up to the project root.
func (t Tools) hasAncestor(target os.FileInfo, path string, self bool) bool {
	for p := path; ; p = filepath.Dir(p) {
		if info, err := os.Lstat(p); err == nil && (self || p != path) && os.SameFile(info, target) {
			return true
		}
		if p == t.Project.Root || filepath.Dir(p) == p {
			return false
		}
	}
}

// within reports whether path is dir or a path under dir.
func within(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// osRename is os.Rename. A test replaces it to make a rename fail.
var osRename = os.Rename

// plan is a checked move or copy. same is true for a case-only rename of one file.
type plan struct {
	src, dst string
	same     bool
}

// pair resolves and checks the source and the destination of fs_move and fs_copy.
func (t Tools) pair(source, destination string, overwrite, move bool) (plan, error) {
	src, err := t.path(source)
	if err != nil {
		return plan{}, err
	}
	dst, err := t.path(destination)
	if err != nil {
		return plan{}, err
	}
	if t.protected(src) || t.protected(dst) {
		return plan{}, errors.New("the project root and the .git directory are not allowed as source or destination")
	}
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return plan{}, fmt.Errorf("%s does not exist", source)
	}
	if t.hasAncestor(srcInfo, dst, false) {
		return plan{}, errors.New("the source and the destination contain each other")
	}
	dstInfo, err := os.Lstat(dst)
	if err != nil {
		return plan{src: src, dst: dst}, nil
	}
	return t.existing(plan{src: src, dst: dst}, srcInfo, dstInfo, destination, overwrite, move)
}

// existing checks a destination that exists.
func (t Tools) existing(pl plan, srcInfo, dstInfo os.FileInfo, destination string, overwrite, move bool) (plan, error) {
	if os.SameFile(srcInfo, dstInfo) {
		if !move || pl.src == pl.dst || !strings.EqualFold(pl.src, pl.dst) {
			return plan{}, errors.New("the source and the destination are the same file")
		}
		pl.same = true
		return pl, nil
	}
	if t.hasAncestor(dstInfo, pl.src, false) {
		return plan{}, errors.New("the source and the destination contain each other")
	}
	if !overwrite {
		return plan{}, fmt.Errorf("%s exists. Set overwrite to true to replace it", destination)
	}
	return pl, nil
}

// Move moves a file or a directory (filesystem-tools#R-7).
func (t Tools) Move(source, destination string, overwrite bool) (string, error) {
	pl, err := t.pair(source, destination, overwrite, true)
	if err != nil {
		return "", err
	}
	if pl.same {
		err = osRename(pl.src, pl.dst)
	} else {
		err = replace(pl.src, pl.dst)
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Moved %s to %s.", source, destination), nil
}

// Copy copies a file, or a directory with all its contents (filesystem-tools#R-7). It copies to a
// temporary name next to the destination, so a failure keeps the destination.
func (t Tools) Copy(source, destination string, overwrite bool) (string, error) {
	pl, err := t.pair(source, destination, overwrite, false)
	if err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(pl.dst), ".fs-copy-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	stage := filepath.Join(tmp, "new")
	if err := copyAny(pl.src, stage); err != nil {
		return "", err
	}
	if err := replace(stage, pl.dst); err != nil {
		return "", err
	}
	return fmt.Sprintf("Copied %s to %s.", source, destination), nil
}

// replace renames src to dst. A file over a file is one atomic rename. In the other cases that
// replace something, dst moves to a backup name first, and comes back if the rename fails.
func replace(src, dst string) error {
	dstInfo, err := os.Lstat(dst)
	if err != nil {
		return rename(src, dst)
	}
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !dstInfo.IsDir() && !srcInfo.IsDir() {
		return rename(src, dst)
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dst), ".fs-old-")
	if err != nil {
		return err
	}
	keep := false // true when the backup is the only copy of the old destination
	defer func() {
		if !keep {
			_ = os.RemoveAll(tmp)
		}
	}()
	backup := filepath.Join(tmp, "old")
	if err := osRename(dst, backup); err != nil {
		return err
	}
	err = rename(src, dst)
	if err == nil {
		return nil
	}
	if rerr := osRename(backup, dst); rerr != nil {
		keep = true
		return fmt.Errorf("%w. The old destination is saved in %s", err, backup)
	}
	return err
}

// rename is os.Rename with a clear error for two devices.
func rename(src, dst string) error {
	err := osRename(src, dst)
	if errors.Is(err, syscall.EXDEV) {
		return errors.New("the source and the destination are on different devices. Use fs_copy, then fs_delete")
	}
	return err
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
	case info.Mode().IsRegular():
		return copyFile(src, dst, info.Mode().Perm())
	default:
		return fmt.Errorf("%s is not a file, a directory or a symbolic link", filepath.Base(src))
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
		if t.protected(a) {
			return "", errors.New("the project root and the .git directory are not allowed")
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
