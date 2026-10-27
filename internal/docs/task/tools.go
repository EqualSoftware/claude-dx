package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/result"
)

type listInput struct {
	Spec   string `json:"spec,omitempty" jsonschema:"spec slug, or general. Default: all specs"`
	Status string `json:"status,omitempty" jsonschema:"In Progress, Blocked, Todo or Done. Default: all"`
}

type getInput struct {
	ID      string `json:"id" jsonschema:"task ID <spec>/<task-slug>, example docs-lock/shared-reads"`
	Section string `json:"section,omitempty" jsonschema:"one section only: Brief, Plan, Notes or Definition of Done"`
}

// createInput has the same fields as New, so one converts to the other.
type createInput struct {
	Spec      string   `json:"spec" jsonschema:"slug of the spec of the task, or general for work that has no spec"`
	Slug      string   `json:"slug" jsonschema:"kebab-case ID of the task in its spec, 40 characters or less, example shared-reads. It does not change"`
	Title     string   `json:"title" jsonschema:"short task title, in ASD-STE100, 40 characters or less"`
	Priority  string   `json:"priority" jsonschema:"P0, P1, P2 or P3"`
	Brief     string   `json:"brief" jsonschema:"the full brief, so a subagent can do the task with no other context: what to do, files, constraints, check commands. The why is in the covers requirements of the spec"`
	Plan      []string `json:"plan" jsonschema:"one or more steps, each one action with its check, example: Add SetDoneLocked. Check: go test ./internal/docs/spec"`
	Covers    []string `json:"covers,omitempty" jsonschema:"requirement IDs of the spec that the task implements, example [R-2, R-3]. A general task has none"`
	DependsOn []string `json:"depends_on,omitempty" jsonschema:"IDs of tasks of the same spec that must be Done first"`
}

type sectionInput struct {
	ID      string `json:"id" jsonschema:"task ID"`
	Section string `json:"section" jsonschema:"Brief, Plan or Notes"`
	Body    string `json:"body" jsonschema:"new content of the section, with no heading. A Plan is a checklist: - [ ] step"`
}

type fieldsInput struct {
	ID        string   `json:"id" jsonschema:"task ID"`
	Title     string   `json:"title,omitempty" jsonschema:"new title, 40 characters or less"`
	Priority  string   `json:"priority,omitempty" jsonschema:"P0, P1, P2 or P3"`
	Covers    []string `json:"covers,omitempty" jsonschema:"the full new list; an empty list removes all"`
	DependsOn []string `json:"depends_on,omitempty" jsonschema:"the full new list; an empty list removes all"`
}

type statusInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	Status string `json:"status" jsonschema:"Todo, In Progress, Blocked or Done"`
	Note   string `json:"note,omitempty" jsonschema:"a line for Notes. Required for Blocked: the blocker"`
}

type stepInput struct {
	ID   string `json:"id" jsonschema:"task ID"`
	Step int    `json:"step" jsonschema:"Plan step number; 1 is the first step"`
}

type dodInput struct {
	ID       string `json:"id" jsonschema:"task ID"`
	Item     int    `json:"item" jsonschema:"Definition of Done item number; 1 is the first item"`
	Evidence string `json:"evidence" jsonschema:"verified evidence: command output, CI run, path:line"`
}

type dodItemInput struct {
	ID   string `json:"id" jsonschema:"task ID"`
	Item int    `json:"item" jsonschema:"Definition of Done item number; 1 is the first item"`
}

type dodReasonInput struct {
	ID     string `json:"id" jsonschema:"task ID"`
	Item   int    `json:"item" jsonschema:"Definition of Done item number; 1 is the first item"`
	Reason string `json:"reason" jsonschema:"why the item goes. The tool adds the item and the reason to Notes"`
}

type dodAddInput struct {
	ID   string `json:"id" jsonschema:"task ID"`
	Text string `json:"text" jsonschema:"the new Definition of Done item, one line"`
}

type moveInput struct {
	IDs []string `json:"ids" jsonschema:"IDs of one or more tasks of one spec"`
	To  string   `json:"to" jsonschema:"slug of the target spec, or general"`
}

type deleteInput struct {
	IDs []string `json:"ids" jsonschema:"IDs of one or more tasks"`
}

