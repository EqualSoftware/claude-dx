package git

import (
	"context"
	"errors"
	"strings"
)

// Sync fetches with prune and rebases the current branch onto its upstream. It keeps the local changes with
// autostash (git-tools#R-9).
func (r Runner) Sync(ctx context.Context, dir string) (string, error) {
	branch, err := r.currentBranch(ctx, dir, "git_sync")
	if err != nil {
		return "", err
	}
	out, err := r.Run(ctx, dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	upstream := strings.TrimSpace(out)
	if err != nil || upstream == "" {
		return "", errors.New("git_sync refused: the branch " + branch + " has no upstream. Push it with git_push first")
	}
	if _, err := r.RunNetwork(ctx, dir, "fetch", "--prune"); err != nil {
		return "", err
	}
	_, err = r.Run(ctx, dir, "rebase", "--autostash", upstream)
	return r.rebaseResult(ctx, dir, "git_sync", "Synced "+branch+" onto "+upstream+".", err)
}
