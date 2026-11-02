package git

import (
	"context"
	"errors"
	"strings"
)

// PushOptions are the options of Push.
type PushOptions struct {
	Remote string
	Branch string
}

// mainBranch returns the main branch of a remote: the branch of <remote>/HEAD, or main or master. It returns
// an empty string when it finds none.
func (r Runner) mainBranch(ctx context.Context, dir, remote string) string {
	if out, err := r.Run(ctx, dir, "symbolic-ref", "--short", "refs/remotes/"+remote+"/HEAD"); err == nil {
		return strings.TrimPrefix(strings.TrimSpace(out), remote+"/")
	}
	for _, name := range []string{"main", "master"} {
		if _, err := r.Run(ctx, dir, "show-ref", "--verify", "--quiet", "refs/remotes/"+remote+"/"+name); err == nil {
			return name
		}
	}
	return ""
}

// Push pushes a branch (git-tools#R-12). If the remote branch is not an ancestor of HEAD, it uses
// --force-with-lease and --force-if-includes, and a push to the main branch needs the user. If the branch has
// no upstream, it sets the upstream.
func (r Runner) Push(ctx context.Context, dir string, o PushOptions, confirm Confirm) (string, error) {
	if o.Remote == "" {
		o.Remote = "origin"
	}
	if err := errors.Join(noDash(o.Remote), noDash(o.Branch)); err != nil {
		return "", errors.New("git_push refused: the remote and the branch must not start with a dash")
	}
	if o.Branch == "" {
		branch, err := r.currentBranch(ctx, dir, "git_push")
		if err != nil {
			return "", err
		}
		o.Branch = branch
	}
	force, err := r.needsForce(ctx, dir, o)
	if err != nil {
		return "", err
	}
	args := []string{"push"}
	done := "Pushed " + o.Branch + " to " + o.Remote + "."
	if force {
		if err := r.guardMain(ctx, dir, o, confirm); err != nil {
			return "", err
		}
		args = append(args, "--force-with-lease", "--force-if-includes")
		done = "Pushed " + o.Branch + " to " + o.Remote + " with force-with-lease."
	}
	if _, err := r.Run(ctx, dir, "config", "--get", "branch."+o.Branch+".merge"); err != nil {
		args = append(args, "--set-upstream")
		done += " Set the upstream to " + o.Remote + "/" + o.Branch + "."
	}
	if _, err := r.RunNetwork(ctx, dir, append(args, o.Remote, "refs/heads/"+o.Branch)...); err != nil {
		return "", err
	}
	return done, nil
}

// needsForce fetches the remote branch and reports whether it is not an ancestor of the branch. A branch that
// is not on the remote needs no force.
func (r Runner) needsForce(ctx context.Context, dir string, o PushOptions) (bool, error) {
	tracking := "refs/remotes/" + o.Remote + "/" + o.Branch
	_, err := r.RunNetwork(ctx, dir, "fetch", o.Remote, "+refs/heads/"+o.Branch+":"+tracking)
	if err != nil && strings.Contains(err.Error(), "couldn't find remote ref") {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = r.Run(ctx, dir, "merge-base", "--is-ancestor", tracking, "refs/heads/"+o.Branch)
	return err != nil, nil
}

// guardMain asks the user before a force push to the main branch (git-tools#R-12).
func (r Runner) guardMain(ctx context.Context, dir string, o PushOptions, confirm Confirm) error {
	if o.Branch != r.mainBranch(ctx, dir, o.Remote) {
		return nil
	}
	question := "The branch " + o.Branch + " is the main branch. git_push must force push it with --force-with-lease. Do you allow this?"
	return askUser(ctx, "git_push", o.Branch+" is the main branch", question, confirm)
}
