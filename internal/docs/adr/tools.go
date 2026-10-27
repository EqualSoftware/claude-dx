package adr

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

type getInput struct {
	ID      string `json:"id" jsonschema:"ADR ID, example ADR-007"`
	Section string `json:"section,omitempty" jsonschema:"one section only: Context, Decision, Alternatives, Consequences or Evidence"`
}

type createInput struct {
	Title        string        `json:"title" jsonschema:"short decision title, in ASD-STE100, 40 characters or less, not the title of another ADR"`
	Context      string        `json:"context" jsonschema:"problem, constraints and why a decision is necessary"`
	Decision     string        `json:"decision" jsonschema:"the selected approach"`
	Alternatives []Alternative `json:"alternatives" jsonschema:"the other options and why they are rejected"`
	Consequences Consequences  `json:"consequences"`
	Evidence     string        `json:"evidence" jsonschema:"verified facts only: path:line, commits, benchmarks, issues"`
	Status       string        `json:"status,omitempty" jsonschema:"Proposed (default) or Accepted. Accepted only when the user approved the decision"`
	Owners       string        `json:"owners,omitempty" jsonschema:"a role or name, default: the git user.name, else Maintainer"`
}

type updateInput struct {
	ID      string `json:"id" jsonschema:"ADR ID, example ADR-007"`
	Status  string `json:"status,omitempty" jsonschema:"Proposed or Accepted"`
	Section string `json:"section,omitempty" jsonschema:"section to replace: Context, Decision, Alternatives, Consequences or Evidence"`
	Body    string `json:"body,omitempty" jsonschema:"new content of the section, with no heading"`
}

type supersedeInput struct {
	OldID string `json:"old_id" jsonschema:"the ADR that is replaced"`
	NewID string `json:"new_id" jsonschema:"the ADR that replaces it"`
}

// Register adds the ADR tools to the server.
func Register(server *mcp.Server, s Store) {
	mcp.AddTool(server, &mcp.Tool{Name: "adr_list", Description: "List the ADRs in docs/decisions: number, status and title."},
		func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			var text string
			err := docfile.RLocked(s.Lock, func() (err error) { text, err = s.listText(); return err })
			return reply(text, err)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "adr_get", Description: "Read one ADR, or one section of it. Read one section when you need only that part."},
		func(_ context.Context, _ *mcp.CallToolRequest, in getInput) (*mcp.CallToolResult, any, error) {
			var text string
			err := docfile.RLocked(s.Lock, func() (err error) { text, err = s.Get(in.ID, in.Section); return err })
			return reply(text, err)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "adr_create", Description: "Record a new ADR in docs/decisions. The tool sets the number, the date and the file name."},
		func(_ context.Context, _ *mcp.CallToolRequest, in createInput) (*mcp.CallToolResult, any, error) {
			return reply(s.createText(in))
		})
	mcp.AddTool(server, &mcp.Tool{Name: "adr_update", Description: "Change the status of an ADR, or replace one section, or both."},
		func(_ context.Context, _ *mcp.CallToolRequest, in updateInput) (*mcp.CallToolResult, any, error) {
			return reply(fmt.Sprintf("Updated ADR %s.", in.ID), s.Update(in.ID, in.Status, in.Section, in.Body))
		})
	mcp.AddTool(server, &mcp.Tool{Name: "adr_supersede", Description: "Mark an ADR as replaced by a newer ADR. The tool changes both ADRs."},
		func(_ context.Context, _ *mcp.CallToolRequest, in supersedeInput) (*mcp.CallToolResult, any, error) {
			return reply(fmt.Sprintf("%s is superseded by %s.", in.OldID, in.NewID), s.Supersede(in.OldID, in.NewID))
		})
}

// reply makes the tool result. A missing docs/decisions is a normal state, so it is not an error.
func reply(text string, err error) (*mcp.CallToolResult, any, error) {
	return result.Reply(text, err, ErrNoDir)
}

func (s Store) listText() (string, error) {
	list, warnings, err := s.List()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return strings.TrimSpace("No ADRs in docs/decisions.\n" + warnings), nil
	}
	rows := make([][]string, 0, len(list))
	for _, e := range list {
		rows = append(rows, []string{kind.ID(e.Number), e.Status, e.Title})
	}
	return result.Table(rows) + warnings, nil
}

func (s Store) createText(in createInput) (string, error) {
	path, err := s.Create(New{
		Title: in.Title, Status: in.Status, Owners: in.Owners, Context: in.Context, Decision: in.Decision,
		Alternatives: in.Alternatives, Consequences: in.Consequences, Evidence: in.Evidence,
	})
	if err != nil {
		return "", err
	}
	status := in.Status
	if status == "" {
		status = "Proposed"
	}
	return fmt.Sprintf("Created docs/decisions/%s (%s).", filepath.Base(path), status), nil
}
