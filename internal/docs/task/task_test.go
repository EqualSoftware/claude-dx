package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/numbering"
)

func newStore(t *testing.T) Store {
	t.Helper()
	return Store{
		Dir:    filepath.Join(t.TempDir(), "docs", "tasks"),
		Alloc:  numbering.Allocator{Dir: t.TempDir()},
		Now:    func() time.Time { return time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC) },
		Author: func() string { return "Tester" },
	}
}

// must stops the test if err is not nil.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func create(t *testing.T, s Store, title string) {
	t.Helper()
	if _, err := s.Create(New{Title: title, Priority: "P1", Area: []string{"cli"}}); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, s Store, id string) string {
	t.Helper()
	text, err := s.Get(id, "")
	if err != nil {
		t.Fatal(err)
	}
	return text
}

func TestCreate(t *testing.T) {
	s := newStore(t)
	path, err := s.Create(New{
		Title: "Add cache: fast", Priority: "P1", Area: []string{"cli", "docs"}, DependsOn: []string{"TASK-009"},
		Fields:   Fields{Type: "`feat`", Size: "`S` (< 1 day)"},
		Sections: map[string]string{"1": "The summary.", "Goals": "- G-1: Fast."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "TASK-001-add-cache-fast.md" {
		t.Fatalf("path = %q", path)
	}
	doc := read(t, s, "TASK-001")
	for _, want := range []string{
		"id: TASK-001\n", "title: 'Add cache: fast'", "status: Draft  # Draft", "priority: P1  # P0",
		"area: [cli, docs]", "depends_on: [TASK-009]", "# TASK-001: Add cache: fast\n",
		"| Type | `feat` |", "| Size | `S` (< 1 day) |", "| Owner | [Name or role] |",
		"| 2026-10-25 | Tester | Initial draft specification |",
		"## 1. Summary\n\nThe summary.\n", "## 3. Goals\n\n- G-1: Fast.\n",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("task does not contain %q", want)
		}
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "INDEX.md")); err != nil {
		t.Error("INDEX.md was not made")
	}
}

func TestCreateRefusesBadInput(t *testing.T) {
	s := newStore(t)
	bad := []New{
		{Title: "", Priority: "P1"},
		{Title: "X", Priority: "P9"},
		{Title: "X", Priority: "P1", Sections: map[string]string{"99": "x"}},
	}
	for _, in := range bad {
		if _, err := s.Create(in); err == nil {
			t.Errorf("Create(%+v) did not fail", in)
		}
	}
	if _, err := os.Stat(s.Dir); err == nil {
		t.Error("a refused Create made docs/tasks")
	}
}

func TestMissingDir(t *testing.T) {
	s := newStore(t)
	if _, _, err := s.List(Filter{}); !errors.Is(err, ErrNoDir) {
		t.Fatalf("List = %v, want ErrNoDir", err)
	}
	if err := s.SetStatus("TASK-001", "Done", ""); !errors.Is(err, ErrNoDir) {
		t.Fatalf("SetStatus = %v, want ErrNoDir", err)
	}
	if _, err := os.Stat(s.Dir); err == nil {
		t.Fatal("a read or update made docs/tasks")
	}
}

func TestListFilter(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	create(t, s, "Two")
	if err := s.SetStatus("TASK-002", "Cancelled", "Not necessary."); err != nil {
		t.Fatal(err)
	}
	open, _, _ := s.List(Filter{Status: "open"})
	all, _, _ := s.List(Filter{Status: "all"})
	cancelled, _, _ := s.List(Filter{Status: "Cancelled", Area: "cli"})
	if len(open) != 1 || len(all) != 2 || len(cancelled) != 1 || cancelled[0].Closed != "2026-10-25" {
		t.Fatalf("open %d, all %d, cancelled %+v", len(open), len(all), cancelled)
	}
}

func TestSetStatusRules(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	tests := []struct {
		status, note, want string
	}{
		{"In Review", "", "section 1 has the template text"},
		{"In Review", "", "Size must be S or M"},
		{"Blocked", "", "give a note"},
		{"Done", "", "Definition of Done item 1 is not checked"},
		{"Finished", "", "is not valid"},
	}
	for _, tt := range tests {
		err := s.SetStatus("TASK-001", tt.status, tt.note)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("SetStatus(%s) = %v, want %q", tt.status, err, tt.want)
		}
	}
	if err := s.SetStatus("TASK-001", "In Development", ""); err != nil {
		t.Fatal(err)
	}
	doc := read(t, s, "TASK-001")
	if !strings.Contains(doc, "status: In Development  # Draft") || !strings.Contains(doc, "| 2026-10-25 | Tester | Status: In Development. |") {
		t.Fatalf("status change not written:\n%s", doc)
	}
}

