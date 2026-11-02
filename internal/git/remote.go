package git

import (
	"context"
	"errors"
	"strings"
)

// Confirm asks the user a yes or no question (MCP elicitation, git-tools#R-7).
// It returns an error when the client cannot ask.
type Confirm func(ctx context.Context, question string) (bool, error)

// OnRemote reports whether a remote-tracking branch contains the commit (git-tools#R-6).
func (r Runner) OnRemote(ctx context.Context, dir, commit string) (bool, error) {
	out, err := r.Run(ctx, dir, "branch", "-r", "--contains", commit)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "->") {
			return true, nil
		}
	}
	return false, nil
}

// errRemoteRefused is the base of the refusal when a tool must not rewrite a commit on the remote.
var errRemoteRefused = errors.New("refused")

// GuardRemote lets a tool rewrite commit only when the commit is not on the remote, or the user
// agrees. It returns an error that says why the tool refuses, and the tool must not change the
// repository then (git-tools#R-7). A nil confirm means the client has no elicitation.
func (r Runner) GuardRemote(ctx context.Context, dir, commit, tool, question string, confirm Confirm) error {
	on, err := r.OnRemote(ctx, dir, commit)
	if err != nil || !on {
		return err
	}
	if confirm == nil {
		return errors.Join(errRemoteRefused, errors.New(tool+" refused: the commit is on the remote and the client cannot ask the user"))
	}
	ok, err := confirm(ctx, question)
	if err != nil {
		return errors.Join(errRemoteRefused, errors.New(tool+" refused: the commit is on the remote and the user was not asked: "+err.Error()))
	}
	if !ok {
		return errors.Join(errRemoteRefused, errors.New(tool+" refused: the commit is on the remote and the user said no"))
	}
	return nil
}
