//go:build !windows

package launcher

import (
	"os"
	"syscall"
)

// The terminal sends SIGINT and SIGQUIT to all processes of the foreground process group, so claude gets
// them. A kill of this process sends SIGTERM or SIGHUP only to this process, so it sends them to claude.
var (
	ignored   = []os.Signal{os.Interrupt, syscall.SIGQUIT}
	forwarded = []os.Signal{syscall.SIGTERM, syscall.SIGHUP}
)
