package task

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/docs/spec"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// newStore returns a task store and its spec store in a new git repository and data directory, as the
// MCP server makes them.
func newStore(t *testing.T) Store {
	t.Helper()
	return storeAt(t.TempDir(), t.TempDir())
}

// storeAt returns the stores of the repository root and the data directory data.
func storeAt(root, data string) Store {
	lock := filepath.Join(data, project.DocsLock)
	specs := spec.Store{Dir: filepath.Join(root, "docs", "specs"), ADRDir: filepath.Join(root, "docs", "decisions"),
		Root: root, Author: func() string { return "Tester" }, Lock: lock}
	return Store{Dir: filepath.Join(data, project.Tasks), Specs: specs, Lock: lock}
}

// addSpec makes the spec slug with n requirements and the status.
func addSpec(t *testing.T, s Store, slug, status string, n int) {
	t.Helper()
	must(t, s.Specs.Create(spec.New{Slug: slug, Title: "Spec " + slug, Milestone: "v0.1.0", Priority: "P1", Area: []string{"docs"}}))
	for range n {
		if _, _, err := s.Specs.AddRequirement(slug, "A behavior."); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(s.Specs.Dir, slug, "SPEC.md")
	data, err := os.ReadFile(path)
	must(t, err)
	front, body := markdown.Split(string(data))
	must(t, os.WriteFile(path, []byte(markdown.Join(markdown.SetField(front, "status", status), body)), 0o644))
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func connect(t *testing.T, s Store) *mcp.ClientSession {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: "test"}, nil)
	Register(server, s)
	spec.Register(server, s.Specs)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

type step struct {
	name string
	args map[string]any
	want string
}

// run calls each tool and compares the text. "ERR " starts the text of a tool error.
func run(t *testing.T, session *mcp.ClientSession, steps []step) {
	t.Helper()
	for i, st := range steps {
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: st.name, Arguments: st.args})
		if err != nil {
			t.Fatalf("step %d %s: %v", i, st.name, err)
		}
		got := res.Content[0].(*mcp.TextContent).Text
		if res.IsError {
			got = "ERR " + got
		}
		if got != st.want {
			t.Errorf("step %d %s = %q, want %q", i, st.name, got, st.want)
		}
	}
}

// createArgs are the arguments of task_create for the task slug of the spec sp.
func createArgs(sp, slug string, extra map[string]any) map[string]any {
	args := map[string]any{"spec": sp, "slug": slug, "title": "Task " + slug, "priority": "P1",
		"brief": "Do the work.", "plan": []string{"Step one. Check: go test ./..."}}
	for k, v := range extra {
		args[k] = v
	}
	return args
}

func read(t *testing.T, s Store, id string) string {
	t.Helper()
	text, err := s.Get(id, "")
	must(t, err)
	return text
}

