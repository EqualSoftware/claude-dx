package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/git"
)

const (
	// defaultDepth is the depth of fs_tree when the caller gives none.
	defaultDepth = 3
	// maxLines is the maximum number of lines of fs_tree.
	maxLines = 500
)

// Tree shows an indented tree of the files below path (filesystem-tools#R-2). The files come from
// git ls-files. Without git, it walks all files except .git.
func (t Tools) Tree(ctx context.Context, path string, depth int) (string, error) {
	if path == "" {
		path = "."
	}
	if depth <= 0 {
		depth = defaultDepth
	}
	dir, err := t.path(path)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	files, err := listFiles(ctx, t.Project.Root, dir)
	if err != nil {
		return "", err
	}
	return render(files, depth), nil
}

// listFiles returns the files below dir, as slash paths relative to dir.
func listFiles(ctx context.Context, root, dir string) ([]string, error) {
	out, err := git.Runner{Root: root}.Run(ctx, dir, "ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return walk(dir)
	}
	var files []string
	for _, f := range strings.Split(out, "\n") {
		if f == "" {
			continue
		}
		if _, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(f))); err == nil {
			files = append(files, f)
		}
	}
	sort.Strings(files)
	return files, nil
}

func walk(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return files, nil
}

// render makes the indented lines. A directory has a trailing slash. A directory deeper than depth is not opened.
func render(files []string, depth int) string {
	var lines []string
	seen := map[string]bool{}
	for _, f := range files {
		parts := strings.Split(f, "/")
		for i := 0; i < len(parts) && i < depth; i++ {
			name := parts[i]
			if i < len(parts)-1 {
				name += "/"
				key := strings.Join(parts[:i+1], "/")
				if seen[key] {
					continue
				}
				seen[key] = true
			}
			lines = append(lines, strings.Repeat("  ", i)+name)
		}
	}
	if len(lines) == 0 {
		return "No files."
	}
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}
	hidden := len(lines) - maxLines
	return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n%d lines not shown.", hidden)
}
