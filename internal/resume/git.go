package resume

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// gitTimeout is the longest time of one git command. A hook must not hang.
// It is a variable for the tests.
var gitTimeout = 5 * time.Second

// errNoRepo means that the directory is not in a git repository (R-10).
var errNoRepo = errors.New("no git repository")

// refsSep separates the Refs trailer values in the output of git log.
const refsSep = "\x1f"

// gitState is the state of the git repository. It comes from local commands only (R-2).
type gitState struct {
	branch   string   // empty when HEAD is detached
	hash     string   // short hash of HEAD, set when HEAD is detached
	upstream string   // empty when the branch has no upstream
	changed  []string // short status lines, for example "M a.go"
	unpushed []commit
}

// commit is one unpushed commit.
type commit struct {
	hash    string
	subject string
	refs    []string
}

// git runs a read-only git command in dir and returns its output with no final line break.
func git(dir string, args ...string) (string, error) {
	full := append([]string{"--no-optional-locks", "-C", dir, "-c", "core.quotePath=false"}, args...)
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", full...).Output()
	return strings.TrimRight(string(out), "\r\n"), err
}

// readGit reads the git state of dir. It returns errNoRepo when dir is not in a work tree.
func readGit(dir string) (gitState, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return gitState{}, err
	}
	if inside, err := git(dir, "rev-parse", "--is-inside-work-tree"); err != nil || inside != "true" {
		return gitState{}, errNoRepo
	}
	var g gitState
	status, err := git(dir, "status", "--porcelain")
	if err != nil {
		return g, err
	}
	g.changed = shortStatus(status)
	branch, err := git(dir, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		g.hash, err = git(dir, "rev-parse", "--short", "HEAD")
		return g, err
	}
	g.branch = branch
	if _, err := git(dir, "rev-parse", "-q", "--verify", "HEAD"); err != nil {
		return g, nil // a branch with no commit has nothing to push
	}
	g.upstream, _ = git(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	g.unpushed, err = unpushed(dir, g.upstream != "")
	return g, err
}

// shortStatus turns the lines of "git status --porcelain" into "<codes> <path>".
func shortStatus(out string) []string {
	var list []string
	for _, line := range splitLines(out) {
		if len(line) < 4 {
			continue
		}
		list = append(list, strings.TrimSpace(line[:2])+" "+line[3:])
	}
	return list
}

// unpushed lists the commits that the upstream does not have or, with no upstream, that no remote
// branch has (R-8).
func unpushed(dir string, upstream bool) ([]commit, error) {
	format := "--format=%h%x00%s%x00%(trailers:key=Refs,valueonly,unfold,separator=%x1f)"
	args := []string{"log", format, "HEAD", "--not", "--remotes"}
	if upstream {
		args = []string{"log", format, "@{u}..HEAD"}
	}
	out, err := git(dir, args...)
	if err != nil {
		return nil, err
	}
	var list []commit
	for _, line := range splitLines(out) {
		c, err := parseCommit(line)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// parseCommit reads one line of the format of unpushed.
func parseCommit(line string) (commit, error) {
	parts := strings.Split(line, "\x00")
	if len(parts) != 3 {
		return commit{}, fmt.Errorf("unexpected git log line %q", line)
	}
	c := commit{hash: parts[0], subject: parts[1]}
	for _, ref := range strings.Split(parts[2], refsSep) {
		if ref = strings.TrimSpace(ref); ref != "" {
			c.refs = append(c.refs, ref)
		}
	}
	return c, nil
}

// splitLines splits text at line breaks and drops the empty lines.
func splitLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
