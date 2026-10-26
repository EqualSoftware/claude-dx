package spec

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

type listInput struct {
	Status    string `json:"status,omitempty" jsonschema:"Draft, In Review or Approved"`
	Milestone string `json:"milestone,omitempty" jsonschema:"one milestone, example v0.1.0"`
	Priority  string `json:"priority,omitempty" jsonschema:"P0, P1, P2 or P3"`
	Area      string `json:"area,omitempty" jsonschema:"one area"`
	Related   string `json:"related,omitempty" jsonschema:"one ADR ID"`
}

type getInput struct {
	Slug    string `json:"slug" jsonschema:"spec slug, example docs-lock"`
	Section string `json:"section,omitempty" jsonschema:"one section only, by number (example 5) or title. Read one section when you need only that part"`
}

type createInput struct {
	Slug      string   `json:"slug" jsonschema:"kebab-case ID of the spec, 40 characters or less, example docs-lock. It does not change"`
	Title     string   `json:"title" jsonschema:"short spec title, in ASD-STE100, 40 characters or less, not the title of another spec"`
	Milestone string   `json:"milestone" jsonschema:"vMAJOR.MINOR.PATCH, example v0.1.0"`
	Priority  string   `json:"priority" jsonschema:"P0, P1, P2 or P3"`
	Area      []string `json:"area" jsonschema:"the areas of the project, example [api] or [docs, ci]"`
	Related   []string `json:"related,omitempty" jsonschema:"IDs of ADRs that exist, example [ADR-016]"`
}

type sectionInput struct {
	Slug    string `json:"slug" jsonschema:"spec slug"`
	Section string `json:"section" jsonschema:"section number (example 6) or title"`
	Body    string `json:"body" jsonschema:"new content of the section, with no heading"`
}

type statusInput struct {
	Slug   string `json:"slug" jsonschema:"spec slug"`
	Status string `json:"status" jsonschema:"Draft, In Review or Approved"`
	Note   string `json:"note,omitempty" jsonschema:"Document History note"`
}

type reqAddInput struct {
	Slug string `json:"slug" jsonschema:"spec slug"`
	Text string `json:"text" jsonschema:"one behavior that the spec requires, one line"`
}

type reqSetInput struct {
	Slug   string `json:"slug" jsonschema:"spec slug"`
	RID    string `json:"rid" jsonschema:"requirement ID, example R-2"`
	Text   string `json:"text,omitempty" jsonschema:"new text, one line. A text change of an Approved spec sets it to In Review"`
	Status string `json:"status,omitempty" jsonschema:"Planned or Done"`
}

type questionInput struct {
	Slug     string `json:"slug" jsonschema:"spec slug"`
	Question string `json:"question" jsonschema:"one decision for the user, with the context it needs"`
	Owner    string `json:"owner,omitempty" jsonschema:"who decides, default Maintainer"`
}

type answerInput struct {
	Slug   string `json:"slug" jsonschema:"spec slug"`
	QID    string `json:"qid" jsonschema:"question ID, example Q-2"`
	Answer string `json:"answer" jsonschema:"the answer of the user"`
}

type deleteInput struct {
	Slug         string `json:"slug" jsonschema:"spec slug"`
	Reason       string `json:"reason" jsonschema:"why the spec is deleted, for the commit message"`
	MoveTo       string `json:"move_to,omitempty" jsonschema:"slug of an existing spec that gets the other files of the folder (diagrams, designs)"`
	DiscardFiles bool   `json:"discard_files,omitempty" jsonschema:"true deletes the other files of the folder"`
}

