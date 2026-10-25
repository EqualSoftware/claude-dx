//go:build !windows

package launcher

import (
	"os"
	"syscall"
)

// start replaces this process with claude, so signals, TTY and exit code are the same as a direct start.
func start(claude string, args []string) (int, error) {
	argv := append([]string{claude}, args...)
	return 1, syscall.Exec(claude, argv, os.Environ())
}
