package learning

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

type topicInput struct {
	Topic string `json:"topic" jsonschema:"file name without .md, in kebab-case, example go-embed-paths"`
}

type createInput struct {
	Topic       string `json:"topic" jsonschema:"file name without .md, in kebab-case, example go-embed-paths"`
	Description string `json:"description" jsonschema:"one line: what the lessons are about and when to use them"`
	Body        string `json:"body" jsonschema:"the lessons, in ASD-STE100. One lesson is one or two sentences"`
}

type updateInput struct {
	Topic       string `json:"topic" jsonschema:"file name without .md"`
	Description string `json:"description,omitempty" jsonschema:"new description, one line"`
	Body        string `json:"body,omitempty" jsonschema:"new body. It replaces all of the old body"`
}

// Register adds the learning tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "learning_list", "List the learnings in docs/learnings: topic and description. Run it before learning_create.",
		func(struct{}) (text string, err error) {
			err = docfile.RLocked(s.Lock, func() (err error) { text, err = s.listText(); return err })
			return text, err
		})
	add(server, "learning_get", "Read one learning file.",
		func(in topicInput) (text string, err error) {
			err = docfile.RLocked(s.Lock, func() (err error) { text, err = s.Get(in.Topic); return err })
			return text, err
		})
	add(server, "learning_create", "Create docs/learnings/<topic>.md. The tool refuses a topic that exists.",
		func(in createInput) (string, error) {
			return "Created docs/learnings/" + in.Topic + ".md.", s.Create(in.Topic, in.Description, in.Body)
		})
	add(server, "learning_update", "Change the description of a learning, or replace its body, or both. To add a lesson, read the file with learning_get first.",
		func(in updateInput) (string, error) {
			return "Updated learning " + in.Topic + ".", s.Update(in.Topic, in.Description, in.Body)
		})
	add(server, "learning_delete", "Delete one learning file.",
		func(in topicInput) (string, error) { return "Deleted learning " + in.Topic + ".", s.Delete(in.Topic) })
}

func add[In any](server *mcp.Server, name, description string, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			return result.Reply(text, err, ErrNoDir)
		})
}

func (s Store) listText() (string, error) {
	list, warnings, err := s.List()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return strings.TrimSpace("No learnings in docs/learnings.\n" + warnings), nil
	}
	var b strings.Builder
	for _, l := range list {
		fmt.Fprintf(&b, "%s: %s\n", l.Topic, l.Description)
	}
	return b.String() + warnings, nil
}
