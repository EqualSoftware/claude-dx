// Package mcpserver runs the dx MCP server on stdio (ADR-002).
package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/docs/adr"
	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/docs/glossary"
	"github.com/EqualSoftware/claude-dx/internal/docs/learning"
	"github.com/EqualSoftware/claude-dx/internal/docs/spec"
	"github.com/EqualSoftware/claude-dx/internal/docs/task"
	"github.com/EqualSoftware/claude-dx/internal/numbering"
	"github.com/EqualSoftware/claude-dx/internal/project"
	"github.com/EqualSoftware/claude-dx/internal/session"
	"github.com/EqualSoftware/claude-dx/internal/update"
)

// New returns the dx server with all tools for the project that contains dir.
func New(version, dir string) (*mcp.Server, error) {
	p, err := project.Open(dir)
	if err != nil {
		return nil, err
	}
	cacheDir, err := cache.Dir()
	if err != nil {
		return nil, err
	}
	alloc := numbering.Allocator{Dir: p.DataDir(cacheDir)}
	lock := filepath.Join(alloc.Dir, project.DocsLock)
	// The file is only information for "claude-dx clean", so an error does not stop the server.
	_ = writeProjectFile(filepath.Join(cacheDir, "projects", p.Key, "project.json"), mainRoot(p), lock)
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: version}, nil)
	adr.Register(server, adr.Store{Dir: filepath.Join(p.Root, "docs", "decisions"), Alloc: alloc, Author: gitUser(p.Root), Lock: lock})
	specs := spec.Store{Dir: filepath.Join(p.Root, "docs", "specs"), ADRDir: filepath.Join(p.Root, "docs", "decisions"),
		Root: p.Root, Author: gitUser(p.Root), Lock: lock}
	task.Register(server, task.Store{Dir: filepath.Join(alloc.Dir, project.Tasks), Specs: specs, Lock: lock})
	spec.Register(server, specs)
	learning.Register(server, learning.Store{Dir: filepath.Join(p.Root, "docs", "learnings"), Lock: lock})
	glossary.Register(server, glossary.Store{Path: filepath.Join(p.Root, "docs", "GLOSSARY.md"), Lock: lock})
	return server, nil
}

// Run serves MCP on stdin and stdout until the client disconnects. It holds a shared cache lock
// while it runs, so "clean" disables the global items while a session runs (ADR-006).
// Claude Code starts the server in the directory of the session.
func Run(ctx context.Context, version string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	cacheDir, err := cache.Dir()
	if err != nil {
		return err
	}
	release, err := session.Use(cacheDir)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()
	server, err := New(version, dir)
	if err != nil {
		return err
	}
	go update.Updater{URL: update.LatestURL, Version: version, Cache: cacheDir}.Check(ctx)
	return server.Run(ctx, &mcp.StdioTransport{})
}

// gitUser returns a function that reads the git user name of the project. It runs at each call,
// because the user can change the config during a session.
func gitUser(root string) func() string {
	return func() string {
		out, err := exec.Command("git", "-C", root, "config", "user.name").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
}

// mainRoot returns the root of the main worktree of the repository, so the repository is not missing
// when a linked worktree is removed. The common directory of the repository is <main root>/.git. Else
// (outside a repository, a bare repository, a submodule) it returns the root of the project.
func mainRoot(p project.Project) string {
	if p.Common == "" || filepath.Base(p.Common) != ".git" {
		return p.Root
	}
	return filepath.Dir(p.Common)
}

// writeProjectFile writes the Project file (cache) of ADR-007 with the project root and the last use
// date, with the docs lock of the project. It changes only these two keys and keeps the others: outside git,
// the number counters are in the same file.
func writeProjectFile(path, root, lock string) error {
	return docfile.Locked(lock, func() error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		file := map[string]json.RawMessage{}
		if data, err := os.ReadFile(path); err == nil {
			if err := json.Unmarshal(data, &file); err != nil {
				_ = docfile.MoveAside(path) // same as the number counters: keep the bad file
			}
			if file == nil {
				file = map[string]json.RawMessage{} // the value null has no data to keep
			}
		}
		file["path"], _ = json.Marshal(root)
		file["last_used"], _ = json.Marshal(time.Now().UTC().Format(time.RFC3339))
		data, err := json.Marshal(file)
		if err != nil {
			return err
		}
		return docfile.Write(path, data)
	})
}
