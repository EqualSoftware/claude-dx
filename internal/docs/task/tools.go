package task

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/result"
)

// listInput has one filter for each frontmatter field. The JSON name is the frontmatter key.
type listInput struct {
	Status    string `json:"status,omitempty" jsonschema:"open (default), all, ready (Approved tasks with all dependencies Done, in milestone order), or one status: Draft, In Review, Approved, In Development, Blocked, Done, Cancelled"`
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty" jsonschema:"the full title"`
	Priority  string `json:"priority,omitempty" jsonschema:"P0, P1, P2 or P3"`
	Milestone string `json:"milestone,omitempty" jsonschema:"one milestone, example v0.1.0"`
	Type      string `json:"type,omitempty" jsonschema:"feat, fix, refactor, perf, test, docs, build, ci or chore"`
	Size      string `json:"size,omitempty" jsonschema:"S or M"`
	Owner     string `json:"owner,omitempty" jsonschema:"git user.name of the person who created the task"`
	Platforms string `json:"platforms,omitempty" jsonschema:"one platform"`
	Area      string `json:"area,omitempty" jsonschema:"one area"`
	DependsOn string `json:"depends_on,omitempty" jsonschema:"one task ID: the tasks that wait for it"`
	Related   string `json:"related,omitempty" jsonschema:"one ADR or task ID"`
}

// filter returns the filter with the frontmatter key of each input field.
func (in listInput) filter() Filter {
	f := Filter{}
	v := reflect.ValueOf(in)
	for i := range v.NumField() {
		key, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
		f[key] = v.Field(i).String()
	}
	return f
}

type getInput struct {
	ID      string `json:"id" jsonschema:"task ID, example TASK-007"`
	Section string `json:"section,omitempty" jsonschema:"one section only, by number (example 7.1) or title. Read one section when you need only that part"`
}

// createInput has the same fields as New, so one converts to the other.
type createInput struct {
	Title     string            `json:"title" jsonschema:"short task title, in ASD-STE100, 40 characters or less, not the title of another task"`
	Priority  string            `json:"priority" jsonschema:"P0, P1, P2 or P3"`
	Milestone string            `json:"milestone" jsonschema:"the release of the task, vMAJOR.MINOR.PATCH, example v0.1.0"`
	Type      string            `json:"type" jsonschema:"feat, fix, refactor, perf, test, docs, build, ci or chore. It sets the commit type"`
	Area      []string          `json:"area" jsonschema:"the areas of the project, example [api] or [docs, ci]"`
	DependsOn []string          `json:"depends_on,omitempty" jsonschema:"open task IDs that must be done first"`
	Size      string            `json:"size,omitempty" jsonschema:"S (less than 1 day) or M (1 to 3 days). Split a larger task. Required for In Review"`
	Platforms []string          `json:"platforms,omitempty" jsonschema:"operating systems, architectures or runtimes, example [macOS, Linux, Windows], or [Not applicable]. Required for In Review"`
	Related   []string          `json:"related,omitempty" jsonschema:"ADR and task IDs, docs, issues, external references"`
	Sections  map[string]string `json:"sections,omitempty" jsonschema:"section number or title -> body with no heading. Sections that you do not give keep the template text"`
}

type sectionInput struct {
	ID      string `json:"id" jsonschema:"task ID"`
	Section string `json:"section" jsonschema:"section number (example 7.1) or title"`
	Body    string `json:"body" jsonschema:"new content of the section, with no heading"`
}

type fieldsInput struct {
	ID        string   `json:"id" jsonschema:"task ID"`
	Title     string   `json:"title,omitempty" jsonschema:"new title, 40 characters or less"`
	Priority  string   `json:"priority,omitempty" jsonschema:"P0, P1, P2 or P3"`
	Milestone string   `json:"milestone,omitempty" jsonschema:"vMAJOR.MINOR.PATCH, example v0.1.0"`
	Type      string   `json:"type,omitempty" jsonschema:"feat, fix, refactor, perf, test, docs, build, ci or chore"`
	Size      string   `json:"size,omitempty" jsonschema:"S or M"`
	Owner     string   `json:"owner,omitempty" jsonschema:"git user.name of the person who owns the task"`
	Platforms []string `json:"platforms,omitempty" jsonschema:"the full new list"`
	Area      []string `json:"area,omitempty" jsonschema:"the full new list"`
	DependsOn []string `json:"depends_on,omitempty" jsonschema:"the full new list; an empty list removes all"`
	Related   []string `json:"related,omitempty" jsonschema:"the full new list; an empty list removes all"`
}

type statusInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	Status string `json:"status" jsonschema:"Draft, In Review, Approved, In Development, Blocked, Done or Cancelled"`
	Note   string `json:"note,omitempty" jsonschema:"Document History note. Required for Blocked (the blocker) and Cancelled (the reason)"`
}

type questionInput struct {
	ID       string `json:"id" jsonschema:"task ID"`
	Question string `json:"question" jsonschema:"one decision for the user, with the context it needs"`
	Owner    string `json:"owner,omitempty" jsonschema:"who decides, default Maintainer"`
}

type answerInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	QID    string `json:"qid" jsonschema:"question ID, example Q-2"`
	Answer string `json:"answer" jsonschema:"the answer of the user"`
}

type dodInput struct {
	ID       string `json:"id" jsonschema:"task ID"`
	Item     int    `json:"item" jsonschema:"Definition of Done item number; 1 is the first item"`
	Evidence string `json:"evidence" jsonschema:"verified evidence: command output, CI run, path:line"`
}

type dodReasonInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	Item   int    `json:"item" jsonschema:"Definition of Done item number; 1 is the first item"`
	Reason string `json:"reason" jsonschema:"why the item changes. The tool adds it to the Document History"`
}

