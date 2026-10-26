// Package numbering allocates ADR and task numbers with no duplicates on one machine (ADR-019).
package numbering

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
)

// Allocator keeps the counters of one project in project.json of Dir, the project data directory.
type Allocator struct {
	Dir string
}

// Next allocates max(counter, highest number in docDir) + 1 for kind and calls create with it.
// pattern must capture the number in its first group. create must make the file with
// exclusive create. The counter changes only if create succeeds. The caller must hold the docs
// lock of the project (docfile.Locked), because the docs lock also protects the counter (ADR-015).
func (a Allocator) Next(kind, docDir string, pattern *regexp.Regexp, create func(n int) error) (int, error) {
	if err := os.MkdirAll(a.Dir, 0o755); err != nil {
		return 0, err
	}
	counters, err := a.read()
	if err != nil {
		return 0, err
	}
	highest, err := Highest(docDir, pattern)
	if err != nil {
		return 0, err
	}
	n := max(counters[kind], highest) + 1
	if err := create(n); err != nil {
		return 0, err
	}
	counters[kind] = n
	return n, a.write(counters)
}

// Highest returns the highest number in the file names of dir, or 0.
// A directory that does not exist has no numbers.
func Highest(dir string, pattern *regexp.Regexp) (int, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	highest := 0
	for _, entry := range entries {
		m := pattern.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		highest = max(highest, n)
	}
	return highest, nil
}

// projectFile is the Project file (data) of ADR-018: project.json in the project data directory.
type projectFile struct {
	Counters map[string]int `json:"counters"`
}

func (a Allocator) path() string {
	return filepath.Join(a.Dir, "project.json")
}

func (a Allocator) read() (map[string]int, error) {
	var f projectFile
	data, err := os.ReadFile(a.path())
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]int{}, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Counters == nil {
		f.Counters = map[string]int{}
	}
	return f.Counters, nil
}

// write replaces project.json atomically, so a crash never leaves a partial file.
func (a Allocator) write(counters map[string]int) error {
	data, err := json.Marshal(projectFile{Counters: counters})
	if err != nil {
		return err
	}
	return docfile.Write(a.path(), data)
}
