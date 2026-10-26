// Command claude-dx starts Claude Code with the dx tools.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/cli"
	"github.com/EqualSoftware/claude-dx/internal/doctor"
	"github.com/EqualSoftware/claude-dx/internal/launcher"
	"github.com/EqualSoftware/claude-dx/internal/mcpserver"
	"github.com/EqualSoftware/claude-dx/internal/plugin"
	"github.com/EqualSoftware/claude-dx/internal/update"
)

// version is set at build time with -ldflags "-X main.version=<tag>".
var version = "dev"

const help = `claude-dx: Claude Code launcher with dx tools.

Usage:
  claude-dx [claude arguments]     Start Claude Code with dx.
  claude-dx -- [claude arguments]  Pass all arguments to Claude Code.
  claude-dx update                 Install the latest release.
  claude-dx cache clean            Remove all data of claude-dx.
  claude-dx doctor                 Check the setup.
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
	case "update":
		return runUpdate(stdout, stderr)
	case "cache":
		return cleanCache(cmd.Args, stdout, stderr)
	case "doctor":
		return runDoctor(stdout, stderr)
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

// runUpdate installs the latest release in place of this binary (ADR-004).
func runUpdate(stdout, stderr io.Writer) int {
	dir, err := cache.Dir()
	if err != nil {
		return fail(stderr, err)
	}
	binary, err := launcher.Executable()
	if err != nil {
		return fail(stderr, err)
	}
	u := update.Updater{URL: update.LatestURL, Version: version, Cache: dir}
	msg, err := u.Update(context.Background(), binary)
	if err != nil {
		return fail(stderr, err)
	}
	fmt.Fprintln(stdout, msg)
	return 0
}

// cleanCache removes the cache. It fails if a path stays, for example a file in use on Windows.
func cleanCache(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] != "clean" {
		fmt.Fprintln(stderr, "Usage: claude-dx cache clean")
		return 1
	}
	dir, err := cache.Dir()
	if err != nil {
		return fail(stderr, err)
	}
	failed, err := cache.Clean(dir)
	if err != nil {
		return fail(stderr, err)
	}
	for _, line := range failed {
		fmt.Fprintf(stderr, "claude-dx: cannot remove %s\n", line)
	}
	if len(failed) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "Removed %s.\n", dir)
	return 0
}

func runDoctor(stdout, stderr io.Writer) int {
	dir, err := cache.Dir()
	if err != nil {
		return fail(stderr, err)
	}
	binary, err := launcher.Executable()
	if err != nil {
		return fail(stderr, err)
	}
	if !doctor.Run(stdout, plugin.DirName(version, binary), dir) {
		return 1
	}
	return 0
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "claude-dx: %v\n", err)
	return 1
}
