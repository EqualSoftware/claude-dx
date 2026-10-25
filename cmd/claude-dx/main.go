// Command claude-dx starts Claude Code with the dx tools.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/is-equal/claude-dx/internal/cli"
	"github.com/is-equal/claude-dx/internal/launcher"
	"github.com/is-equal/claude-dx/internal/mcpserver"
)

// version is set at build time with -ldflags "-X main.version=<tag>".
var version = "dev"

const help = `claude-dx: Claude Code launcher with dx tools.

Usage:
  claude-dx [claude arguments]     Start Claude Code with dx.
  claude-dx -- [claude arguments]  Pass all arguments to Claude Code.
  claude-dx --version              Show the claude-dx version.
  claude-dx --help                 Show this help.

For Claude Code help, run: claude-dx -- --help
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	cmd := cli.Parse(args)
	switch cmd.Name {
	case "":
		return launch(cmd.Args, stderr)
	case "help":
		fmt.Fprint(stdout, help)
		return 0
	case "version":
		fmt.Fprintf(stdout, "claude-dx %s\n", version)
		return 0
	case "mcp":
		return serve(stderr)
	}
	fmt.Fprintf(stderr, "claude-dx: %q is not available yet.\n", args[0])
	return 1
}

func launch(args []string, stderr io.Writer) int {
	code, err := launcher.Run(version, args)
	if err != nil {
		fmt.Fprintf(stderr, "claude-dx: %v\n", err)
	}
	return code
}

// serve runs the MCP server. Claude Code starts it from the plugin (ADR-002).
func serve(stderr io.Writer) int {
	if err := mcpserver.Run(context.Background(), version); err != nil {
		fmt.Fprintf(stderr, "claude-dx mcp: %v\n", err)
		return 1
	}
	return 0
}
