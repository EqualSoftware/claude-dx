package spec

import (
	"os/exec"
	"strings"
	"testing"
)

// commitAll commits all files with message and returns the short hash.
func commitAll(t *testing.T, root, message string) string {
	t.Helper()
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "--allow-empty", "-m", message)
	out, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestCommits(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	get := map[string]any{"slug": "docs-lock", "section": "5"}
	run(t, session, []step{
		{"spec_create", createArgs("docs-lock", "Docs lock"), "Created docs-lock."},
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": "Write lock."}, "Added R-1."},
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": "Read lock."}, "Added R-2."},
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": "No lock."}, "Added R-3."},
		// No HEAD: no commits and no error.
		{"spec_get", get, "| ID | Requirement | Status | Commits |\n| :--- | :--- | :--- | :--- |\n" +
			"| R-1 | Write lock. | Planned | — |\n| R-2 | Read lock. | Planned | — |\n| R-3 | No lock. | Planned | — |"},
	})
	c1 := commitAll(t, s.Root, "feat: one\n\nRefs: docs-lock#R-1\nRefs: docs-lock#R-2")
	c2 := commitAll(t, s.Root, "fix: two\n\nRefs: docs-lock#R-2, other#R-3, docs-lock#bad")
	commitAll(t, s.Root, "chore: three\n\nRefs: other#R-1")
	commitAll(t, s.Root, "chore: four")
	run(t, session, []step{
		{"spec_get", get, "| ID | Requirement | Status | Commits |\n| :--- | :--- | :--- | :--- |\n" +
			"| R-1 | Write lock. | Planned | " + c1 + " |\n| R-2 | Read lock. | Planned | " + c2 + ", " + c1 + " |\n| R-3 | No lock. | Planned | — |\n\n" +
			"Commits:\n- " + c2 + " fix: two\n- " + c1 + " feat: one"},
		{"spec_get", map[string]any{"slug": "docs-lock", "section": "Goals"}, "- G-1: [Measurable result.]"},
	})
	full, err := s.getText(getInput{Slug: "docs-lock"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(full, "| R-1 | Write lock. | Planned | "+c1+" |") || !strings.Contains(full, "\n\nCommits:\n- "+c2+" fix: two\n") {
		t.Errorf("full spec has no commits:\n%s", full)
	}
	if text, _ := s.Get("docs-lock", "5"); strings.Contains(text, "Commits") {
		t.Errorf("SPEC.md has the commits:\n%s", text)
	}
}

func TestCommitsGitFailure(t *testing.T) {
	s := newStore(t)
	if err := s.Create(New{Slug: "a", Title: "A", Milestone: "v0.1.0", Priority: "P1"}); err != nil {
		t.Fatal(err)
	}
	s.Root = t.TempDir() // not a git repository
	_, err := s.getText(getInput{Slug: "a"})
	if err == nil || !strings.HasPrefix(err.Error(), "git log failed: ") {
		t.Fatalf("getText = %v, want a git log failure", err)
	}
}
