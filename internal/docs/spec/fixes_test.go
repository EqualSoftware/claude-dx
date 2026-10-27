package spec

import (
	"os"
	"strings"
	"testing"
)

func TestRelatedIsCanonical(t *testing.T) {
	s := newStore(t)
	args := createArgs("a", "A")
	args["related"] = []string{"8"}
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", args, "Created a."},
		{"spec_update_fields", map[string]any{"slug": "a", "related": []string{"adr-8"}}, "Updated a."},
	})
	if text, _ := s.Get("a", ""); !strings.Contains(text, "related: [ADR-008]\n") {
		t.Errorf("related is not canonical:\n%s", text)
	}
	run(t, session, []step{
		{"spec_create", func() map[string]any { a := createArgs("b", "B"); a["related"] = []string{"9"}; return a }(),
			`ERR Related "9": ADR-009 does not exist.`},
	})
}

func TestQuestionTexts(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_question_add", map[string]any{"slug": "a", "question": "  "}, "ERR Give the question."},
		{"spec_question_add", map[string]any{"slug": "a", "question": "Why?"}, "Added Q-1."},
		{"spec_question_answer", map[string]any{"slug": "a", "qid": "Q-1", "answer": " "}, "ERR Give the answer."},
		{"spec_question_answer", map[string]any{"slug": "a", "qid": "Q-1", "answer": "open"}, `ERR The answer "Open" means that the question has no answer. Write the answer.`},
	})
}

func TestAreaLineBreak(t *testing.T) {
	s := newStore(t)
	args := createArgs("a", "A")
	args["area"] = []string{"x\ny"}
	run(t, connect(t, s), []step{{"spec_create", args, "ERR The area has a line break. Write the area on one line."}})
	run(t, connect(t, s), []step{{"spec_create", func() map[string]any { a := createArgs("a", "A\nB"); return a }(),
		"ERR The title has a line break. Write the title on one line."}})
}

func TestSectionContentChecks(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	run(t, session, []step{
		{"spec_create", createArgs("a", "A"), "Created a."},
		{"spec_update_section", map[string]any{"slug": "a", "section": "1", "body": "```\nx"},
			`ERR The content of section "1" has a code fence that is not closed. Close it with the same fence.`},
	})
}

func TestGetOutsideGit(t *testing.T) {
	s := newStore(t)
	if err := s.Create(New{Slug: "a", Title: "A", Milestone: "v0.1.0", Priority: "P1"}); err != nil {
		t.Fatal(err)
	}
	s.Root = t.TempDir() // not a repository
	text, err := s.getText(getInput{Slug: "a", Section: "5"})
	if err != nil || !strings.Contains(text, "Warning: commits: ") || strings.Contains(text, "| Commits |") {
		t.Errorf("getText = %q, %v", text, err)
	}
}

func TestDeleteOutsideGit(t *testing.T) {
	s := newStore(t)
	if err := s.Create(New{Slug: "a", Title: "A", Milestone: "v0.1.0", Priority: "P1"}); err != nil {
		t.Fatal(err)
	}
	s.Root = t.TempDir() // not a repository: no uncommitted-change check
	if err := s.Delete("a", "test", "", false); err != nil {
		t.Fatalf("Delete = %v", err)
	}
	if _, err := os.Stat(s.Dir + "/a"); err == nil {
		t.Error("the spec folder stays")
	}
}
