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
	if _, _, err := s.Create(New{Title: title, Priority: "P1", Milestone: "v0.1.0", Type: "feat", Area: []string{"cli"}}); err != nil {
		t.Fatal(err)
	}
}

func update(t *testing.T, s Store, id string, u Update) {
	t.Helper()
	if _, err := s.UpdateFields(id, u); err != nil {
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
		status, note, want string
	}{
		{"In Review", "", "section 1 has the template text"},
		{"In Review", "", "give a size"},
		{"In Review", "", "give the platforms"},
		{"Blocked", "", "give a note"},
		{"Done", "", "Definition of Done items 1, 2, 3, 4, 5, 6 are not checked"},
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
	if !strings.Contains(doc, "status: In Development\n") || !strings.Contains(doc, "| 2026-10-25 | Tester | Status: In Development. |") {
		t.Fatalf("status change not written:\n%s", doc)
	}
}

func TestInReviewWhenComplete(t *testing.T) {
	s := newStore(t)
	sections := map[string]string{}
	for _, ref := range leafSections {
		sections[ref] = "Filled " + ref + "."
	}
	if _, _, err := s.Create(New{Title: "Full", Priority: "P1", Milestone: "v0.1.0", Type: "feat", Area: []string{"cli"}, Size: "M", Platforms: []string{"Linux"}, Sections: sections}); err != nil {
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

// force sets the status of a Draft task with no status rules.
func force(t *testing.T, s Store, id, status string) {
	t.Helper()
	must(t, s.edit(id, func(d *doc) error {
		d.front = strings.Replace(d.front, "status: Draft", "status: "+status, 1)
		return nil
	}))
}

func TestUpdateFields(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	u := Update{Title: "New title", Priority: "P0", Milestone: "v0.2.0", Type: "fix", Size: "M", Owner: "Equal", DependsOn: []string{}, Related: []string{"ADR-001"}}
	if _, err := s.UpdateFields("TASK-001", u); err != nil {
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
		if _, err := s.UpdateFields("TASK-001", bad); err == nil {
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
	must(t, s.SetStatus("TASK-003", "Cancelled", "Not necessary."))
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
	_, err = s.UpdateFields("TASK-002", Update{DependsOn: []string{"TASK-001"}})
	refused(err, "TASK-002 -> TASK-001 -> TASK-002")
	if read(t, s, "TASK-002") != before {
		t.Fatal("AC-2: a refused change changed the file")
	}
	_, err = s.UpdateFields("TASK-002", Update{DependsOn: []string{"TASK-002"}})
	refused(err, "itself")
	force(t, s, "TASK-003", "Done")
	force(t, s, "TASK-004", "Cancelled")
	for _, id := range []string{"TASK-003", "TASK-004"} {
		_, _, err = s.Create(New{Title: "X", Priority: "P1", Milestone: "v0.1.0", Type: "feat", DependsOn: []string{id}})
		refused(err, "Put it in related")
	}
	update(t, s, "TASK-002", Update{Milestone: "v0.2.0"})
	_, err = s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002"}})
	refused(err, "later milestone v0.2.0")
	update(t, s, "TASK-002", Update{Milestone: "v0.1.0", Priority: "P3"})
	warnings, err := s.UpdateFields("TASK-001", Update{DependsOn: []string{"TASK-002"}})
	if err != nil || !strings.Contains(warnings, "TASK-002 has the lower priority P3") {
		t.Fatalf("AC-6: warnings = %q, %v", warnings, err)
	}
}

// TestUncheckedDoDMessage is AC-7 of TASK-023.
func TestUncheckedDoDMessage(t *testing.T) {
	s := newStore(t)
	create(t, s, "One")
	must(t, s.CheckDoD("TASK-001", 1, "Evidence."))
	err := s.SetStatus("TASK-001", "Done", "")
	if err == nil || !strings.HasSuffix(err.Error(), ": Definition of Done items 2, 3, 4, 5, 6 are not checked") {
		t.Fatalf("SetStatus(Done) = %v", err)
	}
}
