package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
	"github.com/EqualSoftware/claude-dx/internal/markdown"
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
	if _, _, err := s.Create(New{Title: title, Priority: "P1", Milestone: "v0.1.0", Type: "feat", Area: []string{"cli"}}); err != nil {
		t.Fatal(err)
	}
}

// setStatus calls SetStatus and drops the warnings.
func setStatus(s Store, id, status, note string) error {
	_, err := s.SetStatus(id, status, note)
	return err
}

func update(t *testing.T, s Store, id string, u Update) {
	t.Helper()
	if _, _, err := s.UpdateFields(id, u); err != nil {
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
	path, _, err := s.Create(New{
		Title: "Add cache: fast", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Area: []string{"cli", "docs"},
		Size: "S", Platforms: []string{"macOS", "Windows (amd64, arm64)"}, Related: []string{"ADR-013"},
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
		"id: TASK-001\ntitle: 'Add cache: fast'\nstatus: Draft\npriority: P1\nmilestone: v0.1.0\ntype: feat\nsize: S\nowner: Tester\n" +
			"platforms: [macOS, 'Windows (amd64, arm64)']\narea: [cli, docs]\ndepends_on: []\nrelated: [ADR-013]\n---\n",
		"# TASK-001: Add cache: fast\n\n## Document History\n",
		"| 2026-10-25 | Tester | Initial draft specification |",
		"## 1. Summary\n\nThe summary.\n", "## 3. Goals\n\n- G-1: Fast.\n",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("task does not contain %q", want)
		}
	}
	if strings.Contains(doc, "eureka") {
		t.Error("AC-1 of TASK-026: the Definition of Done has the /dx:eureka item")
	}
	if strings.Contains(doc, " # ") || strings.Contains(doc, "| Field | Value |") {
		t.Error("task has a frontmatter comment or the Field/Value table")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "README.md")); err != nil {
		t.Error("README.md was not made")
	}
}

