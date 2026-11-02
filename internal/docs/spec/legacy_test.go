package spec

import (
	"fmt"
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

// TestReviewRules checks which changes set an Approved spec to In Review (docs-spec#R-11), and the
// question tools, on a spec with and without section 9.
func TestReviewRules(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy %v", legacy), func(t *testing.T) { checkReviewRules(t, legacy) })
	}
}

func checkReviewRules(t *testing.T, legacy bool) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{{"spec_create", createArgs("a", "A"), "Created a."}})
	if legacy {
		appendHistory(t, s)
	}
	approve := []step{
		{"spec_req_add", map[string]any{"slug": "a", "text": "One."}, "Added R-1."},
		{"spec_set_status", map[string]any{"slug": "a", "status": "In Review"}, "a is In Review."},
		{"spec_set_status", map[string]any{"slug": "a", "status": "Approved"}, "a is Approved."},
	}
	fill := func() {
		for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
			run(t, session, []step{{"spec_update_section", map[string]any{"slug": "a", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of a."}})
		}
	}
	review := " The scope changed, so a is In Review. The user must approve it again."
	fill()
	run(t, session, approve)
	run(t, session, []step{{"spec_update_section", map[string]any{"slug": "a", "section": "Goals", "body": "- G-1: New."}, "Updated section Goals of a." + review}})
	run(t, session, approve[2:])
	run(t, session, []step{{"spec_update_section", map[string]any{"slug": "a", "section": "Out of Scope", "body": "- New."}, "Updated section Out of Scope of a." + review}})
	run(t, session, approve[2:])
	run(t, session, []step{
		{"spec_question_add", map[string]any{"slug": "a", "question": "Q?"}, "Added Q-1. Q-1 is Open, so a is In Review. The user must approve it again."},
	})
	if got := readSpec(t, s, "a"); !strings.Contains(got, "status: In Review") || !strings.HasSuffix(got, "\n") {
		t.Errorf("after the question add:\n%s", got)
	}
	run(t, session, []step{{"spec_question_answer", map[string]any{"slug": "a", "qid": "Q-1", "answer": "Yes."}, "Answered Q-1."}})
	got := readSpec(t, s, "a")
	if !strings.HasSuffix(got, "\n") || !strings.Contains(got, "| Q-1 | Q? | Tester | Yes. |") {
		t.Errorf("after the answer:\n%s", got)
	}
	if legacy != strings.HasSuffix(got, legacyHistory) {
		t.Errorf("section 9 is wrong (legacy %v):\n%s", legacy, got)
	}
}
