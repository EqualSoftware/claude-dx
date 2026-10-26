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
	for _, title := range []string{"Old", "New"} {
		if _, err := s.Create(sample(title)); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 { // The second call must not add the line again.
		if err := s.Supersede("ADR-001", "ADR-002"); err != nil {
			t.Fatal(err)
		}
	}
	old, _ := s.Get("ADR-001", "")
	if !strings.Contains(old, "- Status: Superseded by ADR-002") {
		t.Fatalf("old =\n%s", old)
	}
	context, _ := s.Get("ADR-002", "Context")
	if context != "Replaces ADR-001.\n\nThe context." {
		t.Fatalf("new Context = %q", context)
	}
	if err := s.Supersede("ADR-001", "ADR-001"); err == nil {
		t.Error("an ADR replaced itself")
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
