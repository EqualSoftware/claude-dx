// Package mcpserver runs the dx MCP server on stdio (ADR-002).
package mcpserver

import (
	"context"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/is-equal/claude-dx/internal/cache"
	"github.com/is-equal/claude-dx/internal/docs/adr"
	"github.com/is-equal/claude-dx/internal/numbering"
	"github.com/is-equal/claude-dx/internal/project"
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
	return server.Run(ctx, &mcp.StdioTransport{})
}
