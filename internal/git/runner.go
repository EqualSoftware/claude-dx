// Package git runs the system git for the git_* tools (ADR-003, ADR-015).
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NetworkTimeout is the time limit of a network git command (git-tools#R-13).
const NetworkTimeout = 120 * time.Second

// errorLines is the number of stderr lines in a git error (git-tools#R-1).
const errorLines = 5

// Runner runs git in the project root, or in a worktree of the same repository.
type Runner struct {
	Root string
}

// Run runs git with args in dir and returns its standard output.
// It sets LC_ALL=C and GIT_TERMINAL_PROMPT=0, and turns off the pager and the color (git-tools#R-1).
func (r Runner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	return r.RunEnv(ctx, dir, nil, args...)
}

// RunEnv is Run with more environment variables, each as NAME=value.
func (r Runner) RunEnv(ctx context.Context, dir string, env []string, args ...string) (string, error) {
	full := append([]string{"--no-pager", "-c", "color.ui=never", "-c", "core.quotepath=false", "-C", dir}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat", "GIT_OPTIONAL_LOCKS=0")
	cmd.Env = append(cmd.Env, env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", failure(ctx, args, err, stderr.String())
	}
	return stdout.String(), nil
}

// RunNetwork is Run with the time limit NetworkTimeout (git-tools#R-13).
func (r Runner) RunNetwork(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, NetworkTimeout)
	defer cancel()
	return r.Run(ctx, dir, args...)
}

func failure(ctx context.Context, args []string, err error, stderr string) error {
	sub := "command"
	if len(args) > 0 {
		sub = args[0]
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("git %s failed: timeout", sub)
	}
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	if len(lines) > errorLines {
		lines = lines[:errorLines]
	}
	text := strings.TrimSpace(strings.Join(lines, "\n"))
	if text == "" {
		text = err.Error()
	}
	return fmt.Errorf("git %s failed: %s", sub, text)
}

// Dir returns the directory where a tool runs git. An empty path gives the project root. Else the path
// must be the project root or a worktree of the same repository (git-tools#R-21, ADR-015).
func (r Runner) Dir(ctx context.Context, path string) (string, error) {
	if path == "" {
		return r.Root, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	want, err := os.Stat(abs)
	if err != nil {
		return "", errNotWorktree(path)
	}
	out, err := r.Run(ctx, r.Root, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		dir, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "worktree ")
		if !ok {
			continue
		}
		if got, err := os.Stat(filepath.FromSlash(dir)); err == nil && os.SameFile(want, got) {
			return abs, nil
		}
	}
	return "", errNotWorktree(path)
}

func errNotWorktree(path string) error {
	return fmt.Errorf("path %s is not the project root or a worktree of this repository", path)
}
