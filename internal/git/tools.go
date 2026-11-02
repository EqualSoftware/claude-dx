package git

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/result"
)

// errNormal is the "normal" error of result.Reply. No git error is normal, so it never matches.
var errNormal = errors.New("normal")

type pathInput struct {
	Path string `json:"path,omitempty" jsonschema:"the project root or a worktree of the same repository. Omit it for the project root"`
}

type diffInput struct {
	pathInput
	Staged bool     `json:"staged,omitempty" jsonschema:"true to show the staged changes"`
	Ref    string   `json:"ref,omitempty" jsonschema:"a commit or range to compare with, example main..HEAD. Omit it for the working tree"`
	Paths  []string `json:"paths,omitempty" jsonschema:"only these paths"`
	Stat   bool     `json:"stat,omitempty" jsonschema:"true to show file stats and not the diff"`
}

type logInput struct {
	pathInput
	Ref   string   `json:"ref,omitempty" jsonschema:"the branch or commit where the log starts. Omit it for HEAD"`
	N     int      `json:"n,omitempty" jsonschema:"the maximum number of commits. The default is 20"`
	Paths []string `json:"paths,omitempty" jsonschema:"only the commits that change these paths"`
}

type showInput struct {
	pathInput
	Ref  string `json:"ref,omitempty" jsonschema:"the commit to show. The default is HEAD"`
	Stat bool   `json:"stat,omitempty" jsonschema:"true to show file stats and not the diff"`
}

type blameInput struct {
	pathInput
	File  string `json:"file" jsonschema:"the file to blame, relative to the directory of the tool"`
	Ref   string `json:"ref,omitempty" jsonschema:"the commit where blame starts. Omit it for the working tree"`
	Start int    `json:"start,omitempty" jsonschema:"the first line. Omit it with end for all lines"`
	End   int    `json:"end,omitempty" jsonschema:"the last line. Omit it for the end of the file"`
}

// Register adds the read tools to the server. root is the project root.
func Register(server *mcp.Server, root string) {
	r := Runner{Root: root}
	add(server, r, "git_status", "Show the git status: the branch line (## branch...upstream +ahead -behind), then the groups Staged, Unstaged, Conflicted and Untracked with one line for each path. Reads only.",
		func(ctx context.Context, dir string, _ pathInput) (string, error) { return r.Status(ctx, dir) })
	add(server, r, "git_diff", "Show a git diff with no color, or the file stats with stat. Reads only.",
		func(ctx context.Context, dir string, in diffInput) (string, error) {
			return r.Diff(ctx, dir, DiffOptions{Staged: in.Staged, Ref: in.Ref, Paths: in.Paths, Stat: in.Stat})
		})
	add(server, r, "git_log", "Show the git log, one line for each commit: <short-hash> <YYYY-MM-DD> <author> <subject>. Reads only.",
		func(ctx context.Context, dir string, in logInput) (string, error) {
			return r.Log(ctx, dir, LogOptions{Ref: in.Ref, N: in.N, Paths: in.Paths})
		})
	add(server, r, "git_show", "Show one commit: a header line (<short-hash> <YYYY-MM-DD> <author> <subject>) and its diff, or its file stats with stat. Reads only.",
		func(ctx context.Context, dir string, in showInput) (string, error) {
			return r.Show(ctx, dir, in.Ref, in.Stat)
		})
	add(server, r, "git_blame", "Show the author of each line of a file, one line for each source line: <short-hash> <author> <line-no>: <text>. Reads only.",
		func(ctx context.Context, dir string, in blameInput) (string, error) {
			return r.Blame(ctx, dir, BlameOptions{File: in.File, Ref: in.Ref, Start: in.Start, End: in.End})
		})
	registerExec(server, r)
}

// pather gives the path of an input.
type pather interface{ dirPath() string }

func (p pathInput) dirPath() string { return p.Path }

func add[In pather](server *mcp.Server, r Runner, name, description string,
	fn func(ctx context.Context, dir string, in In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			dir, err := r.Dir(ctx, in.dirPath())
			if err != nil {
				return result.Reply("", err, errNormal)
			}
			text, err := fn(ctx, dir, in)
			return result.Reply(text, err, errNormal)
		})
}
