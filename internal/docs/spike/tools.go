package spike

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/result"
)

type createInput struct {
	Slug     string `json:"slug" jsonschema:"file name without .md, in kebab-case, at most 40 characters, example lock-on-nfs"`
	Title    string `json:"title" jsonschema:"short title, one line"`
	Question string `json:"question" jsonschema:"the one question that the exploration answers"`
}

type listInput struct {
	Status string `json:"status,omitempty" jsonschema:"only spikes with this status: Open or Closed"`
}

type getInput struct {
	Slug    string `json:"slug" jsonschema:"spike slug in kebab-case, example lock-on-nfs"`
	Section string `json:"section,omitempty" jsonschema:"one section only: Question, Notes or Result"`
}

type noteInput struct {
	Slug string `json:"slug" jsonschema:"slug of an Open spike"`
	Text string `json:"text" jsonschema:"one finding, on one line. The tool adds the date"`
}

type closeInput struct {
	Slug   string `json:"slug" jsonschema:"slug of an Open spike"`
	Result string `json:"result" jsonschema:"the answer to the question. A Closed spike does not change"`
}

type deleteInput struct {
	Slugs []string `json:"slugs" jsonschema:"slugs of the spikes to delete. The tool deletes none if one does not exist"`
}

// Register adds the spike tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "spike_create", "Create an Open spike: one question to explore, before a spec exists. The tool refuses a slug that exists.", nil,
		func(in createInput) (string, error) {
			return "Created spike " + in.Slug + ".", s.Create(in.Slug, in.Title, in.Question)
		})
	add(server, "spike_list", "List the spikes: slug, status, created date and title. Open spikes come first.", ErrNoDir,
		func(in listInput) (string, error) { return s.listText(in.Status) })
	add(server, "spike_get", "Read one spike, or one section of it.", ErrNoDir,
		func(in getInput) (string, error) { return s.Get(in.Slug, in.Section) })
	add(server, "spike_note", "Add one finding, with today's date, to the Notes of an Open spike.", nil,
		func(in noteInput) (string, error) {
			return "Added a note to spike " + in.Slug + ".", s.Note(in.Slug, in.Text)
		})
	add(server, "spike_close", "Write the Result of an Open spike and close it. A Closed spike is final.", nil,
		func(in closeInput) (string, error) {
			return "Closed spike " + in.Slug + ".", s.Close(in.Slug, in.Result)
		})
	add(server, "spike_delete", "Delete one or more spikes of any status. The tool deletes none if one slug does not exist.", nil,
		func(in deleteInput) (string, error) {
			return "Deleted spikes " + strings.Join(in.Slugs, ", ") + ".", s.Delete(in.Slugs)
		})
}

// errNone is an error that never happens: it is the normal error of a tool that has none.
var errNone = errors.New("none")

// add registers a tool. normal is an error that is a normal state, not a failure; nil means none.
func add[In any](server *mcp.Server, name, description string, normal error, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			if normal == nil {
				return result.Reply(text, err, errNone)
			}
			return result.Reply(text, err, normal)
		})
}

func (s Store) listText(status string) (string, error) {
	list, warnings, err := s.List(status)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return strings.TrimSpace("No spikes.\n" + warnings), nil
	}
	rows := make([][]string, 0, len(list))
	for _, sp := range list {
		rows = append(rows, []string{sp.Slug, sp.Status, sp.Created, sp.Title})
	}
	return result.Table(rows) + warnings, nil
}
