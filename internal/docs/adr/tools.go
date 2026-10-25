package adr

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/is-equal/claude-dx/internal/result"
)

type getInput struct {
	ID      string `json:"id" jsonschema:"ADR ID, example ADR-007"`
	Section string `json:"section,omitempty" jsonschema:"one section only: Context, Decision, Alternatives, Consequences or Evidence"`
}

type createInput struct {
	Title        string        `json:"title" jsonschema:"short decision title, in ASD-STE100"`
	Context      string        `json:"context" jsonschema:"problem, constraints and why a decision is necessary"`
	Decision     string        `json:"decision" jsonschema:"the selected approach"`
	Alternatives []Alternative `json:"alternatives" jsonschema:"the other options and why they are rejected"`
	Consequences Consequences  `json:"consequences"`
	Evidence     string        `json:"evidence" jsonschema:"verified facts only: path:line, commits, benchmarks, issues"`
	Status       string        `json:"status,omitempty" jsonschema:"Proposed (default) or Accepted. Accepted only when the user approved the decision"`
	Owners       string        `json:"owners,omitempty" jsonschema:"a role, default Maintainer"`
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
			return reply(s.listText())
		})
	mcp.AddTool(server, &mcp.Tool{Name: "adr_get", Description: "Read one ADR, or one section of it. Read one section when you need only that part."},
		func(_ context.Context, _ *mcp.CallToolRequest, in getInput) (*mcp.CallToolResult, any, error) {
			return reply(s.Get(in.ID, in.Section))
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
	if errors.Is(err, ErrNoDir) {
		return result.Text(result.Sentence(err)), nil, nil
	}
	if err != nil {
		return result.Errorf("%s", result.Sentence(err)), nil, nil
	}
	return result.Text(text), nil, nil
}

func (s Store) listText() (string, error) {
	list, err := s.List()
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "No ADRs in docs/decisions.", nil
	}
	rows := make([][]string, 0, len(list))
	seen := map[int]int{}
	for _, e := range list {
		rows = append(rows, []string{fmt.Sprintf("ADR-%03d", e.Number), e.Status, e.Title})
		seen[e.Number]++
	}
	return result.Table(rows) + duplicates(seen), nil
}

func duplicates(seen map[int]int) string {
	var b strings.Builder
	for n, count := range seen {
		if count > 1 {
			fmt.Fprintf(&b, "Warning: ADR-%03d has %d files. Give each file a different number.\n", n, count)
		}
	}
	return b.String()
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
