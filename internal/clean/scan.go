// Package clean finds the data of claude-dx that a person can delete, and deletes the items that the
// person selects (ADR-011). It never shows or deletes lock files (ADR-006).
package clean

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/docs/task"
	"github.com/EqualSoftware/claude-dx/internal/filelock"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// ReasonInUse is the reason of a disabled item: a session holds the shared cache lock.
const ReasonInUse = "(in use by a running session)"

const (
	cacheLock = "cache.lock"
	docsLock  = project.DocsLock
)

var errInUse = errors.New(ReasonInUse)

// Item is one thing that a person can delete.
type Item struct {
	Group  string // the name of the group of the item
	Label  string
	Size   int64  // bytes, with no lock files
	Reason string // not empty: the item is disabled, and this is the reason
	remove func() (result, error)
}

// result is what a delete of one item did.
type result struct {
	removed int   // the number of paths that it removed
	bytes   int64 // the size of these paths
	skipped int   // the number of tasks that it skipped, because they changed after the scan
}

// Text returns the group and the label of the item.
func (it Item) Text() string {
	return it.Group + ": " + it.Label
}

// Group is a list of items with a name: the global group or a project.
type Group struct {
	Name  string
	Items []Item
	// ListOnly means that List shows the group, but a person cannot select its items.
	ListOnly bool
}

// Scan returns the groups of the cache in cacheDir. own is the name of the plugin directory of this
// binary: it is not an old version. A group with no items is not in the result.
func Scan(cacheDir, own string) ([]Group, error) {
	if _, err := os.Stat(cacheDir); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	var groups []Group
	g, ok, err := globalGroup(cacheDir, own)
	if err != nil {
		return nil, err
	}
	if ok {
		groups = append(groups, g)
	}
	return append(groups, projectGroups(cacheDir)...), nil
}

// globalGroup returns the item of all old plugin versions: each entry of <cache>/plugin but own.
// The item is disabled while a process holds the shared cache lock.
func globalGroup(cacheDir, own string) (Group, bool, error) {
	entries, _ := os.ReadDir(filepath.Join(cacheDir, "plugin"))
	var paths []string
	for _, e := range entries {
		if e.Name() != own {
			paths = append(paths, filepath.Join(cacheDir, "plugin", e.Name()))
		}
	}
	if len(paths) == 0 {
		return Group{}, false, nil
	}
	lock := filepath.Join(cacheDir, cacheLock)
	it := Item{Group: "Global", Label: "Old plugin versions", Size: size(paths)}
	f, err := filelock.TryLock(lock)
	switch {
	case errors.Is(err, filelock.ErrLocked):
		it.Reason = ReasonInUse
	case err != nil:
		return Group{}, false, err
	default:
		_ = f.Close()
		it.remove = func() (result, error) { return removeOld(lock, paths) }
	}
	return Group{Name: "Global", Items: []Item{it}}, true, nil
}

// removeOld removes paths while it holds the exclusive cache lock, with no wait.
func removeOld(lock string, paths []string) (result, error) {
	f, err := filelock.TryLock(lock)
	if errors.Is(err, filelock.ErrLocked) {
		return result{}, errInUse
	}
	if err != nil {
		return result{}, err
	}
	defer func() { _ = f.Close() }()
	return removeAll(paths)
}

// info is the Project file (cache) of a project folder.
type info struct {
	key      string
	path     string
	lastUsed string
	missing  bool   // the project path does not exist
	dataDir  string // the data directory of a project that exists
}

// projectGroups returns the groups of the known projects, the newest first.
func projectGroups(cacheDir string) []Group {
	entries, _ := os.ReadDir(filepath.Join(cacheDir, "projects"))
	var infos []info
	owned := map[string]bool{} // the data directories of the projects that exist
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		in := readInfo(cacheDir, e.Name())
		if in.dataDir != "" {
			owned[in.dataDir] = true
		}
		infos = append(infos, in)
	}
	slices.SortStableFunc(infos, func(a, b info) int { return cmp.Compare(b.lastUsed, a.lastUsed) })
	var groups []Group
	for _, in := range infos {
		if g, ok := in.group(cacheDir, owned); ok {
			groups = append(groups, g)
		}
	}
	return groups
}