type dodAddInput struct {
	ID   string `json:"id" jsonschema:"task ID"`
	Text string `json:"text" jsonschema:"the new Definition of Done item, one line"`
}

type renumberInput struct {
	File string `json:"file" jsonschema:"task file name, example TASK-005-add-cache.md. Use the file name, because two files can have one number"`
}

// Register adds the task tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "task_list", "List tasks: ID, status, priority, milestone, title, areas and dependencies. Filter by any frontmatter field; a task must match all filters. Default: open tasks. A dependency that is not Done shows its status. Do not read docs/tasks/README.md; use this tool.",
		func(in listInput) (string, error) { return s.listText(in) })
	add(server, "task_get", "Read one task, or one section of it.",
		func(in getInput) (string, error) { return s.Get(in.ID, in.Section) })
	add(server, "task_create", "Create a Draft task in docs/tasks from the template. The tool sets the number, the date, the author and the file name.",
		func(in createInput) (string, error) { return s.createText(in) })
	add(server, "task_update_section", "Replace one section of a task.",
		func(in sectionInput) (string, error) {
			review, err := s.UpdateSection(in.ID, in.Section, in.Body)
			if review {
				return "Updated section " + in.Section + " of " + in.ID + ". The scope changed, so " + in.ID + " is In Review. The user must approve it again.", err
			}
			return "Updated section " + in.Section + " of " + in.ID + ".", err
		})
	add(server, "task_update_fields", "Change the frontmatter fields of a task: title, priority, milestone, type, size, owner, platforms, areas, dependencies or related.",
		func(in fieldsInput) (string, error) {
			u := Update{Title: in.Title, Priority: in.Priority, Milestone: in.Milestone, Type: in.Type, Size: in.Size, Owner: in.Owner,
				Platforms: in.Platforms, Area: in.Area, DependsOn: in.DependsOn, Related: in.Related}
			file, warnings, err := s.UpdateFields(in.ID, u)
			if file != "" {
				file = " File: " + file + "."
			}
			return strings.TrimSpace("Updated " + in.ID + "." + file + "\n" + warnings), err
		})
	add(server, "task_set_status", "Change the status of a task and add a Document History row. The tool refuses a change that breaks a status rule and gives the reasons.",
		func(in statusInput) (string, error) {
			warnings, err := s.SetStatus(in.ID, in.Status, in.Note)
			return strings.TrimSpace(in.ID + " is " + in.Status + ".\n" + warnings), err
		})
	add(server, "task_question_add", "Add an Open question to the Open Questions table of a task.",
		func(in questionInput) (string, error) {
			qid, err := s.AddQuestion(in.ID, in.Question, in.Owner)
			return "Added " + qid + ".", err
		})
	add(server, "task_question_answer", "Write the answer of the user to one question of a task.",
		func(in answerInput) (string, error) {
			return "Answered " + in.QID + ".", s.AnswerQuestion(in.ID, in.QID, in.Answer)
		})
	add(server, "task_dod_check", "Check one Definition of Done item of a task and add its evidence.",
		func(in dodInput) (string, error) {
			return fmt.Sprintf("Checked item %d of %s.", in.Item, in.ID), s.CheckDoD(in.ID, in.Item, in.Evidence)
		})
	add(server, "task_dod_uncheck", "Uncheck one Definition of Done item of a task and remove its evidence. The reason goes to the Document History.",
		func(in dodReasonInput) (string, error) {
			return fmt.Sprintf("Unchecked item %d of %s.", in.Item, in.ID), s.UncheckDoD(in.ID, in.Item, in.Reason)
		})
	add(server, "task_dod_add", "Add an unchecked item at the end of the Definition of Done of a task.",
		func(in dodAddInput) (string, error) {
			n, err := s.AddDoD(in.ID, in.Text)
			return fmt.Sprintf("Added item %d to %s.", n, in.ID), err
		})
	add(server, "task_dod_remove", "Remove one Definition of Done item of a task. The item text and the reason go to the Document History.",
		func(in dodReasonInput) (string, error) {
			return fmt.Sprintf("Removed item %d of %s.", in.Item, in.ID), s.RemoveDoD(in.ID, in.Item, in.Reason)
		})
	add(server, "task_renumber", "Give a task file the next free number, for example when two files have the same number.",
		func(in renumberInput) (string, error) { return s.renumberText(in.File) })
}

// add registers one tool. A missing docs/tasks is a normal state, so it is not an error.
func add[In any](server *mcp.Server, name, description string, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			return result.Reply(text, err, ErrNoDir)
		})
}

func (s Store) listText(in listInput) (string, error) {
	if in.Status == "" {
		in.Status = "open"
	}
	tasks, warnings, err := s.List(in.filter())
	if err != nil {
		return "", err
	}
	if len(tasks) == 0 {
		return "No tasks match.\n" + warnings, nil
	}
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		f := t.Front
		rows = append(rows, []string{f.ID, f.Status, f.Priority, f.Milestone, f.Title, "[" + strings.Join(f.Area, ", ") + "]", "deps:" + t.Deps})
	}
	return result.Table(rows) + warnings, nil
}

func (s Store) createText(in createInput) (string, error) {
	path, warnings, err := s.Create(New(in))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace("Created docs/tasks/" + filepath.Base(path) + " (Draft).\n" + warnings), nil
}

func (s Store) renumberText(file string) (string, error) {
	newID, refs, err := s.Renumber(file)
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("%s is now %s.", file, newID)
	if len(refs) > 0 {
		text += " These tasks refer to the old ID. Check each one and change it with task_update_fields if it means this task: " + strings.Join(refs, ", ") + "."
	}
	return text, nil
}
