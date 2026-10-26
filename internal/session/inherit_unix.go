//go:build !windows

package session

import (
	"os"

	"golang.org/x/sys/unix"
)

// inherit clears close-on-exec, so claude gets the descriptor and its lock after exec.
func inherit(f *os.File) error {
	_, err := unix.FcntlInt(f.Fd(), unix.F_SETFD, 0)
	return err
}
