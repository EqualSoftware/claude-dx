package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorktreeOptions are the options of Worktree.
type WorktreeOptions struct {
	Action       string
	Worktree     string
	Branch       string
	Start        string
	DeleteBranch bool
}

type worktreeEntry struct {
	path   string
	hash   string
	branch string
}

// worktrees returns the worktrees of the repository. The branch is empty for a detached HEAD.
func (r Runner) worktrees(ctx context.Context, dir string) ([]worktreeEntry, error) {
	out, err := r.Run(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var list []worktreeEntry
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		key, value, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			list = append(list, worktreeEntry{path: filepath.FromSlash(value)})
		case "HEAD":
			list[len(list)-1].hash = value
		case "branch":
			list[len(list)-1].branch = strings.TrimPrefix(value, "refs/heads/")
		}
	}
	return list, nil
}

// Worktree lists, adds or removes worktrees (git-tools#R-22). An empty action lists.
func (r Runner) Worktree(ctx context.Context, dir string, o WorktreeOptions) (string, error) {
	if o.Action == "" || o.Action == "list" {
		return r.worktreeList(ctx, dir)
	}
	if err := errors.Join(noDash(o.Branch), noDash(o.Start)); err != nil || o.Worktree == "" {
		return "", errors.New("git_worktree refused: give worktree, and a branch and a start point that do not start with a dash")
	}
	switch o.Action {
	case "add":
		return r.worktreeAdd(ctx, dir, o)
	case "remove":
		return r.worktreeRemove(ctx, dir, o)
	}
	return "", fmt.Errorf("git_worktree refused: action %q is not list, add or remove", o.Action)
}

func (r Runner) worktreeList(ctx context.Context, dir string) (string, error) {
	list, err := r.worktrees(ctx, dir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, w := range list {
		branch := w.branch
		if branch == "" {
			branch = "(detached)"
		}
		fmt.Fprintf(&b, "%s %s %s\n", w.path, shortHash(w.hash), branch)
	}
	return b.String(), nil
}

func (r Runner) worktreeAdd(ctx context.Context, dir string, o WorktreeOptions) (string, error) {
	if o.Branch == "" {
		return "", errors.New("git_worktree refused: give branch, the name of the new branch")
	}
	args := []string{"worktree", "add", "-b", o.Branch, "--", o.Worktree}
	start := o.Start
	if start != "" {
		args = append(args, start)
	} else {
		start = "HEAD"
	}
	if _, err := r.Run(ctx, dir, args...); err != nil {
		return "", err
	}
	return "Added worktree " + o.Worktree + " on the new branch " + o.Branch + " from " + start + ".", nil
}

func (r Runner) worktreeRemove(ctx context.Context, dir string, o WorktreeOptions) (string, error) {
	path := filepath.FromSlash(o.Worktree)
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	entry, err := r.findWorktree(ctx, path)
	if err != nil {
		return "", err
	}
	status, err := r.Run(ctx, entry.path, "status", "--porcelain")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(status) != "" {
		return "", errors.New("git_worktree refused: " + o.Worktree + " has changes that are not committed")
	}
	if _, err := r.Run(ctx, r.Root, "worktree", "remove", "--", entry.path); err != nil {
		return "", err
	}
	done := "Removed worktree " + o.Worktree + "."
	if !o.DeleteBranch || entry.branch == "" {
		return done, nil
	}
	if _, err := r.Run(ctx, r.Root, "branch", "-d", entry.branch); err != nil {
		return done + " The branch " + entry.branch + " was not deleted: " + err.Error(), nil
	}
	return "Removed worktree " + o.Worktree + " and deleted branch " + entry.branch + ".", nil
}

// findWorktree returns the worktree at path. It refuses the main worktree and a path that is not a worktree.
func (r Runner) findWorktree(ctx context.Context, path string) (worktreeEntry, error) {
	list, err := r.worktrees(ctx, r.Root)
	if err != nil {
		return worktreeEntry{}, err
	}
	want, err := os.Stat(path)
	if err != nil {
		return worktreeEntry{}, errors.New("git_worktree refused: " + path + " is not a worktree of this repository")
	}
	for i, w := range list {
		got, err := os.Stat(w.path)
		if err != nil || !os.SameFile(want, got) {
			continue
		}
		if i == 0 {
			return worktreeEntry{}, errors.New("git_worktree refused: " + path + " is the main worktree")
		}
		return w, nil
	}
	return worktreeEntry{}, errors.New("git_worktree refused: " + path + " is not a worktree of this repository")
}
