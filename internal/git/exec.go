package git

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/cache"
	"github.com/EqualSoftware/claude-dx/internal/report"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

type execInput struct {
	pathInput
	Args   []string `json:"args" jsonschema:"the git arguments, without the word git. Example: [\"reflog\", \"-n\", \"5\"]"`
	Reason string   `json:"reason" jsonschema:"why no other git_* tool fits this operation"`
}

// networkSubs are the git subcommands that use the network (git-tools#R-13).
var networkSubs = []string{"fetch", "pull", "push", "clone", "ls-remote"}

// notice tells the model what to say to the user about an operation that no tool maps.
const notice = "\n\nThis git operation is not mapped to a tool. Tell the user, and tell them to run `claude-dx report` to open an issue: " + report.IssueURL

// Exec logs the call in the file log (git-tools#R-16) and runs git with in.Args. The log is best effort:
// a log that cannot be written does not stop the command.
func (r Runner) Exec(ctx context.Context, dir, log string, in execInput) (string, error) {
	if len(in.Args) == 0 || in.Reason == "" {
		return "", errors.New("args and reason are required")
	}
	sub := in.Args[0]
	_ = report.Append(log, report.Entry{Time: time.Now().UTC(), Sub: sub, Args: in.Args, Reason: in.Reason})
	run := r.Run
	if slices.Contains(networkSubs, sub) || (sub == "remote" && len(in.Args) > 1 && in.Args[1] == "update") {
		run = r.RunNetwork
	}
	out, err := run(ctx, dir, in.Args...)
	if err != nil {
		return "", err
	}
	return out + notice, nil
}

// registerExec adds the tool git_exec. It is not read only.
func registerExec(server *mcp.Server, r Runner) {
	mcp.AddTool(server, &mcp.Tool{Name: "git_exec", Description: "Run git with the given arguments for an operation that no other git_* tool covers. Use another git_* tool when one fits. Each call is logged, and the result tells you what to tell the user."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in execInput) (*mcp.CallToolResult, any, error) {
			dir, err := r.Dir(ctx, in.Path)
			if err != nil {
				return result.Reply("", err, errNormal)
			}
			log := ""
			if c, err := cache.Dir(); err == nil {
				log = filepath.Join(c, report.LogName)
			}
			text, err := r.Exec(ctx, dir, log, in)
			return result.Reply(text, err, errNormal)
		})
}
