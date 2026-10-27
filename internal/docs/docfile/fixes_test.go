package docfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateAtomicAndExclusive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.md")
	must(t, Create(path, []byte("one")))
	if err := Create(path, []byte("two")); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second Create = %v, want exists", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "one" {
		t.Fatalf("file = %q, want one", data)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("directory has %d entries, want only a.md", len(entries))
	}
}

func TestWriteRefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real.md")
	must(t, os.WriteFile(target, []byte("real"), 0o644))
	link := filepath.Join(dir, "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symbolic links are not available:", err)
	}
	if err := Write(link, []byte("new")); err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("Write over a link = %v, want a symbolic link error", err)
	}
	if data, _ := os.ReadFile(target); string(data) != "real" {
		t.Fatalf("target = %q, want real", data)
	}
}

func TestWriteRemovesStaleTemp(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, ".a.md.tmp-123")
	fresh := filepath.Join(dir, ".a.md.tmp-456")
	must(t, os.WriteFile(stale, nil, 0o644))
	must(t, os.WriteFile(fresh, nil, 0o644))
	old := time.Now().Add(-time.Hour)
	must(t, os.Chtimes(stale, old, old))
	must(t, Write(filepath.Join(dir, "a.md"), []byte("x")))
	if _, err := os.Stat(stale); err == nil {
		t.Error("the stale temporary file stays")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("the new temporary file is removed")
	}
}

func TestCheckTitleLineBreak(t *testing.T) {
	for _, title := range []string{"A\nB", "A\rB", "A\r\nB"} {
		if err := CheckTitle(title); err == nil || !strings.Contains(err.Error(), "line break") {
			t.Errorf("CheckTitle(%q) = %v, want a line break error", title, err)
		}
	}
	if got := Cell("a\r\nb|c"); got != `a b\|c` {
		t.Errorf("Cell = %q", got)
	}
}

func TestCreateWithoutHardLinks(t *testing.T) {
	old := link
	link = func(string, string) error { return errors.New("links are not supported") }
	t.Cleanup(func() { link = old })
	dir := t.TempDir()
	path := filepath.Join(dir, "a.md")
	must(t, Create(path, []byte("one")))
	if err := Create(path, []byte("two")); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second Create = %v, want exists", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "one" {
		t.Fatalf("file = %q, want one", data)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("directory has %d entries, want only a.md", len(entries))
	}
}
