// Package report keeps the log of unmapped git operations and opens the issue form (git-tools#R-16, R-17).
package report

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// LogName is the file name of the log in the cache directory.
const LogName = "unmapped.log"

// Entry is one line of the log: one git_exec call.
type Entry struct {
	Time   time.Time `json:"time"`
	Sub    string    `json:"sub"`
	Args   []string  `json:"args"`
	Reason string    `json:"reason"`
}

// Append adds one JSON line to the log at path. It writes the line with one write call on a file
// opened for append, so lines of two processes do not mix.
func Append(path string, e Entry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return errors.Join(err, f.Close())
}

// Read returns the entries of the log, oldest first. It skips lines that are not valid.
// A log that does not exist gives no entries.
func Read(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var entries []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		var e Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.Sub != "" {
			entries = append(entries, e)
		}
	}
	return entries, sc.Err()
}

// Clear empties the log at path. A log that does not exist is already empty.
func Clear(path string) error {
	err := os.Truncate(path, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