func TestCreateRefusesBadInput(t *testing.T) {
	s := newStore(t)
	bad := []New{
		{Title: "", Priority: "P1", Milestone: "v0.1.0", Type: "feat"},
		{Title: "X", Priority: "P9", Milestone: "v0.1.0", Type: "feat"},
		{Title: "X", Priority: "P1", Milestone: "v0.1.0"},
		{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feature"},
		{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Size: "L"},
		{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Sections: map[string]string{"99": "x"}},
		// AC-1 of TASK-030: no milestone, or a milestone that is not vMAJOR.MINOR.PATCH.
		{Title: "X", Priority: "P1", Type: "feat"},
		{Title: "X", Priority: "P1", Milestone: "0.1", Type: "feat"},
		// AC-4 of TASK-053: a section and its child section, or two refs of one section.
		{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Sections: map[string]string{"7": "x", "7.1": "y"}},
		{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Sections: map[string]string{"7": "x", "Design": "y"}},
	}
	for _, in := range bad {
		if _, _, err := s.Create(in); err == nil {
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
	if err := setStatus(s, "TASK-001", "Done", ""); !errors.Is(err, ErrNoDir) {
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
	if err := setStatus(s, "TASK-002", "Cancelled", "Not necessary."); err != nil {
		t.Fatal(err)
	}
	open, _, _ := s.List(Filter{"status": "open"})
	all, _, _ := s.List(Filter{"status": "all"})
	cancelled, _, _ := s.List(Filter{"status": "cancelled", "area": "cli"})
	if len(open) != 1 || len(all) != 2 || len(cancelled) != 1 || cancelled[0].Closed != "2026-10-25" {
		t.Fatalf("open %d, all %d, cancelled %+v", len(open), len(all), cancelled)
	}
}

func TestListFilterAnyField(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	create(t, s, "Two")
	update(t, s, "TASK-002", Update{DependsOn: []string{"TASK-001"}, Platforms: []string{"Windows"}})
	tests := []struct {
		filter Filter
		want   int
	}{
		{Filter{"depends_on": "TASK-001"}, 1},
		{Filter{"platforms": "windows"}, 1},
		{Filter{"owner": "tester", "type": "feat"}, 2},
		{Filter{"size": "M"}, 0},
		{Filter{"status": "all", "priority": ""}, 2},
	}
	for _, tt := range tests {
		got, _, err := s.List(tt.filter)
		if err != nil || len(got) != tt.want {
			t.Errorf("List(%v) = %d tasks, %v; want %d", tt.filter, len(got), err, tt.want)
		}
	}
}

func TestSetStatusRules(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	tests := []struct {
		from, status, want string
	}{
		{"Draft", "In Review", "section 1 has the template text"},
		{"Draft", "In Review", "give a size"},
		{"Draft", "In Review", "give the platforms"},
		{"Approved", "Blocked", "give a note"},
		{"In Development", "Done", "Definition of Done items 1, 2, 3, 4, 5 are not checked"},
		{"Draft", "Finished", "is not valid"},
	}
	for _, tt := range tests {
		force(t, s, "TASK-001", tt.from)
		err := setStatus(s, "TASK-001", tt.status, "")
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("SetStatus(%s) = %v, want %q", tt.status, err, tt.want)
		}
	}
	force(t, s, "TASK-001", "Approved")
	if err := setStatus(s, "TASK-001", "In Development", ""); err != nil {
		t.Fatal(err)
	}
	doc := read(t, s, "TASK-001")
	if !strings.Contains(doc, "status: In Development\n") || !strings.Contains(doc, "| 2026-10-25 | Tester | Status: In Development. |") {
		t.Fatalf("status change not written:\n%s", doc)
	}
}

// TestInReviewWhenComplete is also AC-8 of TASK-037: sections 13 and 14 keep the template text.
func TestInReviewWhenComplete(t *testing.T) {
	s := newStore(t)
	sections := map[string]string{}
	for _, ref := range leafSections {
		sections[ref] = "Filled " + ref + "."
	}
	if _, _, err := s.Create(New{Title: "Full", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Area: []string{"cli"}, Size: "M", Platforms: []string{"Linux"}, Sections: sections}); err != nil {
		t.Fatal(err)
	}
	if err := setStatus(s, "TASK-001", "In Review", ""); err != nil {
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
	force(t, s, "TASK-001", "In Review")
	if err := setStatus(s, "TASK-001", "Approved", ""); err == nil || !strings.Contains(err.Error(), "Q-1 is Open; Q-2 is Open") {
		t.Fatalf("Approved with open questions = %v", err)
	}
	must(t, s.AnswerQuestion("TASK-001", "Q-1", "SQLite."))
	must(t, s.AnswerQuestion("TASK-001", "q-2", "SQLite."))
	must(t, setStatus(s, "TASK-001", "Approved", ""))
	table, _ := s.Get("TASK-001", "13")
	if strings.Contains(table, "None.") || !strings.Contains(table, `| Q-1 | Which \| cache? | Maintainer | SQLite. |`) {
		t.Fatalf("questions table =\n%s", table)
	}
	if err := s.AnswerQuestion("TASK-001", "Q-9", "x"); err == nil {
		t.Error("answer to a missing question did not fail")
	}
}

func TestKeepRealQuestions(t *testing.T) {
	for _, first := range []string{"[CLI] keep the -p flag?", "None."} {
		s := newStore(t)
		create(t, s, "One")
		if _, err := s.AddQuestion("TASK-001", first, ""); err != nil {
			t.Fatal(err)
		}
		qid, err := s.AddQuestion("TASK-001", "Second?", "")
		if err != nil || qid != "Q-2" {
			t.Fatalf("second AddQuestion = %q, %v; want Q-2", qid, err)
		}
		table, _ := s.Get("TASK-001", "13")
		if !strings.Contains(table, "| Q-1 | "+first+" |") || !strings.Contains(table, "| Q-2 | Second? |") {
			t.Errorf("questions table =\n%s", table)
		}
	}
}

func TestWithoutPlaceholders(t *testing.T) {
	lines := []string{noQuestions, "| Q-1 | [Question] | Maintainer | Open |", "| Q-2 | [CLI] real? | Maintainer | Open |"}
	got, next := withoutPlaceholders(lines)
	if !slices.Equal(got, lines[2:]) || next != 3 {
		t.Errorf("withoutPlaceholders = %q, %d", got, next)
	}
}

func TestCheckDoD(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	for item := 1; item <= 5; item++ {
		if err := s.CheckDoD("TASK-001", item, "CI run 1."); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CheckDoD("TASK-001", 1, "again"); err == nil {
		t.Error("checking an item two times did not fail")
	}
	if err := s.CheckDoD("TASK-001", 6, "x"); err == nil {
		t.Error("checking a missing item did not fail")
	}
	force(t, s, "TASK-001", "In Development")
	if err := setStatus(s, "TASK-001", "Done", ""); err != nil {
		t.Fatalf("Done with all items checked = %v", err)
	}
}

// TestListMilestone is AC-2 of TASK-030.
func TestListMilestone(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	create(t, s, "Two")
	update(t, s, "TASK-002", Update{Milestone: "v0.2.0"})
	got, _, _ := s.List(Filter{"milestone": "v0.1.0"})
	if len(got) != 1 || got[0].Front.ID != "TASK-001" {
		t.Fatalf("List(milestone: v0.1.0) = %+v", got)
	}
}

// TestListReady is AC-3 and AC-4 of TASK-030.
func TestListReady(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Dep", "A", "B", "C", "Later", "Done dep"} {
		create(t, s, title)
	}
	for _, id := range []string{"TASK-002", "TASK-003", "TASK-005"} {
		force(t, s, id, "Approved")
	}
	update(t, s, "TASK-002", Update{DependsOn: []string{"TASK-001"}})
	update(t, s, "TASK-006", Update{Milestone: "v0.10.0"})
	update(t, s, "TASK-003", Update{Milestone: "v0.10.0", DependsOn: []string{"TASK-006"}})
	update(t, s, "TASK-005", Update{Milestone: "v0.2.0", Priority: "P3"})
	force(t, s, "TASK-006", "Done")
	got, _, err := s.List(Filter{"status": "ready"})
	must(t, err)
	var ids []string
	for _, task := range got {
		ids = append(ids, task.Front.ID)
	}
	if strings.Join(ids, " ") != "TASK-005 TASK-003" {
		t.Fatalf("ready = %v, want TASK-005 TASK-003", ids)
	}
}

// TestListDepsStatus is AC-1 to AC-3 of TASK-048.
func TestListDepsStatus(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Done dep", "Open dep", "Gone dep", "Main"} {
		create(t, s, title)
	}
	update(t, s, "TASK-004", Update{DependsOn: []string{"TASK-001", "TASK-002", "TASK-003"}})
	force(t, s, "TASK-001", "Done")
	gone, _ := filepath.Glob(filepath.Join(s.Dir, "TASK-003-*.md"))
	must(t, os.Remove(gone[0]))
	got, err := s.listText(listInput{ID: "TASK-004"})
	must(t, err)
	if want := "deps:TASK-001, TASK-002 (Draft), TASK-003 (missing)"; !strings.Contains(got, want) {
		t.Fatalf("task_list = %q, want %q", got, want)
	}
}

// force sets the status with no status rules.
func force(t *testing.T, s Store, id, status string) {
	t.Helper()
	must(t, s.edit(id, func(d *doc) error {
		d.front = markdown.SetField(d.front, "status", status)
		return nil
	}))
}

func TestUpdateFields(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	u := Update{Title: "New title", Priority: "P0", Milestone: "v0.2.0", Type: "fix", Size: "M", Owner: "Equal", DependsOn: []string{}, Related: []string{"ADR-001"}}
	if _, _, err := s.UpdateFields("TASK-001", u); err != nil {
		t.Fatal(err)
	}
	doc := read(t, s, "TASK-001")
	for _, want := range []string{
		"title: New title\nstatus: Draft\npriority: P0\nmilestone: v0.2.0\ntype: fix\nsize: M\nowner: Equal\nplatforms: []\narea: [cli]\ndepends_on: []\nrelated: [ADR-001]\n",
		"# TASK-001: New title\n",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("task does not contain %q", want)
		}
	}
	for _, bad := range []Update{{Priority: "P5"}, {Milestone: "v1"}, {Type: "feature"}, {Size: "L"}} {
		if _, _, err := s.UpdateFields("TASK-001", bad); err == nil {
			t.Errorf("UpdateFields(%+v) did not fail", bad)
		}
	}
}

func TestRenumber(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	create(t, s, "Two")
	// A teammate's branch used number 1 too.
	dup := strings.ReplaceAll(read(t, s, "TASK-002"), "TASK-002", "TASK-001")
	must(t, os.WriteFile(filepath.Join(s.Dir, "TASK-001-dup.md"), []byte(dup), 0o644))
	update(t, s, "TASK-002", Update{DependsOn: []string{"TASK-001"}})
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
	// AC-4 of TASK-049: the file moved, so one file has each number.
	if _, err := os.Stat(filepath.Join(s.Dir, "TASK-001-dup.md")); !os.IsNotExist(err) {
		t.Fatalf("the old file exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "TASK-003-dup.md")); err != nil {
		t.Fatalf("the new file does not exist: %v", err)
	}
}

// TestRenumberMoveFails is AC-1 of TASK-054.
func TestRenumberMoveFails(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	path := filepath.Join(s.Dir, "TASK-001-one.md")
	before, err := os.ReadFile(path)
	must(t, err)
	move = func(string, string) error { return os.ErrExist }
	t.Cleanup(func() { move = docfile.Move })
	if _, _, err := s.Renumber("TASK-001-one.md"); !errors.Is(err, os.ErrExist) {
		t.Fatalf("Renumber = %v; want the error of Move", err)
	}
	after, err := os.ReadFile(path)
	must(t, err)
	if string(after) != string(before) {
		t.Fatalf("the old file changed:\n%s", after)
	}
}

// TestRenumberBadFrontmatter is AC-2 of TASK-054.
func TestRenumberBadFrontmatter(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	path := filepath.Join(s.Dir, "TASK-001-one.md")
	bad := []byte("---\nid: TASK-001\ntitle: [One\n---\n\n# TASK-001: One\n")
	must(t, os.WriteFile(path, bad, 0o644))
	if _, _, err := s.Renumber("TASK-001-one.md"); err == nil || !strings.Contains(err.Error(), "bad frontmatter") {
		t.Fatalf("Renumber = %v; want a bad frontmatter error", err)
	}
	after, err := os.ReadFile(path)
	must(t, err)
	if string(after) != string(bad) {
		t.Fatalf("the file changed:\n%s", after)
	}
}

// TestIndexMatchesRepository makes README.md for the tasks of this repository and compares it with
// the committed README.md. The committed file came from the equal-tasks-index script, so the format is the same.
func TestIndexMatchesRepository(t *testing.T) {
	repo := filepath.Join("..", "..", "..", "docs", "tasks")
	want, err := os.ReadFile(filepath.Join(repo, "README.md"))
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
	got, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if string(got) != string(want) {
		t.Fatalf("README.md differs from the committed file.\ngot:\n%s", got)
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

// TestIndexMilestones is AC-5 of TASK-030: one open table for each milestone, then one closed table.
func TestIndexMilestones(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"One", "Two", "Three"} {
		create(t, s, title)
	}
	update(t, s, "TASK-001", Update{Milestone: "v0.2.0"})
	must(t, setStatus(s, "TASK-003", "Cancelled", "Not necessary."))
	data, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	text := string(data)
	v1, v2, closed := strings.Index(text, "## v0.1.0\n"), strings.Index(text, "## v0.2.0\n"), strings.Index(text, "## Closed\n")
	if v1 < 0 || v1 > v2 || v2 > closed || !strings.Contains(text[v1:v2], "TASK-002") || !strings.Contains(text[closed:], "TASK-003") {
		t.Fatalf("README.md =\n%s", text)
	}
}

// TestDependsOnRules is AC-1 to AC-6 of TASK-023.
func TestDependsOnRules(t *testing.T) {
	s := newStore(t)
	refused := func(err error, want string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %v, want %q", err, want)
		}
	}
	_, _, err := s.Create(New{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", DependsOn: []string{"TASK-099"}})
	refused(err, "TASK-099 does not exist")
	if _, err := os.Stat(s.Dir); err == nil {
		t.Fatal("AC-1: a refused Create made docs/tasks")
	}
	for _, title := range []string{"One", "Two", "Three", "Four"} {
		create(t, s, title)
	}
	update(t, s, "TASK-001", Update{DependsOn: []string{"TASK-002"}})
	before := read(t, s, "TASK-002")
	_, _, err = s.UpdateFields("TASK-002", Update{DependsOn: []string{"TASK-001"}})
	refused(err, "TASK-002 -> TASK-001 -> TASK-002")
	if read(t, s, "TASK-002") != before {
		t.Fatal("AC-2: a refused change changed the file")
	}
	_, _, err = s.UpdateFields("TASK-002", Update{DependsOn: []string{"TASK-002"}})
	refused(err, "itself")
	force(t, s, "TASK-003", "Done")
	force(t, s, "TASK-004", "Cancelled")
	for _, id := range []string{"TASK-003", "TASK-004"} {
		_, _, err = s.Create(New{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", DependsOn: []string{id}})
		refused(err, "Put it in related")
	}
	update(t, s, "TASK-001", Update{DependsOn: []string{}})
	update(t, s, "TASK-002", Update{Milestone: "v0.2.0"})
	_, _, err = s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002"}})
	refused(err, "later milestone v0.2.0")
	update(t, s, "TASK-002", Update{Milestone: "v0.1.0", Priority: "P3"})
	_, warnings, err := s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002"}})
	if err != nil || !strings.Contains(warnings, "TASK-002 has the lower priority P3") {
		t.Fatalf("AC-6: warnings = %q, %v", warnings, err)
	}
}

// TestOldDoneDependency is AC-1 to AC-3 of TASK-040.
func TestOldDoneDependency(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Main", "Old", "New", "Done"} {
		create(t, s, title)
	}
	update(t, s, "TASK-001", Update{DependsOn: []string{"TASK-002"}})
	force(t, s, "TASK-002", "Done")
	force(t, s, "TASK-004", "Done")
	if _, _, err := s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002", "TASK-003"}}); err != nil {
		t.Fatalf("AC-1: UpdateFields = %v", err)
	}
	_, _, err := s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002", "TASK-003", "TASK-004"}})
	if err == nil || !strings.Contains(err.Error(), "TASK-004 is Done") {
		t.Fatalf("AC-2: UpdateFields = %v", err)
	}
	_, _, err = s.UpdateFields("TASK-003", Update{DependsOn: []string{"TASK-001"}})
	if err == nil || !strings.Contains(err.Error(), "TASK-003 -> TASK-001 -> TASK-003") {
		t.Fatalf("AC-3: UpdateFields = %v", err)
	}
}

// TestRenameOnTitle is AC-1 to AC-3 of TASK-042.
func TestRenameOnTitle(t *testing.T) {
	s := newStore(t)
	create(t, s, "Old")
	file, _, err := s.UpdateFields("TASK-001", Update{Title: "New name"})
	if err != nil || file != "TASK-001-new-name.md" {
		t.Fatalf("AC-1: UpdateFields = %q, %v", file, err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "TASK-001-old.md")); !os.IsNotExist(err) {
		t.Fatalf("AC-1: the old file exists: %v", err)
	}
	index, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if !strings.Contains(string(index), "TASK-001-new-name.md") {
		t.Fatalf("AC-1: README.md =\n%s", index)
	}
	file, _, err = s.UpdateFields("TASK-001", Update{Title: "New Name"})
	if err != nil || file != "" {
		t.Fatalf("AC-2: UpdateFields = %q, %v", file, err)
	}
	must(t, os.WriteFile(filepath.Join(s.Dir, "TASK-001-taken.md"), nil, 0o644))
	before, _ := os.ReadFile(filepath.Join(s.Dir, "TASK-001-new-name.md"))
	if _, _, err := s.UpdateFields("TASK-001", Update{Title: "Taken"}); err == nil {
		t.Fatal("AC-3: UpdateFields to an existing file name did not fail")
	}
	after, _ := os.ReadFile(filepath.Join(s.Dir, "TASK-001-new-name.md"))
	if string(after) != string(before) {
		t.Fatal("AC-3: a refused change changed the file")
	}
}

// TestRenameExists is FR-2 of TASK-042 in edit: a new path that exists is refused.
func TestRenameExists(t *testing.T) {
	dir := t.TempDir()
	old, path := filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")
	must(t, os.WriteFile(old, []byte("a"), 0o644))
	must(t, os.WriteFile(path, []byte("b"), 0o644))
	if err := rename(old, path, []byte("c")); err == nil || !strings.Contains(err.Error(), "b.md exists") {
		t.Fatalf("rename = %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "b" {
		t.Fatalf("rename changed the existing file: %q", data)
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("rename removed the old file: %v", err)
	}
}

// TestParallelEdits is AC-1 of TASK-047: the docs lock keeps all edits of parallel writers.
func TestParallelEdits(t *testing.T) {
	s := newStore(t)
	s.Lock = filepath.Join(t.TempDir(), "docs.lock")
	create(t, s, "One")
	errs := make(chan error, 20)
	for i := range 20 {
		go func() {
			_, err := s.AddQuestion("TASK-001", fmt.Sprintf("Question %d?", i), "")
			errs <- err
		}()
	}
	for range 20 {
		must(t, <-errs)
	}
	text := read(t, s, "TASK-001")
	for i := range 20 {
		if !strings.Contains(text, fmt.Sprintf("Question %d?", i)) {
			t.Fatalf("question %d is lost:\n%s", i, text)
		}
	}
}

// TestUncheckedDoDMessage is AC-7 of TASK-023.
func TestUncheckedDoDMessage(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	must(t, s.CheckDoD("TASK-001", 1, "Evidence."))
	force(t, s, "TASK-001", "In Development")
	err := setStatus(s, "TASK-001", "Done", "")
	if err == nil || !strings.HasSuffix(err.Error(), ": Definition of Done items 2, 3, 4, 5 are not checked") {
		t.Fatalf("SetStatus(Done) = %v", err)
	}
}

// TestTitleRules is AC-1 and AC-3 of TASK-028, and the duplicate title check.
func TestTitleRules(t *testing.T) {
	s := newStore(t)
	_, _, err := s.Create(New{Title: strings.Repeat("x", 41), Priority: "P1", Milestone: "v0.1.0", Type: "feat"})
	if err == nil || !strings.Contains(err.Error(), "41 characters") {
		t.Fatalf("41 characters = %v", err)
	}
	if _, err := os.Stat(s.Dir); err == nil {
		t.Fatal("AC-1: a refused Create made docs/tasks")
	}
	create(t, s, "One")
	update(t, s, "TASK-001", Update{Title: strings.Repeat("y", 40)})
	_, _, err = s.Create(New{Title: strings.Repeat("Y", 40), Priority: "P1", Milestone: "v0.1.0", Type: "feat"})
	if err == nil || !strings.Contains(err.Error(), "TASK-001") {
		t.Fatalf("duplicate title = %v", err)
	}
}

// TestOverlapWarning is AC-1 to AC-3 of TASK-031.
func TestOverlapWarning(t *testing.T) {
	s := newStore(t)
	plan := func(path string) map[string]string {
		return map[string]string{"8": "| Work batch | Files | Steps | Check |\n| :--- | :--- | :--- | :--- |\n| 1 | `" + path + "` | X. | Y. |"}
	}
	for i, p := range []string{"internal/x.go", "internal/x.go", "internal/"} {
		_, _, err := s.Create(New{Title: fmt.Sprint("Task ", i), Priority: "P1", Milestone: "v0.1.0", Type: "feat", Sections: plan(p)})
		must(t, err)
	}
	force(t, s, "TASK-001", "In Review")
	warnings, err := s.SetStatus("TASK-001", "Approved", "")
	if err != nil || !strings.Contains(warnings, "Warning: TASK-002 also changes internal/x.go.") {
		t.Fatalf("AC-1: warnings = %q, %v", warnings, err)
	}
	update(t, s, "TASK-002", Update{DependsOn: []string{"TASK-001"}})
	force(t, s, "TASK-002", "In Review")
	warnings, err = s.SetStatus("TASK-002", "Approved", "")
	if err != nil || strings.Contains(warnings, "TASK-001") {
		t.Fatalf("AC-2: warnings = %q, %v", warnings, err)
	}
	if !strings.Contains(warnings, "Warning: TASK-003 also changes internal/x.go.") {
		t.Fatalf("AC-3: a directory does not overlap its file: %q", warnings)
	}
}

// TestScopeChangeNeedsReview is AC-1 to AC-3 of TASK-032.
func TestScopeChangeNeedsReview(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	force(t, s, "TASK-001", "Approved")
	review, err := s.UpdateSection("TASK-001", "5.1", "The current state.")
	if err != nil || review || !strings.Contains(read(t, s, "TASK-001"), "status: Approved\n") {
		t.Fatalf("AC-2: section 5.1 = %v, %v", review, err)
	}
	update(t, s, "TASK-001", Update{Priority: "P0"})
	if !strings.Contains(read(t, s, "TASK-001"), "status: Approved\n") {
		t.Fatal("AC-3: task_update_fields changed the status")
	}
	review, err = s.UpdateSection("TASK-001", "6.1", "- FR-1: New.")
	doc := read(t, s, "TASK-001")
	if err != nil || !review || !strings.Contains(doc, "status: In Review\n") || !strings.Contains(doc, "Status: In Review. The scope changed in section 6.1.") {
		t.Fatalf("AC-1: section 6.1 = %v, %v\n%s", review, err, doc)
	}
}

// TestStatusChanges is the table of FR-1 of TASK-034, and AC-1 to AC-3.
func TestStatusChanges(t *testing.T) {
	allowed := map[string]string{
		"Draft":          "In Review, Cancelled",
		"In Review":      "Approved, Draft, Cancelled",
		"Approved":       "In Development, In Review, Blocked, Cancelled",
		"In Development": "Done, Blocked, Cancelled",
		"Blocked":        "Approved, In Development, Cancelled",
		"Cancelled":      "Draft",
		"Done":           "",
	}
	for from, list := range allowed {
		for _, to := range Statuses {
			err := allowedChange(from, to)
			if want := slices.Contains(strings.Split(list, ", "), to); (err == nil) != want {
				t.Errorf("%s -> %s: %v, want allowed %v", from, to, err, want)
			}
		}
	}
	s := newStore(t)
	create(t, s, "One")
	err := setStatus(s, "TASK-001", "Approved", "")
	if err == nil || !strings.Contains(err.Error(), "cannot set Approved from Draft. Allowed: In Review, Cancelled") {
		t.Fatalf("AC-1: Draft -> Approved = %v", err)
	}
	force(t, s, "TASK-001", "Done")
	if err := setStatus(s, "TASK-001", "In Development", ""); err == nil {
		t.Fatal("AC-2: Done -> In Development passed")
	}
}

// TestProtectedSections is AC-1 and AC-2 of TASK-037.
func TestProtectedSections(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	for _, ref := range []string{"Document History", "13", "Open Questions", "14", "definition of done"} {
		if _, err := s.UpdateSection("TASK-001", ref, "- [x] Done."); err == nil || !strings.Contains(err.Error(), "changes only with task_") {
			t.Errorf("AC-1: UpdateSection(%s) = %v", ref, err)
		}
	}
	for _, ref := range []string{"13", "14"} {
		_, _, err := s.Create(New{Title: "Two " + ref, Priority: "P1", Milestone: "v0.1.0", Type: "feat", Sections: map[string]string{ref: "- [x] Done."}})
		if err == nil {
			t.Errorf("AC-2: Create with section %s passed", ref)
		}
	}
}

// TestSectionBodyHeadings is AC-1 to AC-3 of TASK-053.
func TestSectionBodyHeadings(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	before := read(t, s, "TASK-001")
	if _, err := s.UpdateSection("TASK-001", "12", "Revert.\n\n## 14. Definition of Done\n- [x] Done."); err == nil {
		t.Error("AC-1: UpdateSection with a section 14 heading passed")
	}
	if after := read(t, s, "TASK-001"); after != before {
		t.Error("AC-1: the refused UpdateSection changed the file")
	}
	if _, err := s.UpdateSection("TASK-001", "7", "### 7.1 Architecture\n\nFlow."); err != nil {
		t.Errorf("AC-2: UpdateSection(7) = %v", err)
	}
	if _, err := s.UpdateSection("TASK-001", "7.1", "### 6.1 Functional"); err == nil {
		t.Error("AC-3: UpdateSection(7.1) with a level 3 heading passed")
	}
}

// section14 returns the Definition of Done and the Document History of TASK-001.
func section14(s Store) (dod, history string) {
	dod, _ = s.Get("TASK-001", "14")
	history, _ = s.Get("TASK-001", "Document History")
	return dod, history
}

// TestDoDCheckUncheck is AC-3 and AC-7 of TASK-037.
func TestDoDCheckUncheck(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	must(t, s.CheckDoD("TASK-001", 1, "line one\nline two"))
	if dod, _ := section14(s); !strings.HasPrefix(dod, "- [x] All acceptance criteria pass, with evidence.\n  - Evidence: line one line two\n- [ ] ") {
		t.Fatalf("AC-7: section 14 =\n%s", dod)
	}
	if err := s.UncheckDoD("TASK-001", 1, ""); err == nil {
		t.Error("uncheck with no reason passed")
	}
	must(t, s.UncheckDoD("TASK-001", 1, "Wrong run."))
	dod, history := section14(s)
	if strings.Contains(dod, "Evidence") || !strings.HasPrefix(dod, "- [ ] All acceptance") || !strings.Contains(history, "Definition of Done item 1 unchecked: Wrong run.") {
		t.Fatalf("AC-3: section 14 =\n%s\nhistory =\n%s", dod, history)
	}
	if err := s.UncheckDoD("TASK-001", 1, "Again."); err == nil {
		t.Error("uncheck of an unchecked item passed")
	}
}

// TestDoDAddRemove is AC-4 and AC-5 of TASK-037.
func TestDoDAddRemove(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	must(t, s.CheckDoD("TASK-001", 5, "ok"))
	n, err := s.AddDoD("TASK-001", "Benchmark\npasses.")
	if dod, _ := section14(s); err != nil || n != 6 || !strings.HasSuffix(dod, "  - Evidence: ok\n- [ ] Benchmark passes.") {
		t.Fatalf("AC-4: AddDoD = %d, %v\n%s", n, err, dod)
	}
	must(t, s.RemoveDoD("TASK-001", 5, "Not needed."))
	dod, history := section14(s)
	if strings.Contains(dod, "ADRs are recorded") || strings.Contains(dod, "Evidence: ok") || !strings.Contains(history, "item 5 removed (ADRs are recorded for architectural decisions.): Not needed.") {
		t.Fatalf("AC-5: section 14 =\n%s\nhistory =\n%s", dod, history)
	}
}

// TestDoDDoneTask is AC-6 of TASK-037.
func TestDoDDoneTask(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	force(t, s, "TASK-001", "Done")
	_, addErr := s.AddDoD("TASK-001", "x")
	errs := []error{s.CheckDoD("TASK-001", 1, "x"), s.UncheckDoD("TASK-001", 1, "x"), s.RemoveDoD("TASK-001", 1, "x"), addErr}
	for i, err := range errs {
		if err == nil || !strings.Contains(err.Error(), "TASK-001 is Done") {
			t.Errorf("tool %d on a Done task = %v", i, err)
		}
	}
}
