// Command claude-dx starts Claude Code with the dx tools.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is set at build time with -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the command and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintf(stdout, "claude-dx %s\n", version)
		return 0
	}
	fmt.Fprintln(stderr, "claude-dx: the launcher is not available yet.")
	return 1
}
