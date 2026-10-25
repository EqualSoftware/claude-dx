// Package mcpserver runs the dx MCP server on stdio (ADR-002).
package mcpserver

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/docs/adr"
	"github.com/EqualSoftware/claude-dx/internal/docs/glossary"
	"github.com/EqualSoftware/claude-dx/internal/docs/learning"
	"github.com/EqualSoftware/claude-dx/internal/docs/task"
	"github.com/EqualSoftware/claude-dx/internal/numbering"
	"github.com/EqualSoftware/claude-dx/internal/project"
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
	alloc := numbering.Allocator{Dir: filepath.Join(cacheDir, "projects", p.Key)}
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: version}, nil)
	adr.Register(server, adr.Store{Dir: filepath.Join(p.Root, "docs", "decisions"), Alloc: alloc})
	task.Register(server, task.Store{Dir: filepath.Join(p.Root, "docs", "tasks"), Alloc: alloc, Author: gitUser(p.Root)})
	learning.Register(server, learning.Store{Dir: filepath.Join(p.Root, "docs", "learnings")})
	glossary.Register(server, glossary.Store{Path: filepath.Join(p.Root, "docs", "GLOSSARY.md")})
	return server, nil
}

// Run serves MCP on stdin and stdout until the client disconnects.
// Claude Code starts the server in the directory of the session.
func Run(ctx context.Context, version string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	server, err := New(version, dir)
	if err != nil {
		return err
	}
	if cacheDir, err := cache.Dir(); err == nil {
		go update.Updater{URL: update.LatestURL, Version: version, Cache: cacheDir}.Check(ctx)
	}
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