// readInfo reads <cache>/projects/<key>/project.json and finds the data directory of the project.
// A missing or bad file gives an empty path.
func readInfo(cacheDir, key string) info {
	in := info{key: key}
	data, err := os.ReadFile(filepath.Join(cacheDir, "projects", key, "project.json"))
	if err != nil {
		return in
	}
	var f struct {
		Path     string `json:"path"`
		LastUsed string `json:"last_used"`
	}
	if json.Unmarshal(data, &f) != nil || f.Path == "" {
		return in
	}
	in.path, in.lastUsed = f.Path, f.LastUsed
	if _, err := os.Stat(f.Path); errors.Is(err, fs.ErrNotExist) {
		in.missing = true
		return in
	}
	if p, err := project.Open(f.Path); err == nil {
		in.dataDir = filepath.Clean(p.DataDir(cacheDir))
	}
	return in
}

// group returns the group of the project. The project file (cache) and the counters of a project
// that exists are not items. A folder with no path is only in the list, unless it is the data
// directory of a project that exists (a move of the data to the repository is pending).
func (in info) group(cacheDir string, owned map[string]bool) (Group, bool) {
	dir := filepath.Join(cacheDir, "projects", in.key)
	switch {
	case in.path == "":
		if owned[filepath.Clean(dir)] {
			return Group{}, false
		}
		g, ok := folderGroup(cacheDir, in.key+" (unknown)", dir)
		g.ListOnly = true
		for i := range g.Items {
			g.Items[i].remove = nil
		}
		return g, ok
	case in.missing:
		return folderGroup(cacheDir, in.path+" (missing)", dir)
	case in.dataDir == "":
		return Group{}, false
	}
	return taskGroup(cacheDir, in.path, in.dataDir)
}

// shared runs fn with the shared cache lock, waiting for it. A session holds the shared lock too, so a
// project can be cleaned while it has a session. MoveData holds the exclusive lock, so a delete never
// runs during a move.
func shared(cacheDir string, fn func() (result, error)) (result, error) {
	release, err := filelock.RLock(filepath.Join(cacheDir, cacheLock))
	if err != nil {
		return result{}, err
	}
	defer func() { _ = release() }()
	return fn()
}

