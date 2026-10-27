// Package numbering allocates ADR and task numbers with no duplicates on one machine (ADR-007).
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
// lock of the project (docfile.Locked), because the docs lock also protects the counter (ADR-006).
func (a Allocator) Next(kind, docDir string, pattern *regexp.Regexp, create func(n int) error) (int, error) {
	if err := os.MkdirAll(a.Dir, 0o755); err != nil {
		return 0, err
	}
	file, counters, err := a.read()
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
	return n, a.write(file, counters)
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

func (a Allocator) path() string {
	return filepath.Join(a.Dir, "project.json")
}

// read returns all keys of project.json and the counters. The file has other keys: outside git, the
// Project file (cache) of ADR-007 is the same file.
func (a Allocator) read() (file map[string]json.RawMessage, counters map[string]int, err error) {
	file = map[string]json.RawMessage{}
	counters = map[string]int{}
	data, err := os.ReadFile(a.path())
	if errors.Is(err, fs.ErrNotExist) {
		return file, counters, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, nil, err
	}
	if file == nil { // the JSON value null
		file = map[string]json.RawMessage{}
	}
	if raw, ok := file["counters"]; ok {
		if err := json.Unmarshal(raw, &counters); err != nil {
			return nil, nil, err
		}
	}
	if counters == nil {
		counters = map[string]int{}
	}
	return file, counters, nil
}

// write replaces project.json atomically, so a crash never leaves a partial file. It changes only
// the key "counters" and keeps all other keys.
func (a Allocator) write(file map[string]json.RawMessage, counters map[string]int) error {
	raw, err := json.Marshal(counters)
	if err != nil {
		return err
	}
	file["counters"] = raw
	data, err := json.Marshal(file)
	if err != nil {
		return err
	}
	return docfile.Write(a.path(), data)
}
