package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStrayFileInSpecFolder(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "a", "Draft", 2)
	session := connect(t, s)
	run(t, session, []step{{"task_create", createArgs("a", "x", nil), "Created task a/x."}})
	must(t, os.WriteFile(filepath.Join(s.Dir, "a", "README.md"), []byte("# Notes\n"), 0o644))
	run(t, session, []step{
		{"task_list", map[string]any{}, "a/x  Todo  P1  Task x\nWarning: a/README.md: task file a/README.md has the id .\n"},
	})
}

func TestCreateChecks(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "a", "Draft", 2)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", func() map[string]any { a := createArgs("a", "x", nil); a["title"] = "A\nB"; return a }(),
			"ERR The title has a line break. Write the title on one line."},
		{"task_create", func() map[string]any { a := createArgs("a", "x", nil); a["brief"] = "### Notes\nx"; return a }(),
			`ERR The content of section "Brief" has the heading "Notes", and the document has a section with the same title or number. Rename the heading.`},
		{"task_create", func() map[string]any { a := createArgs("a", "x", nil); a["brief"] = "```\nx"; return a }(),
			`ERR The content of section "Brief" has a code fence that is not closed. Close it with the same fence.`},
		{"task_create", createArgs("a", "ok", map[string]any{"covers": []string{"r-2"}}), "Created task a/ok."},
	})
	if text := read(t, s, "a/ok"); !strings.Contains(text, "covers: [R-2]\n") {
		t.Errorf("covers is not canonical:\n%s", text)
	}
	run(t, session, []step{
		{"task_update_fields", map[string]any{"id": "a/ok", "covers": []string{"r-1"}}, "Updated task a/ok."},
	})
	if text := read(t, s, "a/ok"); !strings.Contains(text, "covers: [R-1]\n") {
		t.Errorf("updated covers is not canonical:\n%s", text)
	}
}

func TestPlanUpdate(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "a", "Draft", 1)
	session := connect(t, s)
	run(t, session, []step{
		{"task_create", createArgs("a", "x", nil), "Created task a/x."},
		{"task_update_section", map[string]any{"id": "a/x", "section": "Plan", "body": "- [ ] One\n\n- [ ] Two"}, "Updated section Plan of a/x."},
		{"task_update_section", map[string]any{"id": "a/x", "section": "Plan", "body": "- [x] One"},
			`ERR The Plan has a checked step. Write only steps like "- [ ] text". Use task_step_check to check a step.`},
		{"task_update_section", map[string]any{"id": "a/x", "section": "Plan", "body": "One"},
			`ERR The Plan has the line "One". Write each step as "- [ ] text".`},
	})
}
