// Command fakeclaude prints its arguments, one on each line, and exits with code 3.
// With only --version, it prints a Claude Code version and exits with code 0.
// With FAKECLAUDE_WAIT set, it first prints "ready" and waits for the end of stdin.
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("2.1.308 (Claude Code)")
		return
	}
	if os.Getenv("FAKECLAUDE_WAIT") != "" {
		fmt.Println("ready")
		_, _ = io.Copy(io.Discard, os.Stdin)
	}
	for _, arg := range os.Args[1:] {
		fmt.Println(arg)
	}
	os.Exit(3)
}