func TestCreate(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "docs-lock", "Draft", 3)
	session := connect(t, s)
	long := strings.Repeat("a", 41)
	run(t, session, []step{
		{"task_list", map[string]any{}, "No tasks."},
		{"task_create", createArgs("nope", "x", nil), "ERR No spec has the slug nope."},
		{"task_create", createArgs("docs-lock", "Bad_Slug", nil), "ERR The slug must be kebab-case, 40 characters or less."},
		{"task_create", createArgs("docs-lock", long, nil), "ERR The slug must be kebab-case, 40 characters or less."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"brief": " "}), "ERR Give the brief."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"plan": []string{" "}}), "ERR Give one or more plan steps."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"priority": "P9"}), `ERR Priority "P9" is not valid. Use one of: P0, P1, P2, P3.`},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"covers": []string{"R-9"}}), "ERR Spec docs-lock has no requirement R-9."},
		{"task_create", createArgs("general", "x", map[string]any{"covers": []string{"R-1"}}), "ERR A general task does not cover requirements."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"depends_on": []string{"docs-lock/x"}}), "ERR A task does not depend on itself."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"depends_on": []string{"docs-lock/y"}}), "ERR Task docs-lock/y does not exist."},
		{"task_create", createArgs("docs-lock", "x", map[string]any{"depends_on": []string{"general/y"}}), "ERR Task general/y is in another spec. A task depends only on tasks of its spec."},
		{"task_create", createArgs("docs-lock", "lock-path", map[string]any{"covers": []string{"R-1"}}), "Created task docs-lock/lock-path."},
		{"task_create", createArgs("docs-lock", "lock-path", nil), "ERR Task docs-lock/lock-path already exists."},
		{"task_create", createArgs("docs-lock", "shared-reads", map[string]any{"covers": []string{"R-2", "R-3"}, "depends_on": []string{"docs-lock/lock-path"},
			"plan": []string{"Add SetDoneLocked. Check: go test ./internal/docs/spec", "Call it. Check: go test ./internal/docs/task"}}), "Created task docs-lock/shared-reads."},
		{"task_create", createArgs("general", "notes", nil), "Created task general/notes."},
	})
	want := `---
id: docs-lock/shared-reads
title: Task shared-reads
status: Todo
priority: P1
covers: [R-2, R-3]
depends_on: [docs-lock/lock-path]
---
# Task shared-reads

## Brief

Do the work.

## Plan

- [ ] Add SetDoneLocked. Check: go test ./internal/docs/spec
- [ ] Call it. Check: go test ./internal/docs/task

## Notes

## Definition of Done

- [ ] The covered requirements pass their acceptance criteria, with evidence.
- [ ] Project checks (build, tests, lint) pass.
`
	if got := read(t, s, "docs-lock/shared-reads"); got != want {
		t.Errorf("task file =\n%s\nwant\n%s", got, want)
	}
	if got := read(t, s, "general/notes"); !strings.Contains(got, generalItem) {
		t.Errorf("general task has no %q:\n%s", generalItem, got)
	}
}

func TestListAndGet(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "b-spec", "Approved", 1)
	addSpec(t, s, "a-spec", "Approved", 1)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", createArgs("general", "g", nil), "Created task general/g."},
		{"task_create", createArgs("b-spec", "b", nil), "Created task b-spec/b."},
		{"task_create", createArgs("a-spec", "low", map[string]any{"priority": "P3"}), "Created task a-spec/low."},
		{"task_create", createArgs("a-spec", "high", map[string]any{"priority": "P0"}), "Created task a-spec/high."},
		{"task_create", createArgs("a-spec", "started", map[string]any{"priority": "P3"}), "Created task a-spec/started."},
		{"task_set_status", map[string]any{"id": "a-spec/started", "status": "In Progress"}, "Task a-spec/started is In Progress."},
		{"task_list", map[string]any{}, "a-spec/started  In Progress  P3  Task started\n" +
			"a-spec/high     Todo         P0  Task high\n" +
			"a-spec/low      Todo         P3  Task low\n" +
			"b-spec/b        Todo         P1  Task b\n" +
			"general/g       Todo         P1  Task g\n"},
		{"task_list", map[string]any{"spec": "b-spec"}, "b-spec/b  Todo  P1  Task b\n"},
		{"task_list", map[string]any{"status": "In Progress"}, "a-spec/started  In Progress  P3  Task started\n"},
		{"task_list", map[string]any{"status": "Open"}, `ERR Status "Open" is not valid. Use one of: In Progress, Blocked, Todo, Done.`},
		{"task_get", map[string]any{"id": "b-spec/b", "section": "Brief"}, "Do the work."},
		{"task_get", map[string]any{"id": "b-spec/b", "section": "Goals"}, "ERR Section Goals does not exist. Use Brief, Plan, Notes or Definition of Done."},
		{"task_get", map[string]any{"id": "TASK-001"}, "ERR Task TASK-001 does not exist."},
		{"task_get", map[string]any{"id": "b-spec/nope"}, "ERR Task b-spec/nope does not exist."},
	})
	path := filepath.Join(s.Dir, "b-spec", "b.md")
	data, err := os.ReadFile(path)
	must(t, err)
	must(t, os.WriteFile(path, []byte(strings.Replace(string(data), "id: b-spec/b", "id: b-spec/c", 1)), 0o644))
	run(t, session, []step{
		{"task_get", map[string]any{"id": "b-spec/b"}, "ERR Task file b-spec/b.md has the id b-spec/c."},
		{"task_list", map[string]any{}, "ERR Task file b-spec/b.md has the id b-spec/c."},
	})
}