func TestInReviewWhenComplete(t *testing.T) {
	s := newStore(t)
	sections := map[string]string{}
	for _, ref := range leafSections {
		sections[ref] = "Filled " + ref + "."
	}
	if _, err := s.Create(New{Title: "Full", Priority: "P1", Area: []string{"cli"}, Fields: Fields{Size: "`M` (1–3 days)"}, Sections: sections}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetStatus("TASK-001", "In Review", ""); err != nil {
		t.Fatalf("SetStatus(In Review) = %v", err)
	}
}

func TestQuestions(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	for _, want := range []string{"Q-1", "Q-2"} {
		qid, err := s.AddQuestion("TASK-001", "Which | cache?", "")
		if err != nil || qid != want {
			t.Fatalf("AddQuestion = %q, %v; want %s", qid, err, want)
		}
	}
	if err := s.SetStatus("TASK-001", "Approved", ""); err == nil || !strings.Contains(err.Error(), "Q-1 is Open; Q-2 is Open") {
		t.Fatalf("Approved with open questions = %v", err)
	}
	must(t, s.AnswerQuestion("TASK-001", "Q-1", "SQLite."))
	must(t, s.AnswerQuestion("TASK-001", "q-2", "SQLite."))
	must(t, s.SetStatus("TASK-001", "Approved", ""))
	table, _ := s.Get("TASK-001", "13")
	if strings.Contains(table, "[Question]") || !strings.Contains(table, `| Q-1 | Which \| cache? | Maintainer | SQLite. |`) {
		t.Fatalf("questions table =\n%s", table)
	}
	if err := s.AnswerQuestion("TASK-001", "Q-9", "x"); err == nil {
		t.Error("answer to a missing question did not fail")
	}
}

func TestCheckDoD(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	for item := 1; item <= 6; item++ {
		if err := s.CheckDoD("TASK-001", item, "CI run 1."); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CheckDoD("TASK-001", 1, "again"); err == nil {
		t.Error("checking an item two times did not fail")
	}
	if err := s.CheckDoD("TASK-001", 7, "x"); err == nil {
		t.Error("checking a missing item did not fail")
	}
	if err := s.SetStatus("TASK-001", "Done", ""); err != nil {
		t.Fatalf("Done with all items checked = %v", err)
	}
}

func TestUpdateFields(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	u := Update{Title: "New title", Priority: "P0", DependsOn: []string{}, Fields: Fields{Owner: "Maintainer"}}
	if err := s.UpdateFields("TASK-001", u); err != nil {
		t.Fatal(err)
	}
	doc := read(t, s, "TASK-001")
	for _, want := range []string{"title: New title", "# TASK-001: New title\n", "priority: P0", "depends_on: []", "area: [cli]", "| Owner | Maintainer |"} {
		if !strings.Contains(doc, want) {
			t.Errorf("task does not contain %q", want)
		}
	}
	if err := s.UpdateFields("TASK-001", Update{Priority: "P5"}); err == nil {
		t.Error("bad priority did not fail")
	}
}

func TestRenumber(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	create(t, s, "Two")
	// A teammate's branch used number 1 too.
	dup := strings.ReplaceAll(read(t, s, "TASK-002"), "TASK-002", "TASK-001")
	must(t, os.WriteFile(filepath.Join(s.Dir, "TASK-001-dup.md"), []byte(dup), 0o644))
	must(t, s.UpdateFields("TASK-002", Update{DependsOn: []string{"TASK-001"}}))
	newID, refs, err := s.Renumber("TASK-001-dup.md")
	if err != nil || newID != "TASK-003" {
		t.Fatalf("Renumber = %q, %v", newID, err)
	}
	if len(refs) != 2 || refs[0] != "TASK-001-one.md" || refs[1] != "TASK-002-two.md" {
		t.Fatalf("refs = %v", refs)
	}
	doc := read(t, s, "TASK-003")
	if !strings.Contains(doc, "id: TASK-003") || !strings.Contains(doc, "# TASK-003: Two") {
		t.Fatalf("renumbered task =\n%s", doc)
	}
	_, err = s.Get("TASK-001", "")
	must(t, err)
}

// TestIndexMatchesRepository makes INDEX.md for the tasks of this repository and compares it with
// the committed INDEX.md. The committed file came from the equal-tasks-index script, so the format is the same.
func TestIndexMatchesRepository(t *testing.T) {
	repo := filepath.Join("..", "..", "..", "docs", "tasks")
	want, err := os.ReadFile(filepath.Join(repo, "INDEX.md"))
	if err != nil {
		t.Skip("no repository tasks")
	}
	s := Store{Dir: t.TempDir()}
	files, _ := filepath.Glob(filepath.Join(repo, "TASK-*.md"))
	for _, f := range files {
		data, _ := os.ReadFile(f)
		if err := os.WriteFile(filepath.Join(s.Dir, filepath.Base(f)), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.WriteIndex(); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(s.Dir, "INDEX.md"))
	if string(got) != string(want) {
		t.Fatalf("INDEX.md differs from the committed file.\ngot:\n%s", got)
	}
}

// BenchmarkList200 measures NFR-1: task_list on 200 tasks takes less than 50 ms.
func BenchmarkList200(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "tasks", "TASK-006-task-tools-index-and-skill.md"))
	if err != nil {
		b.Skip("no repository tasks")
	}
	s := Store{Dir: b.TempDir()}
	for i := 1; i <= 200; i++ {
		name := fmt.Sprintf("TASK-%03d-task.md", i)
		text := strings.Replace(string(data), "id: TASK-006", fmt.Sprintf("id: TASK-%03d", i), 1)
		if err := os.WriteFile(filepath.Join(s.Dir, name), []byte(text), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	for b.Loop() {
		if _, err := s.listText(listInput{Status: "all"}); err != nil {
			b.Fatal(err)
		}
	}
}
