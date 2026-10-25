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
		strings.Repeat("word ", 20): "word-word-word-word-word-word-word-word-word-word-word-word",
		strings.Repeat("x", 70):     strings.Repeat("x", 60),
	}
	for title, want := range tests {
		if got := Slug(title); got != want {
			t.Errorf("Slug(%q) = %q, want %q", title, got, want)
		}
	}
}
