package fs

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/project"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

// errNormal is the "normal" error of result.Reply. No fs error is normal, so it never matches.
var errNormal = errors.New("normal")

type transferInput struct {
	Source      string `json:"source" jsonschema:"the path to move or copy from, relative to the project root"`
	Destination string `json:"destination" jsonschema:"the path to move or copy to, relative to the project root"`
	Overwrite   bool   `json:"overwrite,omitempty" jsonschema:"true to replace a destination that exists. The default is false"`
}

type deleteInput struct {
	Paths     []string `json:"paths" jsonschema:"the files and directories to delete"`
	Recursive bool     `json:"recursive,omitempty" jsonschema:"true to delete a directory that is not empty. The default is false"`
}

type mkdirInput struct {
	Paths []string `json:"paths" jsonschema:"the directories to make, with their parents"`
}

type treeInput struct {
	Path  string `json:"path,omitempty" jsonschema:"the directory to show. The default is the project root"`
	Depth int    `json:"depth,omitempty" jsonschema:"the maximum depth. The default is 3"`
}

// Register adds the fs_* tools to the server.
func Register(server *mcp.Server, p project.Project) {
	t := Tools{Project: p}
	no := false
	add(server, "fs_move", "Move a file or a directory inside the project. It refuses a destination that exists unless overwrite is true. It refuses the project root and .git.",
		&mcp.ToolAnnotations{IdempotentHint: false},
		func(_ context.Context, in transferInput) (string, error) {
			return t.Move(in.Source, in.Destination, in.Overwrite)
		})
	add(server, "fs_copy", "Copy a file, or a directory with all its contents, inside the project. It refuses a destination that exists unless overwrite is true. It refuses the project root and .git.",
		&mcp.ToolAnnotations{},
		func(_ context.Context, in transferInput) (string, error) {
			return t.Copy(in.Source, in.Destination, in.Overwrite)
		})
	add(server, "fs_delete", "Delete files and directories inside the project, with no prompt. A directory that is not empty needs recursive. It refuses the project root and .git. A symbolic link is deleted, not its target.",
		&mcp.ToolAnnotations{},
		func(_ context.Context, in deleteInput) (string, error) { return t.Delete(in.Paths, in.Recursive) })
	add(server, "fs_mkdir", "Make directories inside the project, with their parents. A directory that exists is not an error.",
		&mcp.ToolAnnotations{DestructiveHint: &no, IdempotentHint: true},
		func(_ context.Context, in mkdirInput) (string, error) { return t.Mkdir(in.Paths) })
	add(server, "fs_tree", "Show an indented tree of the files of the project (git ls-files: the files that .gitignore names are not shown). A directory has a trailing slash. Maximum 500 lines. Reads only.",
		&mcp.ToolAnnotations{ReadOnlyHint: true},
		func(ctx context.Context, in treeInput) (string, error) { return t.Tree(ctx, in.Path, in.Depth) })
}

func add[In any](server *mcp.Server, name, description string, ann *mcp.ToolAnnotations,
	fn func(ctx context.Context, in In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description, Annotations: ann},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(ctx, in)
			return result.Reply(text, err, errNormal)
		})
}
