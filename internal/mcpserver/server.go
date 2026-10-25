// Package mcpserver runs the dx MCP server on stdio (ADR-002).
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New returns the dx server with all tools.
func New(version string) *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{Name: "dx", Version: version}, nil)
}

// Run serves MCP on stdin and stdout until the client disconnects.
func Run(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}
