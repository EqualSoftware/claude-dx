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

// registerBranch adds the branch, remote and worktree tools (git-tools#R-8 to R-12, R-22).
func registerBranch(server *mcp.Server, r Runner) {
	addLocal(server, r, "git_branch", "List, create, switch, rename or delete branches. delete uses git branch -d and refuses the current branch. If the branch is not merged it asks the user first.", true,
		func(ctx context.Context, dir string, in branchInput, confirm Confirm) (string, error) {
			return r.Branch(ctx, dir, BranchOptions{Action: in.Action, Name: in.Name, NewName: in.NewName, Start: in.Start}, confirm)
		})
}