// Register adds the task tools to the server.
func Register(server *mcp.Server, s Store) {
	add(server, "task_list", "List the local tasks: ID, status, priority, title. Order: spec (general last), status (In Progress, Blocked, Todo, Done), priority, ID.",
		func(in listInput) (text string, err error) {
			err = docfile.RLocked(s.Lock, func() (err error) { text, err = s.listText(in); return err })
			return text, err
		})
	add(server, "task_get", "Read one local task, or one section of it.",
		func(in getInput) (text string, err error) {
			err = docfile.RLocked(s.Lock, func() (err error) { text, err = s.Get(in.ID, in.Section); return err })
			return text, err
		})
	add(server, "task_create", "Create a Todo task of a spec in the project data directory, with a full Brief and a Plan.",
		func(in createInput) (string, error) {
			id, err := s.Create(New(in))
			return "Created task " + id + ".", err
		})
	add(server, "task_update_section", "Replace the Brief, the Plan or the Notes of a task.",
		func(in sectionInput) (string, error) {
			return "Updated section " + in.Section + " of " + in.ID + ".", s.UpdateSection(in.ID, in.Section, in.Body)
		})
	add(server, "task_update_fields", "Change the title, the priority, the covers or the dependencies of a task.",
		func(in fieldsInput) (string, error) {
			u := Update{Title: in.Title, Priority: in.Priority, Covers: in.Covers, DependsOn: in.DependsOn}
			return "Updated task " + in.ID + ".", s.UpdateFields(in.ID, u)
		})
	add(server, "task_set_status", "Change the status of a task. In Progress needs an Approved spec and Done dependencies. Done needs all Plan steps and Definition of Done items checked, and sets the covers requirements to Done in the spec. Done is final.",
		func(in statusInput) (string, error) {
			covers, err := s.SetStatus(in.ID, in.Status, in.Note)
			return statusText(in, covers), err
		})
	add(server, "task_step_check", "Check one Plan step of a task.",
		func(in stepInput) (string, error) {
			return fmt.Sprintf("Checked step %d of %s.", in.Step, in.ID), s.CheckStep(in.ID, in.Step)
		})
	add(server, "task_dod_check", "Check one Definition of Done item of a task and add its evidence.",
		func(in dodInput) (string, error) {
			return fmt.Sprintf("Checked item %d of %s.", in.Item, in.ID), s.CheckDoD(in.ID, in.Item, in.Evidence)
		})
	add(server, "task_dod_uncheck", "Uncheck one Definition of Done item of a task and remove its evidence.",
		func(in dodItemInput) (string, error) {
			return fmt.Sprintf("Unchecked item %d of %s.", in.Item, in.ID), s.UncheckDoD(in.ID, in.Item)
		})
	add(server, "task_dod_add", "Add an unchecked item at the end of the Definition of Done of a task.",
		func(in dodAddInput) (string, error) {
			n, err := s.AddDoD(in.ID, in.Text)
			return fmt.Sprintf("Added item %d to %s.", n, in.ID), err
		})
	add(server, "task_dod_remove", "Remove one Definition of Done item of a task. The item text and the reason go to Notes.",
		func(in dodReasonInput) (string, error) {
			return fmt.Sprintf("Removed item %d of %s.", in.Item, in.ID), s.RemoveDoD(in.ID, in.Item, in.Reason)
		})
	add(server, "task_move", "Move one or more tasks of one spec to another spec or to general. The tool removes covers and adds the old covers to Notes. Move or delete the tasks of a spec before spec_delete.",
		func(in moveInput) (string, error) {
			return "Moved " + count(len(in.IDs)) + " to " + in.To + ".", s.Move(in.IDs, in.To)
		})
	add(server, "task_delete", "Delete one or more tasks. A task that another task depends on must be in the same call.",
		func(in deleteInput) (string, error) {
			return "Deleted " + count(len(in.IDs)) + ".", s.Delete(in.IDs)
		})
}

// add registers one tool.
func add[In any](server *mcp.Server, name, description string, fn func(In) (string, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(_ context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			text, err := fn(in)
			if err != nil {
				return result.Errorf("%s", result.Sentence(err)), nil, nil
			}
			return result.Text(text), nil, nil
		})
}

func (s Store) listText(in listInput) (string, error) {
	tasks, warnings, err := s.List(in.Spec, in.Status)
	if err != nil {
		return "", err
	}
	if len(tasks) == 0 {
		return strings.TrimSpace("No tasks.\n" + warnings), nil
	}
	rows := make([][]string, 0, len(tasks))
	for _, f := range tasks {
		rows = append(rows, []string{f.ID, f.Status, f.Priority, f.Title})
	}
	return result.Table(rows) + warnings, nil
}

// statusText is the result of task_set_status. A Done task with covers names the requirements.
func statusText(in statusInput, covers []string) string {
	text := "Task " + in.ID + " is " + in.Status + "."
	if len(covers) == 0 {
		return text
	}
	slug, _, _ := strings.Cut(in.ID, "/")
	return text + " Spec " + slug + ": " + strings.Join(covers, ", ") + ": Done."
}

// count returns "1 task" or "n tasks".
func count(n int) string {
	if n == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", n)
}
