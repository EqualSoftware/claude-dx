package git

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type branchInput struct {
	pathInput
	Action  string `json:"action,omitempty" jsonschema:"list, create, switch, rename or delete. The default is list"`
	Name    string `json:"name,omitempty" jsonschema:"the branch, for create, switch, rename and delete"`
	NewName string `json:"new_name,omitempty" jsonschema:"for rename: the new name"`
	Start   string `json:"start,omitempty" jsonschema:"for create: the start point. The default is HEAD"`
}

type rebaseInput struct {
	pathInput
	Action string `json:"action" jsonschema:"start, continue, abort or skip"`
	Onto   string `json:"onto,omitempty" jsonschema:"for start: the branch or commit to rebase onto"`
}

type pushInput struct {
	pathInput
	Remote string `json:"remote,omitempty" jsonschema:"the remote. The default is origin"`
	Branch string `json:"branch,omitempty" jsonschema:"the branch to push. The default is the current branch"`
}

type worktreeInput struct {
	pathInput
	Action       string `json:"action,omitempty" jsonschema:"list, add or remove. The default is list"`
	Worktree     string `json:"worktree,omitempty" jsonschema:"for add and remove: the directory of the worktree"`
	Branch       string `json:"branch,omitempty" jsonschema:"for add: the name of the new branch"`
	Start        string `json:"start,omitempty" jsonschema:"for add: the start point of the new branch. The default is HEAD"`
	DeleteBranch bool   `json:"delete_branch,omitempty" jsonschema:"for remove: also delete the branch of the worktree with git branch -d"`
}

// registerBranch adds the branch, remote and worktree tools (git-tools#R-8 to R-12, R-22).
func registerBranch(server *mcp.Server, r Runner) {
	addLocal(server, r, "git_branch", "List, create, switch, rename or delete branches. delete uses git branch -d and refuses the current branch. If the branch is not merged it asks the user first.", true,
		func(ctx context.Context, dir string, in branchInput, confirm Confirm) (string, error) {
			return r.Branch(ctx, dir, BranchOptions{Action: in.Action, Name: in.Name, NewName: in.NewName, Start: in.Start}, confirm)
		})
	addLocal(server, r, "git_sync", "Fetch with prune and rebase the current branch onto its upstream. It keeps the local changes. On a conflict it returns the conflicts, as git_conflicts does.", false,
		func(ctx context.Context, dir string, _ pathInput, _ Confirm) (string, error) {
			return r.Sync(ctx, dir)
		})
	addLocal(server, r, "git_rebase", "Start a rebase onto a ref, or continue, abort or skip a rebase in progress. start asks the user first if it rewrites a commit that is on the remote.", true,
		func(ctx context.Context, dir string, in rebaseInput, confirm Confirm) (string, error) {
			return r.Rebase(ctx, dir, RebaseOptions{Action: in.Action, Onto: in.Onto}, confirm)
		})
	add(server, r, "git_conflicts", "Show each conflicted path with its conflict hunks and line numbers. Reads only.",
		func(ctx context.Context, dir string, _ pathInput) (string, error) { return r.Conflicts(ctx, dir) })
	addLocal(server, r, "git_push", "Push a branch. It sets the upstream on the first push. If the remote branch is not an ancestor it uses --force-with-lease and --force-if-includes, and for the main branch it asks the user first.", true,
		func(ctx context.Context, dir string, in pushInput, confirm Confirm) (string, error) {
			return r.Push(ctx, dir, PushOptions{Remote: in.Remote, Branch: in.Branch}, confirm)
		})
	addLocal(server, r, "git_worktree", "List, add or remove worktrees. add makes a new branch from a start point. remove refuses a worktree with changes that are not committed.", true,
		func(ctx context.Context, dir string, in worktreeInput, _ Confirm) (string, error) {
			return r.Worktree(ctx, dir, WorktreeOptions{Action: in.Action, Worktree: in.Worktree, Branch: in.Branch, Start: in.Start, DeleteBranch: in.DeleteBranch})
		})
}
