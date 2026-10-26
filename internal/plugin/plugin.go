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
)

// The "all:" prefix includes files that start with "." (example: .claude-plugin).
//
//go:embed all:assets
var assets embed.FS

// Ensure extracts the plugin to <cacheDir>/plugin/<DirName> and returns that directory.
// A directory that is already extracted is used again.
// binary is the absolute path of claude-dx. It replaces {{BINARY}} in the plugin files.
func Ensure(cacheDir, version, binary string) (string, error) {
	root := filepath.Join(cacheDir, "plugin")
	dir := filepath.Join(root, DirName(version, binary))
	if exists(dir) {
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
	return dir, install(tmp, dir)
}

// DirName returns the name of the plugin directory in <cache>/plugin. A release uses its version.
// A "dev" build uses dev-<key>, with a key of the embedded files and the binary path, so a new dev
// build does not change the files of a session that runs.
func DirName(version, binary string) string {
	if version != "dev" {
		return version
	}
	h := sha256.New()
	h.Write([]byte(binary))
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

// install moves the extracted tmp directory to dir.
// Another process can extract the same directory at the same time. Its result is equal, so it is used.
func install(tmp, dir string) error {
	err := os.Rename(tmp, dir)
	if err != nil && exists(dir) {
		return nil
	}
	return err
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

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