func TestUpdate(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "a", "Draft", 2)
	session := connect(t, s)
	id := "a/x"
	run(t, session, []step{
		{"task_create", createArgs("a", "y", nil), "Created task a/y."},
		{"task_create", createArgs("a", "x", nil), "Created task a/x."},
		{"task_update_section", map[string]any{"id": id, "section": "notes", "body": "- A note."}, "Updated section notes of a/x."},
		{"task_update_section", map[string]any{"id": id, "section": "Brief", "body": " "}, "ERR Give the brief."},
		{"task_update_section", map[string]any{"id": id, "section": "Plan", "body": ""}, "ERR Give one or more plan steps."},
		{"task_update_section", map[string]any{"id": id, "section": "Definition of Done", "body": "x"},
			"ERR Section Definition of Done changes only with task_dod_add, task_dod_remove, task_dod_check and task_dod_uncheck."},
		{"task_update_section", map[string]any{"id": id, "section": "Goals", "body": "x"}, "ERR Section Goals does not exist. Use Brief, Plan or Notes."},
		{"task_update_fields", map[string]any{"id": id}, "ERR Give a field to change."},
		{"task_update_fields", map[string]any{"id": id, "title": "New title", "priority": "P0", "covers": []string{"R-2"}, "depends_on": []string{"a/y"}}, "Updated task a/x."},
		{"task_update_fields", map[string]any{"id": "a/y", "depends_on": []string{"a/x"}}, "ERR The dependency makes a cycle: a/y -> a/x -> a/y."},
		{"task_update_fields", map[string]any{"id": id, "covers": []string{"R-7"}}, "ERR Spec a has no requirement R-7."},
		{"task_update_fields", map[string]any{"id": id, "depends_on": []string{}}, "Updated task a/x."},
	})
	got := read(t, s, id)
	for _, want := range []string{"title: New title", "priority: P0", "covers: [R-2]", "depends_on: []", "# New title", "## Notes\n\n- A note."} {
		if !strings.Contains(got, want) {
			t.Errorf("task has no %q:\n%s", want, got)
		}
	}
}

func TestStatus(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "draft", "Draft", 1)
	addSpec(t, s, "a", "Approved", 3)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", createArgs("draft", "d", nil), "Created task draft/d."},
		{"task_set_status", map[string]any{"id": "draft/d", "status": "In Progress"}, "ERR Spec draft is Draft. A task starts only when its spec is Approved."},
		{"task_create", createArgs("a", "first", nil), "Created task a/first."},
		{"task_create", createArgs("a", "second", map[string]any{"covers": []string{"R-1", "R-3"}, "depends_on": []string{"a/first"},
			"plan": []string{"One.", "Two."}}), "Created task a/second."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Done"}, "ERR A task does not go from Todo to Done. Allowed: In Progress, Blocked."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Open"}, `ERR Status "Open" is not valid. Use one of: In Progress, Blocked, Todo, Done.`},
		{"task_set_status", map[string]any{"id": "a/second", "status": "In Progress"}, "ERR Task a/first is not Done."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Blocked"}, "ERR Give the blocker in note."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Blocked", "note": "Waits for a/first."}, "Task a/second is Blocked."},
		{"task_set_status", map[string]any{"id": "a/first", "status": "In Progress", "note": "Start."}, "Task a/first is In Progress."},
		{"task_set_status", map[string]any{"id": "a/first", "status": "Done"},
			"ERR Plan step 1 and Definition of Done items 1, 2 are not checked."},
		{"task_step_check", map[string]any{"id": "a/first", "step": 2}, "ERR Step 2 does not exist. The Plan has 1 steps."},
		{"task_step_check", map[string]any{"id": "a/first", "step": 1}, "Checked step 1 of a/first."},
		{"task_step_check", map[string]any{"id": "a/first", "step": 1}, "ERR Step 1 is already checked."},
		{"task_dod_check", map[string]any{"id": "a/first", "item": 1, "evidence": "Test."}, "Checked item 1 of a/first."},
		{"task_set_status", map[string]any{"id": "a/first", "status": "Done"}, "ERR Definition of Done item 2 is not checked."},
		{"task_dod_check", map[string]any{"id": "a/first", "item": 2, "evidence": "CI."}, "Checked item 2 of a/first."},
		{"task_set_status", map[string]any{"id": "a/first", "status": "Done"}, "Task a/first is Done."},
		{"task_set_status", map[string]any{"id": "a/first", "status": "Todo"}, "ERR Task a/first is Done and does not change."},
		{"task_update_section", map[string]any{"id": "a/first", "section": "Notes", "body": "x"}, "ERR Task a/first is Done and does not change."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "In Progress"}, "Task a/second is In Progress."},
		{"task_step_check", map[string]any{"id": "a/second", "step": 1}, "Checked step 1 of a/second."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Done"}, "ERR Plan step 2 and Definition of Done items 1, 2 are not checked."},
		{"task_step_check", map[string]any{"id": "a/second", "step": 2}, "Checked step 2 of a/second."},
		{"task_dod_check", map[string]any{"id": "a/second", "item": 1, "evidence": "Test."}, "Checked item 1 of a/second."},
		{"task_dod_check", map[string]any{"id": "a/second", "item": 2, "evidence": "CI."}, "Checked item 2 of a/second."},
		{"task_set_status", map[string]any{"id": "a/second", "status": "Done"}, "Task a/second is Done. Spec a: R-1, R-3: Done."},
	})
	if got := read(t, s, "a/second"); !strings.Contains(got, "## Notes\n\n- Blocked: Waits for a/first.") {
		t.Errorf("Notes have no Blocked line:\n%s", got)
	}
	if got := read(t, s, "a/first"); !strings.Contains(got, "- In Progress: Start.") {
		t.Errorf("Notes have no In Progress line:\n%s", got)
	}
	sp, err := s.Specs.Get("a", "")
	must(t, err)
	for _, want := range []string{"| R-1 | A behavior. | Done |", "| R-2 | A behavior. | Planned |", "| R-3 | A behavior. | Done |", "| Tester | R-1, R-3: Done. |"} {
		if !strings.Contains(sp, want) {
			t.Errorf("spec has no %q:\n%s", want, sp)
		}
	}
}

