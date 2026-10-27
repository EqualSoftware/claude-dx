package project

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Names of entries in a project data directory (ADR-006, ADR-007).
const (
	DocsLock = "docs.lock" // the docs lock
	Tasks    = "tasks"     // the folder of the local tasks: tasks/<spec>/<task-slug>.md
)

// dataNames are the entries of a project data directory that hold data. docs.lock is not data,
// because it has no content and no process removes it (ADR-006).
var dataNames = []string{"project.json", Tasks}

// DataDir returns the project data directory (ADR-007): <Common>/dx in git, else
// <cacheDir>/projects/<Key>. In git, the cache folder of the top level stays the data directory while it
// has data and <Common>/dx does not exist, so all processes use one docs lock until MoveData runs.
func (p Project) DataDir(cacheDir string) string {
	if p.Common == "" {
		return filepath.Join(cacheDir, "projects", p.Key)
	}
	if old, _, ok := p.pending(cacheDir); ok {
		return old
	}
	return filepath.Join(p.Common, "dx")
}

// MoveData moves the data of the cache folder of the top level to <Common>/dx if a move is pending
// (ADR-007). Call it only while no other claude-dx process runs. It copies the data to <Common>/dx.tmp-*
// and renames that folder to dx, because the cache and the repository can be on different file systems.
// Then it removes the old data but docs.lock. It also removes the dx.tmp-* folders of moves that stopped.
func (p Project) MoveData(cacheDir string) error {
	if p.Common == "" {
		return nil
	}
	removeTemps(p.Common)
	old, dx, ok := p.pending(cacheDir)
	if !ok {
		return nil
	}
	tmp, err := os.MkdirTemp(p.Common, "dx.tmp-")
	if err != nil {
		return err
	}
	if err := copyData(old, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if err := os.Rename(tmp, dx); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	// The data is in dx now. Old data that stays is garbage that "claude-dx clean" shows.
	for _, name := range dataNames {
		_ = os.RemoveAll(filepath.Join(old, name))
	}
	return nil
}

// pending returns the cache folder of the top level and <Common>/dx, and reports whether the data must
// move from the first to the second.
func (p Project) pending(cacheDir string) (old, dx string, ok bool) {
	old = filepath.Join(cacheDir, "projects", key(p.Root))
	dx = filepath.Join(p.Common, "dx")
	if _, err := os.Stat(dx); err == nil {
		return old, dx, false
	}
	for _, name := range dataNames {
		if _, err := os.Stat(filepath.Join(old, name)); err == nil {
			return old, dx, true
		}
	}
	return old, dx, false
}

// removeTemps removes the dx.tmp-* folders in common.
func removeTemps(common string) {
	entries, _ := os.ReadDir(common)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "dx.tmp-") {
			_ = os.RemoveAll(filepath.Join(common, e.Name()))
		}
	}
}

// copyData copies the data entries of src to dst. dst exists and gets the mode of a new directory.
func copyData(src, dst string) error {
	if err := os.Chmod(dst, 0o755); err != nil {
		return err
	}
	for _, name := range dataNames {
		if err := copyTree(filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
			return err
		}
	}
	return nil
}

// copyTree copies the regular files and directories of src to dst, with their permission bits. A src
// that does not exist copies nothing. Another file type is an error.
func copyTree(src, dst string) error {
	if _, err := os.Lstat(src); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, info.Mode().Perm())
		case d.Type().IsRegular():
			return copyFile(path, target, info.Mode().Perm())
		}
		return fmt.Errorf("%s is not a regular file or a directory", path)
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err := errors.Join(err, out.Close()); err != nil {
		return err
	}
	// OpenFile uses the umask, so Chmod sets the exact bits.
	return os.Chmod(dst, mode)
}
