package git

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/result"
)

type pathsInput struct {
	pathInput
	Paths []string `json:"paths,omitempty" jsonschema:"the paths, relative to the directory of the tool"`
	All   bool     `json:"all,omitempty" jsonschema:"true to stage all changes, also untracked files. Use it with no paths"`
}

type messageInput struct {
	pathInput
	Type      string   `json:"type,omitempty" jsonschema:"feat, fix, refactor, perf, test, docs, build, ci, style, chore or revert"`
	Scope     string   `json:"scope,omitempty" jsonschema:"the scope: lowercase letters, digits and dashes. Optional"`
	Summary   string   `json:"summary,omitempty" jsonschema:"imperative, lowercase, no period. The header has at most 72 characters"`
	Body      string   `json:"body,omitempty" jsonschema:"the body of the message. Optional"`
	Breaking  bool     `json:"breaking,omitempty" jsonschema:"true for an incompatible change. It adds ! and a BREAKING CHANGE footer"`
	Paths     []string `json:"paths,omitempty" jsonschema:"paths to stage first. Optional"`
	CoAuthors []string `json:"co_authors,omitempty" jsonschema:"co-authors as Name <email>. Each adds a Co-Authored-By trailer"`
}

func (m messageInput) message() Message {
	return Message{Type: m.Type, Scope: m.Scope, Summary: m.Summary, Body: m.Body, Breaking: m.Breaking, CoAuthors: m.CoAuthors}
}

type undoInput struct {
	pathInput
	N int `json:"n,omitempty" jsonschema:"the number of commits to undo. The default is 1"`
}

type commitRefInput struct {
	pathInput
	Commit string `json:"commit" jsonschema:"the commit"`
}

type revertInput struct {
	pathInput
	Commit   string `json:"commit" jsonschema:"the commit to revert"`
	NoCommit bool   `json:"no_commit,omitempty" jsonschema:"true to change the index and not make a commit"`
}

type cherryPickInput struct {
	pathInput
	Commits []string `json:"commits" jsonschema:"the commits to apply, in order"`
}

type tagInput struct {
	pathInput
	Action  string `json:"action,omitempty" jsonschema:"list, create or delete. The default is list"`
	Name    string `json:"name,omitempty" jsonschema:"the tag name, for create and delete"`
	Ref     string `json:"ref,omitempty" jsonschema:"the commit to tag, for create. The default is HEAD"`
	Message string `json:"message,omitempty" jsonschema:"for create: a message makes an annotated tag"`
}

type stashInput struct {
	pathInput
	Action           string `json:"action,omitempty" jsonschema:"list, push, pop, apply or drop. The default is list"`
	Message          string `json:"message,omitempty" jsonschema:"for push: the stash message"`
	Index            int    `json:"index,omitempty" jsonschema:"for pop, apply and drop: the entry number. The default is 0"`
	IncludeUntracked bool   `json:"include_untracked,omitempty" jsonschema:"for push: also stash untracked files"`
}