func TestDoD(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	id := "general/x"
	run(t, session, []step{
		{"task_create", createArgs("general", "x", nil), "Created task general/x."},
		{"task_dod_check", map[string]any{"id": id, "item": 1, "evidence": " "}, "ERR Give the evidence."},
		{"task_dod_check", map[string]any{"id": id, "item": 3, "evidence": "x"}, "ERR Item 3 does not exist. The Definition of Done has 2 items."},
		{"task_dod_check", map[string]any{"id": id, "item": 1, "evidence": "Test."}, "Checked item 1 of general/x."},
		{"task_dod_check", map[string]any{"id": id, "item": 1, "evidence": "Test."}, "ERR Item 1 is already checked."},
		{"task_dod_uncheck", map[string]any{"id": id, "item": 2}, "ERR Item 2 is not checked."},
		{"task_dod_uncheck", map[string]any{"id": id, "item": 1}, "Unchecked item 1 of general/x."},
		{"task_dod_add", map[string]any{"id": id, "text": "Docs are updated."}, "Added item 3 to general/x."},
		{"task_dod_remove", map[string]any{"id": id, "item": 3, "reason": " "}, "ERR Give the reason."},
		{"task_dod_remove", map[string]any{"id": id, "item": 3, "reason": "No docs change."}, "Removed item 3 of general/x."},
	})
	got := read(t, s, id)
	want := "## Notes\n\n- Definition of Done item 3 removed (Docs are updated.): No docs change.\n\n## Definition of Done\n\n" +
		generalItem + "\n- [ ] Project checks (build, tests, lint) pass.\n"
	if !strings.HasSuffix(got, want) {
		t.Errorf("task =\n%s\nwant suffix\n%s", got, want)
	}
}

