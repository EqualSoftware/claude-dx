package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RebaseOptions are the options of Rebase.
type RebaseOptions struct {
	Action string
	Onto   string
}

// rebaseRunning reports whether a rebase is in progress in dir.
func (r Runner) rebaseRunning(ctx context.Context, dir string) bool {
	for _, name := range []string{"rebase-merge", "rebase-apply"} {
		out, err := r.Run(ctx, dir, "rev-parse", "--git-path", name)
		if err != nil {
			continue
		}
		path := filepath.FromSlash(strings.TrimSpace(out))
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// Rebase starts, continues, aborts or skips a rebase (git-tools#R-10). start rewrites the commits that HEAD
// has and onto has not, and it asks the user if the oldest is on the remote (git-tools#R-7).
func (r Runner) Rebase(ctx context.Context, dir string, o RebaseOptions, confirm Confirm) (string, error) {
	switch o.Action {
	case "start":
		return r.rebaseStart(ctx, dir, o.Onto, confirm)
	case "continue", "abort", "skip":
		if !r.rebaseRunning(ctx, dir) {
			return "", errors.New("git_rebase refused: no rebase is in progress")
		}
	default:
		return "", fmt.Errorf("git_rebase refused: action %q is not start, continue, abort or skip", o.Action)
	}
	_, err := r.RunEnv(ctx, dir, []string{"GIT_EDITOR=true"}, "rebase", "--"+o.Action)
	if o.Action == "abort" {
		return "Rebase aborted.", err
	}
	return r.rebaseResult(ctx, dir, "git_rebase", "Rebase finished.", err)
}

func (r Runner) rebaseStart(ctx context.Context, dir, onto string, confirm Confirm) (string, error) {
	if onto == "" || noDash(onto) != nil {
		return "", errors.New("git_rebase refused: give onto, a ref that does not start with a dash")
	}
	if r.rebaseRunning(ctx, dir) {
		return "", errors.New("git_rebase refused: a rebase is already in progress")
	}
	out, err := r.Run(ctx, dir, "rev-list", "--reverse", "--end-of-options", onto+"..HEAD")
	if err != nil {
		return "", err
	}
	if oldest, _, _ := strings.Cut(strings.TrimSpace(out), "\n"); oldest != "" {
		question := fmt.Sprintf("The commit %s is on the remote. git_rebase rewrites it. Do you allow this?", shortHash(oldest))
		if err := r.GuardRemote(ctx, dir, oldest, "git_rebase", question, confirm); err != nil {
			return "", err
		}
	}
	_, err = r.Run(ctx, dir, "rebase", "--autostash", onto)
	return r.rebaseResult(ctx, dir, "git_rebase", "Rebase finished.", err)
}

// rebaseResult turns the end of a rebase command into a result. A stop on conflicts is a normal result with
// the conflicts, so that the caller can resolve them.
func (r Runner) rebaseResult(ctx context.Context, dir, tool, done string, runErr error) (string, error) {
	conflicts, err := r.Conflicts(ctx, dir)
	if err == nil && conflicts != "No conflicts." {
		return tool + " stopped: conflicts. Resolve them, stage the files, then run git_rebase action=continue.\n" + conflicts, nil
	}
	if runErr != nil {
		return "", runErr
	}
	head, err := r.headIs(ctx, dir)
	return done + " " + head, err
}
