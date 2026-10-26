package docfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	k := NewKind("TASK")
	for _, id := range []string{"TASK-007", "task-7", "7", "007"} {
		if n, err := k.Parse(id); err != nil || n != 7 {
			t.Errorf("Parse(%q) = %d, %v", id, n, err)
		}
	}
	for _, id := range []string{"ADR-007", "TASK-1234", "x"} {
		if _, err := k.Parse(id); err == nil {
			t.Errorf("Parse(%q) did not fail", id)
		}
	}
}

func TestFindAndDuplicates(t *testing.T) {
	k := NewKind("ADR")
	dir := t.TempDir()
	for _, name := range []string{"ADR-001-a.md", "ADR-002-b.md", "ADR-002-c.md", "notes.md"} {
		if err := Create(filepath.Join(dir, name), nil); err != nil {
			t.Fatal(err)
		}
	}
	if path, err := k.Find(dir, "1"); err != nil || filepath.Base(path) != "ADR-001-a.md" {
		t.Fatalf("Find(1) = %q, %v", path, err)
	}
	if _, err := k.Find(dir, "2"); err == nil || !strings.Contains(err.Error(), "has 2 files") {
		t.Fatalf("Find(2) = %v, want duplicate error", err)
	}
	files, _ := k.Files(dir)
	if got := Duplicates(k, files); got != "Warning: ADR-002 has 2 files. Give each file a different number.\n" {
		t.Fatalf("Duplicates = %q", got)
	}
}

func TestCreateIsExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new", "x.md")
	if err := Create(path, []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := Create(path, []byte("b")); err == nil {
		t.Fatal("second Create did not fail")
	}
	if data, _ := os.ReadFile(path); string(data) != "a" {
		t.Fatalf("file = %q", data)
	}
}

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Use stdio":                 "use-stdio",
		"  Load dx: with flags!  ":  "load-dx-with-flags",
		"Ação em português":         "a-o-em-portugu-s",
		strings.Repeat("word ", 20): "word-word-word-word-word-word-word-word",
		strings.Repeat("x", 70):     strings.Repeat("x", 40),
	}
	for title, want := range tests {
		if got := Slug(title); got != want {
			t.Errorf("Slug(%q) = %q, want %q", title, got, want)
		}
	}
}

// TestCheckTitleSlug is AC-1 and AC-2 of TASK-044. "ção" has the ASCII letter "o", so AC-1 uses a title
// with no ASCII letter or digit.
func TestCheckTitleSlug(t *testing.T) {
	for _, title := range []string{"日本語", "ç ã é", "—"} {
		if err := CheckTitle(title); err == nil || !strings.Contains(err.Error(), "ASCII letters or digits") {
			t.Errorf("CheckTitle(%q) = %v", title, err)
		}
	}
	if err := CheckTitle("Ação 2"); err != nil {
		t.Errorf("CheckTitle(Ação 2) = %v", err)
	}
}

func TestCheckTitle(t *testing.T) {
	if err := CheckTitle("x" + strings.Repeat("é", 39)); err != nil {
		t.Errorf("40 characters: %v", err)
	}
	if err := CheckTitle(strings.Repeat("x", 41)); err == nil || !strings.Contains(err.Error(), "41 characters") {
		t.Errorf("41 characters: %v", err)
	}
	if err := CheckTitle("  "); err == nil {
		t.Error("an empty title passed")
	}
	titles := map[string]string{"ADR-001": "Use SQLite", "ADR-002": "Use stdio"}
	if err := CheckNewTitle("use sqlite", titles); err == nil || !strings.Contains(err.Error(), "ADR-001") {
		t.Errorf("duplicate title: %v", err)
	}
	if err := CheckNewTitle("Use HTTP", titles); err != nil {
		t.Errorf("new title: %v", err)
	}
}
