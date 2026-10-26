//go:build !windows

package session

import (
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// inherit clears close-on-exec, so claude gets the descriptor and its lock after exec.
func inherit(f *os.File) error {
	_, err := unix.FcntlInt(f.Fd(), unix.F_SETFD, 0)
	return err
}

// alive reports whether a process has pid. A child of claude that runs after claude ends keeps the
// inherited lock, so a held session file can be of an ended session (TASK-050). Only ESRCH means
// that no process has pid: another error (example: EPERM) counts as alive, because a wrong removal
// is worse than a stale file.
func alive(pid string) bool {
	n, err := strconv.Atoi(pid)
	if err != nil {
		return true
	}
	return unix.Kill(n, 0) != unix.ESRCH
}
