package adr

import (
	"errors"
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
		Dir:   filepath.Join(t.TempDir(), "docs", "decisions"),
		Alloc: numbering.Allocator{Dir: t.TempDir()},
		Now:   func() time.Time { return time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC) },
		Lock:  filepath.Join(t.TempDir(), "docs.lock"),
	}
}

func sample(title string) New {
	return New{
		Title:        title,
		Context:      "The context.",
		Decision:     "The decision.",
		Alternatives: []Alternative{{Option: "Other", Reason: "rejected, too slow"}},
		Consequences: Consequences{Positive: "Fast.", Negative: "None.", FollowUp: "None."},
		Evidence:     "`main.go:1`",
	}
}

const want001 = `# ADR-001: Use stdio

- Status: Proposed
- Date: 2026-10-25
- Owners: Maintainer

## Context
The context.

## Decision
The decision.

## Alternatives
- Other: rejected, too slow

## Consequences
- Positive: Fast.
- Negative: None.
- Follow-up: None.

## Evidence
` + "`main.go:1`" + `
`

func TestCreate(t *testing.T) {
	s := newStore(t)
	path, err := s.Create(sample("Use stdio"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "ADR-001-use-stdio.md" {
		t.Fatalf("path = %q", path)
	}
	data, _ := os.ReadFile(path)
	if string(data) != want001 {
		t.Fatalf("file =\n%s\nwant\n%s", data, want001)
	}
}

func TestCreateNextNumber(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"One", "Two", "Three"} {
		if _, err := s.Create(sample(title)); err != nil {
			t.Fatal(err)
		}
	}
	path, err := s.Create(sample("Four"))
	if err != nil || filepath.Base(path) != "ADR-004-four.md" {
		t.Fatalf("Create = %q, %v; want ADR-004-four.md", path, err)
	}
}

func TestCreateRefusesStatus(t *testing.T) {
	adr := sample("X")
	adr.Status = "Superseded"
	if _, err := newStore(t).Create(adr); err == nil {
		t.Fatal("Create accepted status Superseded")
	}
}

func TestMissingDir(t *testing.T) {
	s := newStore(t)
	if _, _, err := s.List(); !errors.Is(err, ErrNoDir) {
		t.Fatalf("List = %v, want ErrNoDir", err)
	}
	if err := s.Update("ADR-001", "Accepted", "", ""); !errors.Is(err, ErrNoDir) {
		t.Fatalf("Update = %v, want ErrNoDir", err)
	}
	if _, err := os.Stat(s.Dir); err == nil {
		t.Fatal("a read or update made docs/decisions")
	}
}

func TestList(t *testing.T) {
	s := newStore(t)
	if _, err := s.Create(sample("Use stdio")); err != nil {
		t.Fatal(err)
	}
	list, _, err := s.List()
	if err != nil || len(list) != 1 || list[0].Status != "Proposed" || list[0].Title != "Use stdio" {
		t.Fatalf("List = %+v, %v", list, err)
	}
}

// TestListBadFile checks docs-adr#R-9. A directory with an ADR file name cannot be read on all platforms.
func TestListBadFile(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"One", "Two"} {
		if _, err := s.Create(sample(title)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(s.Dir, "ADR-003-bad.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	text, err := s.listText()
	if err != nil || !strings.Contains(text, "ADR-001") || !strings.Contains(text, "ADR-002") ||
		strings.Count(text, "Warning: ADR-003-bad.md:") != 1 {
		t.Fatalf("adr_list = %q, %v", text, err)
	}
	if err := s.WriteIndex(); err != nil {
		t.Fatalf("WriteIndex = %v", err)
	}
}

func TestGet(t *testing.T) {
	s := newStore(t)
	if _, err := s.Create(sample("Use stdio")); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"ADR-001", "adr-1", "1", "001"} {
		if got, err := s.Get(id, "Decision"); err != nil || got != "The decision." {
			t.Errorf("Get(%q, Decision) = %q, %v", id, got, err)
		}
	}
	if _, err := s.Get("ADR-002", ""); err == nil {
		t.Error("Get of a missing ADR did not fail")
	}
	if _, err := s.Get("ADR-001", "Nothing"); err == nil {
		t.Error("Get of a missing section did not fail")
	}
}

func TestUpdate(t *testing.T) {
	s := newStore(t)
	if _, err := s.Create(sample("Use stdio")); err != nil {
		t.Fatal(err)
	}
	if err := s.Update("ADR-001", "Accepted", "Evidence", "CI run 1."); err != nil {
		t.Fatal(err)
	}
	doc, _ := s.Get("ADR-001", "")
	if !strings.Contains(doc, "- Status: Accepted") || !strings.Contains(doc, "## Evidence\n\nCI run 1.") {
		t.Fatalf("doc =\n%s", doc)
	}
	if err := s.Update("ADR-001", "", "", ""); err == nil {
		t.Error("Update with no change did not fail")
	}
}

func TestSupersede(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Old", "New", "Draft"} {
		if _, err := s.Create(sample(title)); err != nil {
			t.Fatal(err)
		}
	}
	// docs-adr#R-6 to R-8: a Proposed ADR cannot replace an ADR.
	if err := s.Supersede("ADR-001", "ADR-003"); err == nil || !strings.Contains(err.Error(), "must be Accepted") {
		t.Fatalf("Supersede by a Proposed ADR = %v", err)
	}
	if err := s.Update("ADR-002", "Accepted", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Supersede("ADR-001", "ADR-002"); err != nil {
		t.Fatal(err)
	}
	old, _ := s.Get("ADR-001", "")
	if !strings.Contains(old, "- Status: Superseded by ADR-002") {
		t.Fatalf("old =\n%s", old)
	}
	context, _ := s.Get("ADR-002", "Context")
	if context != "The context." {
		t.Fatalf("new Context = %q", context)
	}
	checkSupersedesLine(t, s)
	if err := s.Supersede("ADR-001", "ADR-001"); err == nil {
		t.Error("an ADR replaced itself")
	}
}

func checkSupersedesLine(t *testing.T, s Store) {
	t.Helper()
	doc, _ := s.Get("ADR-002", "")
	if !strings.Contains(doc, "- Owners: Maintainer\n- Supersedes: ADR-001\n\n## Context") {
		t.Fatalf("new =\n%s", doc)
	}
}

// TestSupersedeNoOwners: an error on the new ADR writes nothing.
func TestSupersedeNoOwners(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Old", "New"} {
		_, err := s.Create(sample(title))
		must(t, err)
	}
	must(t, s.Update("ADR-002", "Accepted", "", ""))
	path, err := s.find("ADR-002")
	must(t, err)
	data, _ := os.ReadFile(path)
	broken := strings.Replace(string(data), "- Owners: Maintainer\n", "", 1)
	must(t, os.WriteFile(path, []byte(broken), 0o644))
	oldBefore, _ := s.Get("ADR-001", "")
	refused(t, s.Supersede("ADR-001", "ADR-002"), "no Owners line")
	if oldAfter, _ := s.Get("ADR-001", ""); oldAfter != oldBefore {
		t.Fatal("a refused Supersede changed ADR-001")
	}
	if newAfter, _ := s.Get("ADR-002", ""); newAfter != broken {
		t.Fatal("a refused Supersede changed ADR-002")
	}
}

// TestSupersedeNoDuplicate: a second call adds no second Supersedes line, also with CRLF line ends.
func TestSupersedeNoDuplicate(t *testing.T) {
	doc := "# ADR-002: New\r\n\r\n- Status: Accepted\r\n- Owners:Maintainer\r\n\r\n## Context\r\nx\r\n"
	once, err := addSupersedes(doc, 1)
	must(t, err)
	if !strings.Contains(once, "- Owners:Maintainer\r\n- Supersedes: ADR-001\r\n\r\n") {
		t.Fatalf("once = %q", once)
	}
	twice, err := addSupersedes(once, 1)
	if err != nil || twice != once {
		t.Fatalf("twice = %q, %v", twice, err)
	}
	s := newStore(t)
	for _, title := range []string{"Old", "New"} {
		_, err := s.Create(sample(title))
		must(t, err)
	}
	must(t, s.Update("ADR-002", "Accepted", "", ""))
	must(t, s.Supersede("ADR-001", "ADR-002"))
	_ = s.Supersede("ADR-001", "ADR-002") // refused: ADR-001 is superseded
	if got, _ := s.Get("ADR-002", ""); strings.Count(got, "- Supersedes: ADR-001") != 1 {
		t.Fatalf("new =\n%s", got)
	}
}

// TestSupersedeRefusesBadChain checks docs-adr#R-6 to R-8.
func TestSupersedeRefusesBadChain(t *testing.T) {
	s := newStore(t)
	for i, title := range []string{"Old", "New", "Other"} {
		_, err := s.Create(sample(title))
		must(t, err)
		must(t, s.Update(kind.ID(i+1), "Accepted", "", ""))
	}
	must(t, s.Supersede("ADR-001", "ADR-003"))
	before, _ := s.Get("ADR-003", "")
	refused(t, s.Supersede("ADR-003", "ADR-001"), "must be Accepted")
	if after, _ := s.Get("ADR-003", ""); after != before {
		t.Fatal("AC-1: a refused Supersede changed ADR-003")
	}
	refused(t, s.Supersede("ADR-001", "ADR-002"), "one replacement")
	refused(t, s.Update("ADR-001", "Accepted", "", ""), "cannot change")
	must(t, s.Update("ADR-001", "", "Evidence", "More evidence."))
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// refused fails the test if err is nil or does not contain want.
func refused(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestIndex(t *testing.T) {
	s := newStore(t)
	for _, title := range []string{"Old", "New"} {
		if _, err := s.Create(sample(title)); err != nil {
			t.Fatal(err)
		}
	}
	readme := filepath.Join(s.Dir, "README.md")
	data, _ := os.ReadFile(readme)
	if !strings.HasSuffix(string(data), "| [ADR-002](ADR-002-new.md) | New | Proposed | 2026-10-25 |\n") {
		t.Fatalf("README.md after create =\n%s", data)
	}
	if err := s.Update("ADR-002", "Accepted", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Supersede("ADR-001", "ADR-002"); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(readme)
	want := indexHeader +
		"| [ADR-001](ADR-001-old.md) | Old | Superseded by ADR-002 | 2026-10-25 |\n" +
		"| [ADR-002](ADR-002-new.md) | New | Accepted | 2026-10-25 |\n"
	if string(data) != want {
		t.Fatalf("README.md =\n%s\nwant\n%s", data, want)
	}
	list, _, _ := s.List()
	if len(list) != 2 {
		t.Fatalf("List has README.md: %v", list)
	}
}

// TestTitleRules checks docs-common#R-1, R-4.
// TestIndexEscapesPipes checks docs-common#R-6 for ADRs.
func TestIndexEscapesPipes(t *testing.T) {
	s := newStore(t)
	if _, err := s.Create(sample("Fix a|b parse")); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if !strings.Contains(string(data), "| Fix a\\|b parse | Proposed |") {
		t.Fatalf("README.md =\n%s", data)
	}
}

func TestTitleRules(t *testing.T) {
	s := newStore(t)
	_, err := s.Create(sample(strings.Repeat("x", 41)))
	refused(t, err, "41 characters")
	_, err = s.Create(sample("Use SQLite"))
	must(t, err)
	_, err = s.Create(sample("use sqlite"))
	refused(t, err, "ADR-001")
}