// Register adds the spec tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "spec_list", "List specs: slug, status, priority, milestone, title, areas and Done/Total requirements. Filter by status, milestone, priority, area or related; a spec must match all filters. Default: all specs. Do not read docs/specs/README.md; use this tool.",
		func(in listInput) (text string, err error) {
			err = docfile.RLocked(s.Lock, func() (err error) { text, err = s.listText(in); return err })
			return text, err
		})
	add(server, "spec_get", "Read one spec, or one section of it. The full spec and section 5 also give the commits of each requirement from the Refs: <slug>#R-N trailers of HEAD.",
		func(in getInput) (string, error) { return s.getText(in) })
	add(server, "spec_create", "Create a Draft spec in docs/specs/<slug>/SPEC.md from the template. The tool sets the date and the author.",
		func(in createInput) (string, error) {
			return "Created " + in.Slug + ".", s.Create(New(in))
		})
	add(server, "spec_update_section", "Replace one section of a spec.",
		func(in sectionInput) (string, error) {
			review, err := s.UpdateSection(in.Slug, in.Section, in.Body)
			return "Updated section " + in.Section + " of " + in.Slug + "." + scopeText(review, in.Slug), err
		})
	add(server, "spec_set_status", "Change the status of a spec and add a Document History row. The tool refuses a change that breaks a status rule and gives the reasons.",
		func(in statusInput) (string, error) {
			return in.Slug + " is " + in.Status + ".", s.SetStatus(in.Slug, in.Status, in.Note)
		})
	add(server, "spec_req_add", "Add a requirement R-N with the status Planned to a spec.",
		func(in reqAddInput) (string, error) {
			rid, review, err := s.AddRequirement(in.Slug, in.Text)
			return "Added " + rid + "." + scopeText(review, in.Slug), err
		})
	add(server, "spec_req_set", "Change the text, the status, or both of one requirement of a spec.",
		func(in reqSetInput) (string, error) {
			review, err := s.SetRequirement(in.Slug, in.RID, in.Text, in.Status)
			return "Updated " + strings.ToUpper(in.RID) + "." + scopeText(review, in.Slug), err
		})
	add(server, "spec_question_add", "Add an Open question to the Open Questions table of a spec.",
		func(in questionInput) (string, error) {
			qid, review, err := s.AddQuestion(in.Slug, in.Question, in.Owner)
			text := "Added " + qid + "."
			if review {
				text += " " + qid + " is Open, so " + in.Slug + " is In Review. The user must approve it again."
			}
			return text, err
		})
	add(server, "spec_question_answer", "Write the answer of the user to one question of a spec.",
		func(in answerInput) (string, error) {
			return "Answered " + in.QID + ".", s.AnswerQuestion(in.Slug, in.QID, in.Answer)
		})
	add(server, "spec_delete", "Delete a spec folder. Git keeps the old version. The folder must have no changes that are not committed. Give move_to or discard_files for the other files of the folder.",
		func(in deleteInput) (string, error) {
			err := s.Delete(in.Slug, in.Reason, in.MoveTo, in.DiscardFiles)
			return "Deleted " + in.Slug + ". Give this reason in the commit message: " + in.Reason, err
		})
}

// add registers one tool. A missing docs/specs is a normal state, so it is not an error.
func add[In any](server *mcp.Server, name, description string, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			return result.Reply(text, err, ErrNoDir)
		})
}

func scopeText(review bool, slug string) string {
	if !review {
		return ""
	}
	return " The scope changed, so " + slug + " is In Review. The user must approve it again."
}

func (s Store) listText(in listInput) (string, error) {
	specs, warnings, err := s.List(Filter{"status": in.Status, "milestone": in.Milestone, "priority": in.Priority, "area": in.Area, "related": in.Related})
	if err != nil {
		return "", err
	}
	if len(specs) == 0 {
		return "No specs match.\n" + warnings, nil
	}
	rows := make([][]string, 0, len(specs))
	for _, sp := range specs {
		f := sp.Front
		rows = append(rows, []string{f.Slug, f.Status, f.Priority, f.Milestone, f.Title, "[" + strings.Join(f.Area, ", ") + "]",
			fmt.Sprintf("%d/%d", sp.Done, sp.Total)})
	}
	return result.Table(rows) + warnings, nil
}

// getText reads the spec with the shared docs lock, then adds the commits. git runs after the lock,
// because the git data does not change with the docs.
func (s Store) getText(in getInput) (string, error) {
	var text string
	err := docfile.RLocked(s.Lock, func() (err error) { text, err = s.Get(in.Slug, ""); return err })
	if err != nil {
		return "", err
	}
	full := in.Section == ""
	if !full {
		var number string
		if text, number, err = section(text, in.Section); err != nil || number != "5" {
			return text, err
		}
	}
	list, err := s.commits(in.Slug)
	if err != nil {
		return "", err
	}
	return withCommits(text, full, list), nil
}