// registerLocal adds the local tools that change the repository (git-tools#R-4).
func registerLocal(server *mcp.Server, r Runner) {
	addLocal(server, r, "git_stage", "Stage paths, or all changes with all.", false,
		func(ctx context.Context, dir string, in pathsInput, _ Confirm) (string, error) {
			return r.Stage(ctx, dir, in.Paths, in.All)
		})
	addLocal(server, r, "git_unstage", "Unstage paths. The changes stay in the working tree.", false,
		func(ctx context.Context, dir string, in pathsInput, _ Confirm) (string, error) {
			return r.Unstage(ctx, dir, in.Paths)
		})
	addLocal(server, r, "git_discard", "DESTRUCTIVE: drop the changes of tracked paths in the working tree. The changes are lost and cannot be restored. It does not touch untracked files.", true,
		func(ctx context.Context, dir string, in pathsInput, _ Confirm) (string, error) {
			return r.Discard(ctx, dir, in.Paths)
		})
	addLocal(server, r, "git_commit", "Commit the staged changes with a Conventional Commit message built from type, scope, summary, body, breaking and co_authors. It refuses an invalid message and a commit with nothing staged. With paths it stages them first.", false,
		func(ctx context.Context, dir string, in messageInput, _ Confirm) (string, error) {
			return r.Commit(ctx, dir, in.message(), in.Paths)
		})
	addLocal(server, r, "git_amend", "Change HEAD with the staged changes. With a summary it builds a new message as git_commit does. With no summary it keeps the message. If HEAD is on the remote it asks the user first.", true,
		func(ctx context.Context, dir string, in messageInput, confirm Confirm) (string, error) {
			return r.Amend(ctx, dir, in.message(), in.Paths, confirm)
		})
	addLocal(server, r, "git_undo", "Remove the last n commits and keep their changes staged (reset --soft). If a commit is on the remote it asks the user first. Use git_revert instead.", true,
		func(ctx context.Context, dir string, in undoInput, confirm Confirm) (string, error) {
			return r.Undo(ctx, dir, in.N, confirm)
		})
	addLocal(server, r, "git_fixup", "Make a fixup commit for a commit from the staged changes. If the commit is on the remote it asks the user first.", true,
		func(ctx context.Context, dir string, in commitRefInput, confirm Confirm) (string, error) {
			return r.Fixup(ctx, dir, in.Commit, confirm)
		})
	addLocal(server, r, "git_revert", "Make a commit that reverts a commit.", false,
		func(ctx context.Context, dir string, in revertInput, _ Confirm) (string, error) {
			return r.Revert(ctx, dir, in.Commit, in.NoCommit)
		})
	addLocal(server, r, "git_cherry_pick", "Apply commits on HEAD, in order.", false,
		func(ctx context.Context, dir string, in cherryPickInput, _ Confirm) (string, error) {
			return r.CherryPick(ctx, dir, in.Commits)
		})
	addLocal(server, r, "git_tag", "List, create or delete tags.", false,
		func(ctx context.Context, dir string, in tagInput, _ Confirm) (string, error) {
			return r.Tag(ctx, dir, TagOptions{Action: in.Action, Name: in.Name, Ref: in.Ref, Message: in.Message})
		})
	addLocal(server, r, "git_stash", "List, push, pop, apply or drop stash entries.", false,
		func(ctx context.Context, dir string, in stashInput, _ Confirm) (string, error) {
			return r.Stash(ctx, dir, StashOptions{Action: in.Action, Message: in.Message, Index: in.Index, IncludeUntracked: in.IncludeUntracked})
		})
}

func addLocal[In pather](server *mcp.Server, r Runner, name, description string, destructive bool,
	fn func(ctx context.Context, dir string, in In, confirm Confirm) (string, error)) {
	tool := &mcp.Tool{Name: name, Description: description, Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive}}
	mcp.AddTool(server, tool, func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		dir, err := r.Dir(ctx, in.dirPath())
		if err != nil {
			return result.Reply("", err, errNormal)
		}
		text, err := fn(ctx, dir, in, elicit(req))
		return result.Reply(text, err, errNormal)
	})
}

// allowSchema is the form of the question: one boolean.
var allowSchema = map[string]any{
	"type":       "object",
	"properties": map[string]any{"allow": map[string]any{"type": "boolean", "title": "Allow"}},
	"required":   []string{"allow"},
}

// elicit returns the Confirm of a request. Decline and cancel are no (git-tools#R-7).
func elicit(req *mcp.CallToolRequest) Confirm {
	if req == nil || req.Session == nil {
		return nil
	}
	return func(ctx context.Context, question string) (bool, error) {
		res, err := req.Session.Elicit(ctx, &mcp.ElicitParams{Message: question, RequestedSchema: allowSchema})
		if err != nil {
			return false, err
		}
		allow, _ := res.Content["allow"].(bool)
		return res.Action == "accept" && allow, nil
	}
}
