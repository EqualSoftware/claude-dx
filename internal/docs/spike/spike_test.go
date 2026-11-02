package spike

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newStore(t *testing.T) Store {
	t.Helper()
	data := t.TempDir()
	return Store{Dir: filepath.Join(data, "spikes"), Lock: filepath.Join(data, "docs.lock")}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func fixDate(t *testing.T) {
	t.Helper()
	old := now
	now = func() time.Time { return time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = old })
}

func refuses(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want text %q", err, want)
	}
}

func read(t *testing.T, s Store, slug string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(s.Dir, slug+".md"))
	must(t, err)
	return string(data)
}

// R-2, R-3.
func TestCreate(t *testing.T) {
	fixDate(t)
	s := newStore(t)
	must(t, s.Create("lock-on-nfs", "File locks on NFS", "Do the lock calls work on NFS?"))
	want := "---\nslug: lock-on-nfs\ntitle: File locks on NFS\nstatus: Open\ncreated: 2026-11-02\n---\n\n" +
		"## Question\n\nDo the lock calls work on NFS?\n\n## Notes\n\n## Result\n"
	if got := read(t, s, "lock-on-nfs"); got != want {
		t.Fatalf("file =\n%q\nwant\n%q", got, want)
	}
	refuses(t, s.Create("lock-on-nfs", "Other", "Q?"), "exists")
	refuses(t, s.Create("a-b", "T", "  "), "give the question")
	refuses(t, s.Create("Bad_Slug", "T", "Q?"), "kebab-case")
	refuses(t, s.Create("a--b", "T", "Q?"), "kebab-case")
	refuses(t, s.Create(strings.Repeat("a", 41), "T", "Q?"), "limit is 40")
	must(t, s.Create(strings.Repeat("a", 40), "T", "Q?"))
	refuses(t, s.Create("no-title", "", "Q?"), "title")
	refuses(t, s.Create("heading", "T", "# Head\n"), "heading")
}

// R-4.
func TestNote(t *testing.T) {
	fixDate(t)
	s := newStore(t)
	must(t, s.Create("a", "A", "Q?"))
	must(t, s.Note("a", "First."))
	must(t, s.Note("a", " Second. "))
	got, err := s.Get("a", "Notes")
	must(t, err)
	if want := "- 2026-11-02: First.\n- 2026-11-02: Second."; got != want {
		t.Fatalf("Notes = %q, want %q", got, want)
	}
	refuses(t, s.Note("a", "  "), "give the text")
	refuses(t, s.Note("none", "x"), "does not exist")
	if q, _ := s.Get("a", "Question"); q != "Q?" {
		t.Fatalf("Question = %q", q)
	}
}

// R-5, R-6.
func TestClose(t *testing.T) {
	fixDate(t)
	s := newStore(t)
	must(t, s.Create("a", "A", "Q?"))
	must(t, s.Note("a", "Found."))
	refuses(t, s.Close("a", " "), "give the result")
	must(t, s.Close("a", "It works."))
	if got, _ := s.Get("a", "Result"); got != "It works." {
		t.Fatalf("Result = %q", got)
	}
	list, _, err := s.List(Closed)
	must(t, err)
	if len(list) != 1 || list[0].Slug != "a" {
		t.Fatalf("Closed list = %+v", list)
	}
	before := read(t, s, "a")
	refuses(t, s.Note("a", "More."), "Closed")
	refuses(t, s.Close("a", "Again."), "Closed")
	if read(t, s, "a") != before {
		t.Fatal("a Closed spike changed")
	}
	must(t, s.Delete([]string{"a"}))
}

// R-7.
func TestListAndGet(t *testing.T) {
	fixDate(t)
	s := newStore(t)
	if _, _, err := s.List(""); !errors.Is(err, ErrNoDir) {
		t.Fatalf("List without folder = %v", err)
	}
	must(t, s.Create("b", "B title", "QB?"))
	must(t, s.Create("a", "A title", "QA?"))
	must(t, s.Create("c", "C title", "QC?"))
	must(t, s.Close("a", "Done."))
	text, err := s.listText("")
	must(t, err)
	want := "b  Open    2026-11-02  B title\nc  Open    2026-11-02  C title\na  Closed  2026-11-02  A title\n"
	if text != want {
		t.Fatalf("list =\n%q\nwant\n%q", text, want)
	}
	if text, _ = s.listText(Open); strings.Contains(text, "Closed") {
		t.Fatalf("Open list = %q", text)
	}
	if _, err := s.listText("Bad"); err == nil {
		t.Fatal("bad status accepted")
	}
	full, err := s.Get("b", "")
	must(t, err)
	if !strings.HasPrefix(full, "---\nslug: b\n") || !strings.Contains(full, "## Question") {
		t.Fatalf("Get = %q", full)
	}
	if _, err := s.Get("b", "Nope"); err == nil {
		t.Fatal("bad section accepted")
	}
	_, err = s.Get("zz", "")
	refuses(t, err, "does not exist")
	must(t, os.WriteFile(filepath.Join(s.Dir, "bad.md"), []byte("---\nslug: [\n---\n"), 0o644))
	if text, _ = s.listText(""); !strings.Contains(text, "Warning: bad.md") {
		t.Fatalf("no warning: %q", text)
	}
}

// R-8.
func TestDelete(t *testing.T) {
	fixDate(t)
	s := newStore(t)
	must(t, s.Create("a", "A", "Q?"))
	must(t, s.Create("b", "B", "Q?"))
	must(t, s.Close("b", "R."))
	refuses(t, s.Delete([]string{"a", "none"}), "does not exist")
	if _, err := os.Stat(filepath.Join(s.Dir, "a.md")); err != nil {
		t.Fatal("a refused call changed a spike")
	}
	refuses(t, s.Delete(nil), "at least one")
	must(t, s.Delete([]string{"a", "b"}))
	if list, _, _ := s.List(""); len(list) != 0 {
		t.Fatalf("spikes left: %+v", list)
	}
}
