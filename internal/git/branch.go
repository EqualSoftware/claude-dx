package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// BranchOptions are the options of Branch.
type BranchOptions struct {
	Action  string
	Name    string
	NewName string
	Start   string
}

// currentBranch returns the name of the current branch. It refuses a detached HEAD.
func (r Runner) currentBranch(ctx context.Context, dir, tool string) (string, error) {
	out, err := r.Run(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(out)
	if name == "" {
		return "", errors.New(tool + " refused: HEAD is detached")
	}
	return name, nil
}

// Branch lists, creates, switches, renames or deletes branches (git-tools#R-8). An empty action lists.
func (r Runner) Branch(ctx context.Context, dir string, o BranchOptions, confirm Confirm) (string, error) {
	if o.Action == "" || o.Action == "list" {
		return r.branchList(ctx, dir)
	}
	if err := errors.Join(noDash(o.Name), noDash(o.NewName), noDash(o.Start)); err != nil || o.Name == "" {
		return "", errors.New("git_branch refused: give a name that does not start with a dash")
	}
	switch o.Action {
	case "create":
		return r.branchDo(ctx, dir, "Created branch "+o.Name+".", "branch", o.Name, o.Start)
	case "switch":
		return r.branchDo(ctx, dir, "Switched to branch "+o.Name+".", "switch", o.Name)
	case "rename":
		if o.NewName == "" {
			return "", errors.New("git_branch refused: give new_name")
		}
		return r.branchDo(ctx, dir, "Renamed branch "+o.Name+" to "+o.NewName+".", "branch", "-m", o.Name, o.NewName)
	case "delete":
		return r.branchDelete(ctx, dir, o.Name, confirm)
	}
	return "", fmt.Errorf("git_branch refused: action %q is not list, create, switch, rename or delete", o.Action)
}

func (r Runner) branchDo(ctx context.Context, dir, done string, args ...string) (string, error) {
	if args[len(args)-1] == "" {
		args = args[:len(args)-1]
	}
	if _, err := r.Run(ctx, dir, args...); err != nil {
		return "", err
	}
	return done, nil
}

func (r Runner) branchList(ctx context.Context, dir string) (string, error) {
	out, err := r.Run(ctx, dir, "branch", "--list", "--format=%(HEAD) %(refname:short) %(objectname:short) %(contents:subject)")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "No branches.", nil
	}
	return out, nil
}

// branchDelete runs git branch -d. A branch that is not merged needs the user (git-tools#R-8).
func (r Runner) branchDelete(ctx context.Context, dir, name string, confirm Confirm) (string, error) {
	current, err := r.Run(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(current) == name {
		return "", errors.New("git_branch refused: " + name + " is the current branch")
	}
	_, err = r.Run(ctx, dir, "branch", "-d", name)
	if err == nil {
		return "Deleted branch " + name + ".", nil
	}
	if !strings.Contains(err.Error(), "not fully merged") {
		return "", err
	}
	question := "The branch " + name + " is not merged. git_branch deletes it with git branch -D. Do you allow this?"
	if err := askUser(ctx, "git_branch", "the branch "+name+" is not merged", question, confirm); err != nil {
		return "", err
	}
	if _, err := r.Run(ctx, dir, "branch", "-D", name); err != nil {
		return "", err
	}
	return "Deleted branch " + name + ".", nil
}

// askUser asks the question. It returns a refusal error with the reason when the user cannot be asked or says no.
func askUser(ctx context.Context, tool, reason, question string, confirm Confirm) error {
	if confirm == nil {
		return errors.New(tool + " refused: " + reason + " and the client cannot ask the user")
	}
	ok, err := confirm(ctx, question)
	if err != nil {
		return errors.New(tool + " refused: " + reason + " and the user was not asked: " + err.Error())
	}
	if !ok {
		return errors.New(tool + " refused: " + reason + " and the user said no")
	}
	return nil
}
