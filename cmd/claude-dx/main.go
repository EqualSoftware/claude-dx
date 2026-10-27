// Command claude-dx starts Claude Code with the dx tools.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/clean"
	"github.com/EqualSoftware/claude-dx/internal/cli"
	"github.com/EqualSoftware/claude-dx/internal/doctor"
	"github.com/EqualSoftware/claude-dx/internal/launcher"
	"github.com/EqualSoftware/claude-dx/internal/mcpserver"
	"github.com/EqualSoftware/claude-dx/internal/plugin"
	"github.com/EqualSoftware/claude-dx/internal/session"
	"github.com/EqualSoftware/claude-dx/internal/update"
)

// version is set at build time with -ldflags "-X main.version=<tag>".
var version = "dev"

const help = `claude-dx: Claude Code launcher with dx tools.

Usage:
  claude-dx [claude arguments]     Start Claude Code with dx.
  claude-dx -- [claude arguments]  Pass all arguments to Claude Code.
  claude-dx update                 Install the latest release.
  claude-dx clean                  Select the data of claude-dx to delete.
  claude-dx clean --list           Show the data of claude-dx and its sizes.
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
	case "clean":
		return runClean(cmd.Args, stdout, stderr)
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

// runUpdate installs the latest release in place of this binary (ADR-005).
func runUpdate(stdout, stderr io.Writer) int {
	dir, err := cache.Dir()
	if err != nil {
		return fail(stderr, err)
	}
	release, err := session.Use(dir)
	if err != nil {
		return fail(stderr, err)
	}
	defer func() { _ = release() }()
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

// runClean lists the data of claude-dx, or asks the person which data to delete (ADR-011). Only
// the list works without an interactive terminal.
func runClean(args []string, stdout, stderr io.Writer) int {
	list := len(args) == 1 && args[0] == "--list"
	if len(args) > 0 && !list {
		fmt.Fprintln(stderr, "Usage: claude-dx clean [--list]")
		return 1
	}
	if !list && !interactive() {
		fmt.Fprintln(stderr, "claude-dx clean needs an interactive terminal. Run claude-dx clean --list to see the data.")
		return 1
	}
	if err := cleanData(list, stdout); err != nil {
		return fail(stderr, err)
	}
	return 0
}

// cleanData lists the data, or asks the person which data to delete.
func cleanData(list bool, stdout io.Writer) error {
	dir, err := cache.Dir()
	if err != nil {
		return err
	}
	binary, err := launcher.Executable()
	if err != nil {
		return err
	}
	own := plugin.DirName(version, binary)
	if !list {
		return clean.Run(stdout, dir, own, clean.Interactive())
	}
	groups, err := clean.Scan(dir, own)
	if err != nil {
		return err
	}
	clean.List(stdout, groups)
	return nil
}

// interactive reports whether stdin and stdout are terminals.
func interactive() bool {
	return term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
}

func runDoctor(stdout, stderr io.Writer) int {
	dir, err := cache.Dir()
	if err != nil {
		return fail(stderr, err)
	}
	// With no lock, the cache is not writable, and the doctor checks report it.
	if release, err := session.Use(dir); err == nil {
		defer func() { _ = release() }()
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
