package glossary

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/result"
)

type termInput struct {
	Term string `json:"term" jsonschema:"the term or a word to find. The tool compares the full text with no case difference"`
}

type listInput struct {
	Area string `json:"area,omitempty" jsonschema:"only the terms of this area, example process. Omit it for all terms"`
}

type upsertInput struct {
	Term       string   `json:"term" jsonschema:"the approved word"`
	Definition string   `json:"definition" jsonschema:"one or two sentences: what it is, not how to use it. In ASD-STE100"`
	Area       string   `json:"area" jsonschema:"project, process, build or the name of an area of the project"`
	DoNotUse   []string `json:"do_not_use,omitempty" jsonschema:"a list of synonyms to avoid, example [\"ticket\", \"issue\"]; empty list if none"`
	Source     string   `json:"source" jsonschema:"the file, ADR or task that defines or uses the term"`
}

// Register adds the glossary tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "glossary_lookup", "Find a term in docs/GLOSSARY.md, also in the Do not use column. Returns the row (| Term | Definition | Area | Do not use | Source |), the approved term, or No match.",
		func(in termInput) (string, error) {
			text, ok, err := s.Lookup(in.Term)
			if err == nil && !ok {
				return "No match.", nil
			}
			return text, err
		})
	add(server, "glossary_list", "List the terms of docs/GLOSSARY.md, one row for each term (| Term | Definition | Area | Do not use | Source |). Use it when you need many terms; use glossary_lookup for one term.",
		func(in listInput) (string, error) { return s.listText(in.Area) })
	add(server, "glossary_upsert", "Add a term to docs/GLOSSARY.md, or replace the row of the term. The tool keeps alphabetical order and refuses a conflict with another row.",
		func(in upsertInput) (string, error) {
			added, err := s.Upsert(Row(in))
			if added {
				return fmt.Sprintf("Added %q.", in.Term), err
			}
			return fmt.Sprintf("Changed %q.", in.Term), err
		})
	add(server, "glossary_delete", "Remove the row of a term from docs/GLOSSARY.md.",
		func(in termInput) (string, error) { return fmt.Sprintf("Deleted %q.", in.Term), s.Delete(in.Term) })
}

func add[In any](server *mcp.Server, name, description string, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			return result.Reply(text, err, ErrNoFile)
		})
}

func (s Store) listText(area string) (string, error) {
	rows, err := s.List(area)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 && area != "" {
		return fmt.Sprintf("No terms with the area %q.", area), nil
	}
	if len(rows) == 0 {
		return "No terms in docs/GLOSSARY.md.", nil
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(r.line())
		b.WriteByte('\n')
	}
	return b.String(), nil
}
