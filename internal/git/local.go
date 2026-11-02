package git

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Stage adds paths to the index, or all changes with all (git-tools#R-4).
func (r Runner) Stage(ctx context.Context, dir string, paths []string, all bool) (string, error) {
	if all {
		_, err := r.Run(ctx, dir, "add", "-A")
		return "Staged all changes.", err
	}
	if len(paths) == 0 {
		return "", errors.New("git_stage refused: give paths, or all")
	}
	_, err := r.Run(ctx, dir, append([]string{"add", "--"}, paths...)...)
	return fmt.Sprintf("Staged %d paths.", len(paths)), err
}

// Unstage removes paths from the index, or all staged changes with all, and keeps the changes in the
// working tree (git-tools#R-25). Reset also works in a repository with no commit.
func (r Runner) Unstage(ctx context.Context, dir string, paths []string, all bool) (string, error) {
	if all && len(paths) > 0 {
		return "", errors.New("git_unstage refused: all and paths together. Use all with no paths")
	}
	if all {
		_, err := r.Run(ctx, dir, "reset", "-q")
		return "Unstaged all changes.", err
	}
	if len(paths) == 0 {
		return "", errors.New("git_unstage refused: give paths, or all")
	}
	_, err := r.Run(ctx, dir, append([]string{"restore", "--staged", "--"}, paths...)...)
	return fmt.Sprintf("Unstaged %d paths.", len(paths)), err
}

// Discard drops the changes of tracked paths in the working tree. It does not touch untracked files.
func (r Runner) Discard(ctx context.Context, dir string, paths []string) (string, error) {
	if len(paths) == 0 {
		return "", errors.New("git_discard refused: give paths")
	}
	if _, err := r.Run(ctx, dir, append([]string{"ls-files", "--error-unmatch", "--"}, paths...)...); err != nil {
		return "", errors.New("git_discard refused: a path is not tracked. It only discards changes of tracked files")
	}
	_, err := r.Run(ctx, dir, append([]string{"restore", "--worktree", "--"}, paths...)...)
	return fmt.Sprintf("Discarded %d paths.", len(paths)), err
}

// requireStaged refuses when the index has no change.
func (r Runner) requireStaged(ctx context.Context, dir, tool string) error {
	out, err := r.Run(ctx, dir, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "" {
		return errors.New(tool + " refused: nothing is staged. Stage files with git_stage, or give paths")
	}
	return nil
}

// Commit makes a commit of the staged changes with the message of ADR-003. With paths, it stages them first
// (git-tools#R-5).
func (r Runner) Commit(ctx context.Context, dir string, m Message, paths []string) (string, error) {
	args, err := m.Args("git_commit")
	if err != nil {
		return "", err
	}
	if len(paths) > 0 {
		if _, err := r.Stage(ctx, dir, paths, false); err != nil {
			return "", err
		}
	}
	if err := r.requireStaged(ctx, dir, "git_commit"); err != nil {
		return "", err
	}
	if _, err := r.Run(ctx, dir, append([]string{"commit"}, args...)...); err != nil {
		return "", err
	}
	return r.head(ctx, dir)
}

// head returns "<short-hash> <subject>" of HEAD.
func (r Runner) head(ctx context.Context, dir string) (string, error) {
	out, err := r.Run(ctx, dir, "log", "-1", "--format=%H %s")
	if err != nil {
		return "", err
	}
	hash, subject, _ := strings.Cut(strings.TrimSpace(out), " ")
	return shortHash(hash) + " " + subject, nil
}

func (r Runner) headIs(ctx context.Context, dir string) (string, error) {
	line, err := r.head(ctx, dir)
	return "HEAD is now " + line + ".", err
}

// guard asks the user before the tool rewrites ref, when ref is on the remote (git-tools#R-7).
func (r Runner) guard(ctx context.Context, dir, ref, tool string, confirm Confirm) error {
	if err := noDash(ref); err != nil {
		return err
	}
	out, err := r.Run(ctx, dir, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return err
	}
	hash := strings.TrimSpace(out)
	question := fmt.Sprintf("The commit %s is on the remote. %s rewrites it. Do you allow this?", shortHash(hash), tool)
	return r.GuardRemote(ctx, dir, hash, tool, question, confirm)
}

// Amend changes HEAD. With a summary it builds a new message. Else it keeps the message.
func (r Runner) Amend(ctx context.Context, dir string, m Message, paths []string, confirm Confirm) (string, error) {
	args := []string{"--no-edit"}
	if m.Summary != "" {
		var err error
		if args, err = m.Args("git_amend"); err != nil {
			return "", err
		}
	}
	if err := r.guard(ctx, dir, "HEAD", "git_amend", confirm); err != nil {
		return "", err
	}
	if len(paths) > 0 {
		if _, err := r.Stage(ctx, dir, paths, false); err != nil {
			return "", err
		}
	}
	if _, err := r.Run(ctx, dir, append([]string{"commit", "--amend"}, args...)...); err != nil {
		return "", err
	}
	return r.headIs(ctx, dir)
}

// Undo removes the last n commits and keeps their changes staged (ADR-003).
func (r Runner) Undo(ctx context.Context, dir string, n int, confirm Confirm) (string, error) {
	if n == 0 {
		n = 1
	}
	if n < 1 {
		return "", errors.New("git_undo refused: n must be 1 or more")
	}
	if err := r.guard(ctx, dir, "HEAD~"+strconv.Itoa(n-1), "git_undo", confirm); err != nil {
		return "", err
	}
	if _, err := r.Run(ctx, dir, "reset", "--soft", "HEAD~"+strconv.Itoa(n)); err != nil {
		return "", err
	}
	return r.headIs(ctx, dir)
}

// Fixup makes a fixup commit for commit from the staged changes.
func (r Runner) Fixup(ctx context.Context, dir, commit string, confirm Confirm) (string, error) {
	if err := r.guard(ctx, dir, commit, "git_fixup", confirm); err != nil {
		return "", err
	}
	if err := r.requireStaged(ctx, dir, "git_fixup"); err != nil {
		return "", err
	}
	if _, err := r.Run(ctx, dir, "commit", "--fixup="+commit); err != nil {
		return "", err
	}
	return r.headIs(ctx, dir)
}

// Revert makes a commit that undoes commit. With noCommit it only changes the index.
func (r Runner) Revert(ctx context.Context, dir, commit string, noCommit bool) (string, error) {
	if err := noDash(commit); err != nil {
		return "", err
	}
	args := []string{"revert", "--no-edit"}
	if noCommit {
		args = []string{"revert", "--no-commit"}
	}
	if _, err := r.Run(ctx, dir, append(args, commit)...); err != nil {
		return "", err
	}
	if noCommit {
		return "Reverted " + commit + " in the index.", nil
	}
	return r.headIs(ctx, dir)
}

// CherryPick applies commits on HEAD.
func (r Runner) CherryPick(ctx context.Context, dir string, commits []string) (string, error) {
	if len(commits) == 0 {
		return "", errors.New("git_cherry_pick refused: give commits")
	}
	for _, c := range commits {
		if err := noDash(c); err != nil {
			return "", err
		}
	}
	if _, err := r.Run(ctx, dir, append([]string{"cherry-pick"}, commits...)...); err != nil {
		return "", err
	}
	return r.headIs(ctx, dir)
}

// TagOptions are the options of Tag.
type TagOptions struct {
	Action  string
	Name    string
	Ref     string
	Message string
}

// Tag lists, creates or deletes tags. An empty action lists. A message makes an annotated tag.
func (r Runner) Tag(ctx context.Context, dir string, o TagOptions) (string, error) {
	if o.Action == "" || o.Action == "list" {
		out, err := r.Run(ctx, dir, "tag", "--list")
		if strings.TrimSpace(out) == "" {
			return "No tags.", err
		}
		return out, err
	}
	if err := errors.Join(noDash(o.Name), noDash(o.Ref)); err != nil || o.Name == "" {
		return "", errors.New("git_tag refused: give a name that does not start with a dash")
	}
	switch o.Action {
	case "create":
		return r.tagCreate(ctx, dir, o)
	case "delete":
		_, err := r.Run(ctx, dir, "tag", "-d", o.Name)
		return "Deleted tag " + o.Name + ".", err
	}
	return "", fmt.Errorf("git_tag refused: action %q is not list, create or delete", o.Action)
}

func (r Runner) tagCreate(ctx context.Context, dir string, o TagOptions) (string, error) {
	args := []string{"tag"}
	if o.Message != "" {
		args = append(args, "-a", "-m", o.Message)
	}
	args = append(args, o.Name)
	if o.Ref != "" {
		args = append(args, o.Ref)
	}
	_, err := r.Run(ctx, dir, args...)
	return "Created tag " + o.Name + ".", err
}

// StashOptions are the options of Stash.
type StashOptions struct {
	Action           string
	Message          string
	Index            int
	IncludeUntracked bool
}

// Stash lists, pushes, pops, applies or drops stash entries. An empty action lists.
func (r Runner) Stash(ctx context.Context, dir string, o StashOptions) (string, error) {
	var args []string
	switch o.Action {
	case "", "list":
		out, err := r.Run(ctx, dir, "stash", "list")
		if strings.TrimSpace(out) == "" {
			return "No stash entries.", err
		}
		return out, err
	case "push":
		args = []string{"stash", "push"}
		if o.IncludeUntracked {
			args = append(args, "--include-untracked")
		}
		if o.Message != "" {
			args = append(args, "-m", o.Message)
		}
	case "pop", "apply", "drop":
		args = []string{"stash", o.Action, "stash@{" + strconv.Itoa(max(o.Index, 0)) + "}"}
	default:
		return "", fmt.Errorf("git_stash refused: action %q is not list, push, pop, apply or drop", o.Action)
	}
	out, err := r.Run(ctx, dir, args...)
	return strings.TrimSpace(out), err
}
