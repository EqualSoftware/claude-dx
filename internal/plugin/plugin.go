// Package plugin extracts the embedded dx plugin to the cache (ADR-001).
package plugin

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// The "all:" prefix includes files that start with "." (example: .claude-plugin).
//
//go:embed all:assets
var assets embed.FS

// Ensure extracts the plugin to <cacheDir>/plugin/<DirName> and returns that directory.
// A directory that is already extracted is used again.
// binary is the absolute path of claude-dx. It replaces {{BINARY}} in the plugin files.
func Ensure(cacheDir, version, binary string) (string, error) {
	if err := checkBinary(binary); err != nil {
		return "", err
	}
	root := filepath.Join(cacheDir, "plugin")
	dir := filepath.Join(root, DirName(version, binary))
	if complete(dir) {
		return dir, nil
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(root, version+".tmp-")
	if err != nil {
		return "", err
	}
	// Best effort: after a rename, tmp does not exist.
	defer func() { _ = os.RemoveAll(tmp) }()
	if err := write(tmp, version, binary); err != nil {
		return "", err
	}
	return dir, install(tmp, dir, filepath.Join(cacheDir, "plugin.repair.lock"))
}

// DirName returns the name of the plugin directory in <cache>/plugin: <version>-<key>, with a key of the
// binary path, because the plugin files start that path (plugin#R-9). A "dev" build also puts the embedded
// files in the key, so a new dev build does not change the files of a session that runs.
func DirName(version, binary string) string {
	h := sha256.New()
	h.Write([]byte(binary))
	if version != "dev" {
		return version + "-" + hex.EncodeToString(h.Sum(nil)[:8])
	}
	_ = fs.WalkDir(assets, "assets", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := assets.ReadFile(name)
		fmt.Fprintf(h, "\x00%s\x00%d\x00", name, len(data))
		h.Write(data)
		return err
	})
	return "dev-" + hex.EncodeToString(h.Sum(nil)[:8])
}

// checkBinary refuses a path that the shell of the hook command reads in a different way inside double
// quotes: ", $, a backtick, a line break and, for a UNC path, a start with two backslashes.
func checkBinary(binary string) error {
	if strings.ContainsAny(binary, "\"$`\r\n") || strings.HasPrefix(binary, `\\`) {
		return fmt.Errorf("the path of claude-dx has a character that the hook shell does not accept (double quote, $, backtick, line break or a start with two backslashes): %q", binary)
	}
	return nil
}

// install moves the extracted tmp directory to dir.
// Another process can extract the same directory at the same time. Its result is equal, so it is used.
// If the rename fails and dir is not complete, it comes from a stop during a write or from a user change.
// Then install takes the lock, so only one process repairs, and moves dir aside, not deleted. The session
// cleanup removes the moved directory. The lock file is outside the plugin directory, so the cleanup does
// not remove it.
func install(tmp, dir, lockPath string) error {
	err := os.Rename(tmp, dir)
	if err == nil || complete(dir) {
		return nil
	}
	unlock, err := filelock.Lock(lockPath)
	if err != nil {
		return err
	}
	defer func() { _ = unlock() }()
	if complete(dir) {
		return nil
	}
	if exists(dir) {
		aside := fmt.Sprintf("%s.old-%d-%d", dir, os.Getpid(), time.Now().UnixNano())
		if err := os.Rename(dir, aside); err != nil {
			return err
		}
	}
	// A process that does not take the lock can put its directory here after the move.
	if err := os.Rename(tmp, dir); err != nil && !complete(dir) {
		return err
	}
	return nil
}

// write copies the embedded files to dir and fills the placeholders.
func write(dir, version, binary string) error {
	replacer := strings.NewReplacer("{{VERSION}}", version, "{{BINARY}}", jsonEscape(binary))
	return fs.WalkDir(assets, "assets", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(name, "assets")))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := assets.ReadFile(name)
		if err != nil {
			return err
		}
		if path.Ext(name) == ".json" {
			data = []byte(replacer.Replace(string(data)))
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// jsonEscape returns s escaped for use inside a JSON string. Windows paths have backslashes.
func jsonEscape(s string) string {
	quoted, _ := json.Marshal(s)
	return string(quoted[1 : len(quoted)-1])
}

// required lists the files that a plugin directory must have.
var required = []string{".claude-plugin/plugin.json", ".mcp.json", "hooks/hooks.json", "rules.md"}

// complete reports whether dir has all the required files.
func complete(dir string) bool {
	for _, name := range required {
		if !exists(filepath.Join(dir, filepath.FromSlash(name))) {
			return false
		}
	}
	return true
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