func TestMoveAndDelete(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "old", "Approved", 2)
	addSpec(t, s, "new", "Draft", 0)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", createArgs("old", "a", map[string]any{"covers": []string{"R-1", "R-2"}}), "Created task old/a."},
		{"task_create", createArgs("old", "b", map[string]any{"depends_on": []string{"old/a"}}), "Created task old/b."},
		{"task_create", createArgs("old", "c", nil), "Created task old/c."},
		{"task_create", createArgs("new", "c", nil), "Created task new/c."},
		{"task_move", map[string]any{"ids": []string{}, "to": "new"}, "ERR Give one or more task IDs."},
		{"task_move", map[string]any{"ids": []string{"old/a", "new/c"}, "to": "general"}, "ERR The tasks must be of one spec."},
		{"task_move", map[string]any{"ids": []string{"old/a"}, "to": "old"}, "ERR The tasks are in old already."},
		{"task_move", map[string]any{"ids": []string{"old/a"}, "to": "nope"}, "ERR No spec has the slug nope."},
		{"task_move", map[string]any{"ids": []string{"old/zzz"}, "to": "new"}, "ERR Task old/zzz does not exist."},
		{"task_move", map[string]any{"ids": []string{"old/c"}, "to": "new"}, "ERR Task new/c already exists."},
		{"task_move", map[string]any{"ids": []string{"old/a"}, "to": "new"}, "ERR Task old/b depends on task old/a, which goes to new."},
		{"task_move", map[string]any{"ids": []string{"old/b"}, "to": "new"}, "ERR Task old/b depends on task old/a, which stays in old."},
		{"spec_delete", map[string]any{"slug": "old", "reason": "Split."}, "ERR Spec old has 3 tasks. Move or delete the tasks first, with task_move or task_delete."},
		{"task_move", map[string]any{"ids": []string{"old/a", "old/b"}, "to": "new"}, "Moved 2 tasks to new."},
		// A retry after the work finds the moved tasks and changes nothing.
		{"task_move", map[string]any{"ids": []string{"old/a", "old/b"}, "to": "new"}, "Moved 2 tasks to new."},
		{"task_delete", map[string]any{"ids": []string{}}, "ERR Give one or more task IDs."},
		{"task_delete", map[string]any{"ids": []string{"new/a"}}, "ERR Task new/b depends on task new/a."},
		{"task_delete", map[string]any{"ids": []string{"old/zzz"}}, "ERR Task old/zzz does not exist."},
		{"task_delete", map[string]any{"ids": []string{"old/c"}}, "Deleted 1 task."},
		{"task_list", map[string]any{"spec": "new"}, "new/a  Todo  P1  Task a\nnew/b  Todo  P1  Task b\nnew/c  Todo  P1  Task c\n"},
	})
	got := read(t, s, "new/a")
	for _, want := range []string{"id: new/a", "covers: []", "## Notes\n\n- Moved from old. Old covers: R-1, R-2."} {
		if !strings.Contains(got, want) {
			t.Errorf("moved task has no %q:\n%s", want, got)
		}
	}
	if got := read(t, s, "new/b"); !strings.Contains(got, "depends_on: [new/a]") || !strings.Contains(got, "- Moved from old.\n") {
		t.Errorf("moved task b:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "old")); !os.IsNotExist(err) {
		t.Errorf("the empty folder of old exists: %v", err)
	}
	commitAll(t, s.Specs.Root)
	run(t, session, []step{
		{"spec_delete", map[string]any{"slug": "old", "reason": "Split."}, "Deleted old. Give this reason in the commit message: Split."},
		{"task_delete", map[string]any{"ids": []string{"new/a", "new/b", "new/c"}}, "Deleted 3 tasks."},
		{"task_list", map[string]any{}, "No tasks."},
	})
}

// TestMoveRetry is Q-43: a crash after the write of the new file and before the remove of the old file
// leaves both files. A retry removes the old file.
func TestMoveRetry(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "new", "Draft", 0)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", createArgs("general", "a", nil), "Created task general/a."},
	})
	old := filepath.Join(s.Dir, "general", "a.md")
	data, err := os.ReadFile(old)
	must(t, err)
	run(t, session, []step{
		{"task_move", map[string]any{"ids": []string{"general/a"}, "to": "new"}, "Moved 1 task to new."},
	})
	must(t, os.MkdirAll(filepath.Dir(old), 0o755))
	must(t, os.WriteFile(old, data, 0o644))
	run(t, session, []step{
		{"task_move", map[string]any{"ids": []string{"general/a"}, "to": "new"}, "Moved 1 task to new."},
		{"task_list", map[string]any{}, "new/a  Todo  P1  Task a\n"},
	})
	if got := read(t, s, "new/a"); strings.Count(got, "Moved from general.") != 1 {
		t.Errorf("task =\n%s", got)
	}
}

// commitAll commits all files of the repository root, so spec_delete accepts the folder.
func commitAll(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "commit.gpgsign=false", "-c", "user.name=Tester", "-c", "user.email=t@example.com",
		"commit", "-q", "--allow-empty", "-m", "docs: specs"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
