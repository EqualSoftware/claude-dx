package spec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readSpec(t *testing.T, s Store, slug string) string {
	t.Helper()
	data, err := os.ReadFile(s.path(slug))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestUpdateFieldsErrors(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_create", createArgs("b", "B"), "Created b."},
	})
	before := readSpec(t, s, "a")
	run(t, session, []step{
		{"spec_update_fields", map[string]any{"slug": "a"}, "ERR Give a field to change."},
		{"spec_update_fields", map[string]any{"slug": "a", "title": "b", "priority": "P0"}, "ERR B has the same title. Give a different title, or change b."},
		{"spec_update_fields", map[string]any{"slug": "a", "title": strings.Repeat("a", 41)}, "ERR The title has 41 characters. The limit is 40. Write a shorter title."},
		{"spec_update_fields", map[string]any{"slug": "a", "milestone": "1.0", "priority": "P0"}, `ERR Milestone "1.0" is not valid. Use vMAJOR.MINOR.PATCH, example v0.1.0.`},
		{"spec_update_fields", map[string]any{"slug": "a", "priority": "P9"}, `ERR Priority "P9" is not valid. Use one of: P0, P1, P2, P3.`},
		{"spec_update_fields", map[string]any{"slug": "nope", "priority": "P0"}, `ERR Spec "nope" does not exist.`},
	})
	err := s.UpdateFields("a", Update{Related: []string{"ADR-099"}, Priority: "P0"})
	if err == nil || !strings.HasPrefix(err.Error(), `related "ADR-099"`) {
		t.Errorf("related error = %v", err)
	}
	if got := readSpec(t, s, "a"); got != before {
		t.Errorf("an error changed the spec:\n%s", got)
	}
}

func TestUpdateFieldsChanges(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
	})
	for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
		run(t, session, []step{{"spec_update_section", map[string]any{"slug": "a", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of a."}})
	}
	run(t, session, []step{
		{"spec_set_status", map[string]any{"slug": "a", "status": "In Review"}, "a is In Review."},
		{"spec_set_status", map[string]any{"slug": "a", "status": "Approved"}, "a is Approved."},
		// The spec keeps its own title, and a list replaces the full list.
		{"spec_update_fields", map[string]any{"slug": "a", "title": "A", "area": []string{"mcp", "docs"}}, "Updated a."},
		// Fields that are not given stay.
		{"spec_update_fields", map[string]any{"slug": "a", "title": "New title", "milestone": "v0.2.0", "priority": "P0", "related": []string{}}, "Updated a."},
		{"spec_update_fields", map[string]any{"slug": "a", "area": []string{}}, "Updated a."},
	})
	got := readSpec(t, s, "a")
	for _, want := range []string{"title: New title", "milestone: v0.2.0", "priority: P0", "area: []", "related: []",
		"status: Approved", "# New title",
		"| Tester | Changed fields: title, area. |",
		"| Tester | Changed fields: title, milestone, priority, related. |"} {
		if !strings.Contains(got, want) {
			t.Errorf("spec has no %q:\n%s", want, got)
		}
	}
	index, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if !strings.Contains(string(index), "| New title | `Approved` | v0.2.0 | `P0` |") {
		t.Errorf("README.md is not updated:\n%s", index)
	}
}

func TestUpdateFieldsSchema(t *testing.T) {
	session := connect(t, newStore(t))
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name != "spec_update_fields" {
			continue
		}
		props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
		for _, name := range []string{"title", "milestone", "priority", "area", "related"} {
			if _, ok := props[name]; !ok {
				t.Errorf("no parameter %q", name)
			}
		}
		for _, name := range []string{"status", "new_slug"} {
			if _, ok := props[name]; ok {
				t.Errorf("unexpected parameter %q", name)
			}
		}
		return
	}
	t.Fatal("no tool spec_update_fields")
}

func TestUpdateFieldsKeepsFencedComment(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	block := "```sh\n# install deps\nmake\n```"
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_update_section", map[string]any{"slug": "a", "section": "6", "body": block}, "Updated section 6 of a."},
		{"spec_update_fields", map[string]any{"slug": "a", "title": "New title"}, "Updated a."},
	})
	got := readSpec(t, s, "a")
	if !strings.Contains(got, "\n# install deps\n") || !strings.Contains(got, "\n# New title\n") || strings.Count(got, "\n# New title\n") != 1 {
		t.Errorf("heading change is wrong:\n%s", got)
	}
}

func TestStoreUpdateFields(t *testing.T) {
	s := newStore(t)
	run(t, connect(t, s), []step{{"spec_create", createArgs("a", "A"), "Created a."}})
	if err := s.UpdateFields("a", Update{Area: []string{}}); err != nil {
		t.Fatal(err)
	}
	got := readSpec(t, s, "a")
	for _, want := range []string{"area: []", "related: [ADR-008]", "priority: P1", "milestone: v0.1.0", "title: A"} {
		if !strings.Contains(got, want) {
			t.Errorf("spec has no %q:\n%s", want, got)
		}
	}
	if err := s.UpdateFields("a", Update{}); err == nil {
		t.Error("an empty Update: no error")
	}
}
