// Package plugin extracts the embedded dx plugin to the cache (ADR-001).
package plugin

import (
	"embed"
	"encoding/json"
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

// Ensure extracts the plugin to <cacheDir>/plugin/<version> and returns that directory.
// A version that is already extracted is used again, except for "dev" builds.
// binary is the absolute path of claude-dx. It replaces {{BINARY}} in the plugin files.
func Ensure(cacheDir, version, binary string) (string, error) {
	root := filepath.Join(cacheDir, "plugin")
	dir := filepath.Join(root, version)
	if version != "dev" && exists(dir) {
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
	return dir, install(tmp, dir, version)
}

// install moves the extracted tmp directory to dir.
// Another process can extract the same version at the same time. Its result is equal, so it is used.
func install(tmp, dir, version string) error {
	if version == "dev" {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
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
