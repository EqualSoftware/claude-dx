package spec

import (
	"os"
	"strings"
	"testing"
)

const legacyHistory = "\n## 9. Document History\n| Date | Author | Description / Change Notes |\n| :--- | :--- | :--- |\n" +
	"| 2026-10-26 | Tester | Initial draft specification |\n"

// TestWithAndWithoutHistorySection runs the tools on a spec that still has section 9 and on a spec that
// has none (ADR-018). The tools read both and never write to section 9.
func TestWithAndWithoutHistorySection(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		name := "without section 9"
		if legacy {
			name = "with section 9"
		}
		t.Run(name, func(t *testing.T) { checkHistorySection(t, legacy) })
	}
}

func appendHistory(t *testing.T, s Store) {
	t.Helper()
	f, err := os.OpenFile(s.path("a"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(legacyHistory); err != nil {
		t.Fatal(err)
	}
}

func checkHistorySection(t *testing.T, legacy bool) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{{"spec_create", createArgs("a", "A"), "Created a."}})
	if legacy {
		appendHistory(t, s)
	}
	steps := []step{
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Two."}, "Added R-2."},
		{"spec_req_set", map[string]any{"slug": "a", "rid": "R-1", "status": "Done"}, "Updated R-1."},
		{"spec_req_delete", map[string]any{"slug": "a", "rid": "R-2"}, "Deleted R-2."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "Three."}, "Added R-2."},
		{"spec_update_fields", map[string]any{"slug": "a", "priority": "P0"}, "Updated a."},
	}
	for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
		steps = append(steps, step{"spec_update_section", map[string]any{"slug": "a", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of a."})
	}
	steps = append(steps,
		step{"spec_set_status", map[string]any{"slug": "a", "status": "In Review"}, "a is In Review."},
		step{"spec_set_status", map[string]any{"slug": "a", "status": "Approved"}, "a is Approved."},
	)
	run(t, session, steps)
	got := readSpec(t, s, "a")
	for _, want := range []string{"status: Approved", "priority: P0", "| R-1 | One. | Done |", "| R-2 | Three. | Planned |"} {
		if !strings.Contains(got, want) {
			t.Errorf("spec has no %q:\n%s", want, got)
		}
	}
	if legacy != strings.HasSuffix(got, legacyHistory) {
		t.Errorf("section 9 is wrong (legacy %v):\n%s", legacy, got)
	}
	if text, err := s.Get("a", "Document History"); legacy != (err == nil) || legacy && !strings.Contains(text, "Initial draft") {
		t.Errorf("read of section 9: %q, %v", text, err)
	}
}

// TestReviewRules checks which changes set an Approved spec to In Review (docs-spec#R-11).
func TestReviewRules(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	steps := []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
	}
	for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
		steps = append(steps, step{"spec_update_section", map[string]any{"slug": "a", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of a."})
	}
	steps = append(steps,
		step{"spec_set_status", map[string]any{"slug": "a", "status": "In Review"}, "a is In Review."},
		step{"spec_set_status", map[string]any{"slug": "a", "status": "Approved"}, "a is Approved."},
		step{"spec_question_add", map[string]any{"slug": "a", "question": "Q?"}, "Added Q-1. Q-1 is Open, so a is In Review. The user must approve it again."},
	)
	run(t, session, steps)
	if got := readSpec(t, s, "a"); !strings.Contains(got, "status: In Review") {
		t.Errorf("a new question did not set In Review:\n%s", got)
	}
}
