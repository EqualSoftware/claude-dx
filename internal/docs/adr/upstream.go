package adr

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkUpstream refuses a change to an ADR that is Accepted or superseded in the upstream branch
// (ADR-021). It reads only the local ref of the upstream. Any git failure allows the change.
func checkUpstream(id, path string) error {
	// "./name" is relative to the directory of -C, so the path has no OS separators.
	cmd := exec.Command("git", "-C", filepath.Dir(path), "show", "@{upstream}:./"+filepath.Base(path))
	// A partial clone must not fetch from the network.
	cmd.Env = append(os.Environ(), "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil { // every git failure allows the change
		return nil
	}
	status := first(statusLine, out)
	if status != "Accepted" && !strings.HasPrefix(status, "Superseded by") {
		return nil
	}
	return errors.New(id + " is " + status + " in the upstream branch. A pushed decision does not change. Record a new ADR and use adr_supersede.")
}
