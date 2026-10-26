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
)

// Allocator keeps the counters of one project in Dir (<cache>/projects/<key>).
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

func (a Allocator) path() string {
	return filepath.Join(a.Dir, "counters.json")
}

func (a Allocator) read() (map[string]int, error) {
	counters := map[string]int{}
	data, err := os.ReadFile(a.path())
	if errors.Is(err, fs.ErrNotExist) {
		return counters, nil
	}
	if err != nil {
		return nil, err
	}
	return counters, json.Unmarshal(data, &counters)
}

// write replaces counters.json atomically, so a crash never leaves a partial file.
func (a Allocator) write(counters map[string]int) error {
	data, err := json.Marshal(counters)
	if err != nil {
		return err
	}
	tmp := a.path() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.path())
}