// folderGroup returns the group with one item that deletes the folder dir but its lock files.
func folderGroup(cacheDir, name, dir string) (Group, bool) {
	entries, _ := os.ReadDir(dir)
	var paths []string
	for _, e := range entries {
		if e.Name() != docsLock {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	if len(paths) == 0 {
		return Group{}, false
	}
	it := Item{Group: name, Label: "All data", Size: size(paths)}
	it.remove = func() (result, error) {
		return shared(cacheDir, func() (result, error) {
			var res result
			err := docfile.Locked(filepath.Join(dir, docsLock), func() (err error) {
				res, err = removeAll(paths)
				return err
			})
			return res, err
		})
	}
	return Group{Name: name, Items: []Item{it}}, true
}

// taskKey identifies a task group.
type taskKey struct{ spec, status string }

// taskFiles are the task files of one task group.
type taskFiles struct {
	paths   []string
	size    int64
	dataDir string
}

// taskGroup returns the group with one item for each spec and status of the local tasks in dataDir.
// It skips symbolic links.
func taskGroup(cacheDir, name, dataDir string) (Group, bool) {
	root := filepath.Join(dataDir, project.Tasks)
	if !realDirectory(dataDir) || !realDirectory(root) {
		return Group{}, false
	}
	found := findTasks(root, dataDir)
	keys := make([]taskKey, 0, len(found))
	for k := range found {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b taskKey) int {
		return cmp.Or(cmp.Compare(a.spec, b.spec), cmp.Compare(slices.Index(task.Statuses, a.status), slices.Index(task.Statuses, b.status)))
	})
	lock := filepath.Join(dataDir, docsLock)
	g := Group{Name: name}
	for _, k := range keys {
		g.Items = append(g.Items, taskItem(cacheDir, name, lock, k, found[k]))
	}
	return g, len(g.Items) > 0
}

// findTasks returns the task files in the spec folders of root, by spec and status. It skips symbolic links.
func findTasks(root, dataDir string) map[taskKey]*taskFiles {
	found := map[taskKey]*taskFiles{}
	specs, _ := os.ReadDir(root)
	for _, s := range specs {
		if !s.IsDir() {
			continue
		}
		files, _ := os.ReadDir(filepath.Join(root, s.Name()))
		for _, f := range files {
			if !f.Type().IsRegular() || filepath.Ext(f.Name()) != ".md" {
				continue
			}
			path := filepath.Join(root, s.Name(), f.Name())
			k := taskKey{s.Name(), readStatus(path)}
			if found[k] == nil {
				found[k] = &taskFiles{dataDir: dataDir}
			}
			found[k].paths = append(found[k].paths, path)
			found[k].size += size([]string{path})
		}
	}
	return found
}

// taskItem returns the item that deletes the files of one task group with the docs lock.
func taskItem(cacheDir, name, lock string, k taskKey, files *taskFiles) Item {
	noun := "tasks"
	if len(files.paths) == 1 {
		noun = "task"
	}
	return Item{
		Group: name,
		Label: fmt.Sprintf("%s %s: %d %s", k.spec, k.status, len(files.paths), noun),
		Size:  files.size,
		remove: func() (result, error) {
			return shared(cacheDir, func() (res result, err error) {
				err = docfile.Locked(lock, func() error {
					res, err = removeTasks(files, k.status)
					return err
				})
				return res, err
			})
		},
	}
}

// realDirectory reports whether path is a directory and not a symbolic link.
func realDirectory(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.IsDir()
}

// removeTasks removes the task files that still have the status, and then their folders if the
// folders are empty. A file that is gone is ignored. It skips a file if the status changed after the
// scan, or if the data directory, the tasks folder or the spec folder of the file is not a real
// directory any more (for example, a symbolic link). It does not remove such a folder.
func removeTasks(files *taskFiles, status string) (res result, err error) {
	var errs []error
	dirs := realDirs{base: realDirectory(files.dataDir) && realDirectory(filepath.Join(files.dataDir, project.Tasks)), seen: map[string]bool{}}
	for _, p := range files.paths {
		ok := dirs.ok(filepath.Dir(p))
		if _, err := os.Lstat(p); err != nil {
			continue
		}
		if !ok || readStatus(p) != status {
			res.skipped++
			continue
		}
		n, err := removeFile(p)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		res.removed++
		res.bytes += n
	}
	dirs.removeEmpty()
	return res, errors.Join(errs...)
}

// realDirs remembers which spec folders are real directories. base is false if the data directory
// or the tasks folder is not.
type realDirs struct {
	base bool
	seen map[string]bool
}

func (r realDirs) ok(dir string) bool {
	ok, known := r.seen[dir]
	if !known {
		ok = r.base && realDirectory(dir)
		r.seen[dir] = ok
	}
	return ok
}

// removeEmpty removes the real spec folders that are empty. It never removes a link.
func (r realDirs) removeEmpty() {
	for dir, ok := range r.seen {
		if ok {
			_ = os.Remove(dir)
		}
	}
}

// removeFile removes the file p and returns its size.
func removeFile(p string) (int64, error) {
	n := size([]string{p})
	return n, os.Remove(p)
}

// readStatus reads the status field of the frontmatter of a task file.
func readStatus(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	front, _ := markdown.Split(string(data))
	var f struct {
		Status string `yaml:"status"`
	}
	if markdown.Decode(front, &f) != nil || strings.TrimSpace(f.Status) == "" {
		return "unknown"
	}
	return f.Status
}

// removeAll removes each path that exists and continues after an error. The result has only the paths
// that it removed, with the size that they had just before.
func removeAll(paths []string) (res result, err error) {
	var errs []error
	for _, p := range paths {
		if _, err := os.Lstat(p); err != nil {
			continue
		}
		n := size([]string{p})
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, err)
			continue
		}
		res.removed++
		res.bytes += n
	}
	return res, errors.Join(errs...)
}

// size returns the bytes of the files under paths, but not lock files.
func size(paths []string) int64 {
	var total int64
	for _, p := range paths {
		_ = filepath.WalkDir(p, func(_ string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || d.Name() == docsLock || d.Name() == cacheLock {
				return nil
			}
			if fi, err := d.Info(); err == nil {
				total += fi.Size()
			}
			return nil
		})
	}
	return total
}
