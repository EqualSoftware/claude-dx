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
	lock := filepath.Join(alloc.Dir, "docs.lock")
	// The file is only information for "claude-dx clean", so an error does not stop the server.
	_ = writeProjectFile(filepath.Join(cacheDir, "projects", p.Key, "project.json"), p.Root, lock)
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: version}, nil)
	adr.Register(server, adr.Store{Dir: filepath.Join(p.Root, "docs", "decisions"), Alloc: alloc, Lock: lock})
	task.Register(server, task.Store{Dir: filepath.Join(p.Root, "docs", "tasks"), Alloc: alloc, Author: gitUser(p.Root), Lock: lock})
	spec.Register(server, spec.Store{Dir: filepath.Join(p.Root, "docs", "specs"), ADRDir: filepath.Join(p.Root, "docs", "decisions"),
		Root: p.Root, Author: gitUser(p.Root), Lock: lock})
	learning.Register(server, learning.Store{Dir: filepath.Join(p.Root, "docs", "learnings"), Lock: lock})
	glossary.Register(server, glossary.Store{Path: filepath.Join(p.Root, "docs", "GLOSSARY.md"), Lock: lock})
	return server, nil
}

// Run serves MCP on stdin and stdout until the client disconnects. It holds a shared cache lock
// while it runs, so "cache clean" does not remove the cache of a session (ADR-015).
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

// writeProjectFile writes the Project file (cache) of ADR-018 with the project root and the last use
// date, with the docs lock of the project.
func writeProjectFile(path, root, lock string) error {
	data, err := json.Marshal(struct {
		Path     string `json:"path"`
		LastUsed string `json:"last_used"`
	}{root, time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		return err
	}
	return docfile.Locked(lock, func() error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return docfile.Write(path, data)
	})
}
