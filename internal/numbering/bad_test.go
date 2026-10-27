package numbering

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBadProjectFileIsKept(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(a.Dir, "project.json"), []byte("{not json"), 0o644))
	must(createTask(docs)(4))
	n, err := a.Next("task", docs, taskPattern, createTask(docs))
	if err != nil || n != 5 {
		t.Fatalf("Next = %d, %v; want 5 from the highest file", n, err)
	}
	bad, err := os.ReadFile(filepath.Join(a.Dir, "project.json.bad"))
	if err != nil || string(bad) != "{not json" {
		t.Fatalf("project.json.bad = %q, %v", bad, err)
	}
}

func TestBadCountersKeepOtherKeys(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	if err := os.WriteFile(filepath.Join(a.Dir, "project.json"), []byte(`{"path":"/x","counters":[1]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil || n != 1 {
		t.Fatalf("Next = %d, %v", n, err)
	}
	data, _ := os.ReadFile(filepath.Join(a.Dir, "project.json"))
	if !strings.Contains(string(data), `"path":"/x"`) || !strings.Contains(string(data), `"counters":{"task":1}`) {
		t.Fatalf("project.json = %s", data)
	}
	if _, err := os.Stat(filepath.Join(a.Dir, "project.json.bad")); err == nil {
		t.Error("a file with good JSON was moved aside")
	}
}

func TestBadFileNamesAreUnique(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	for i, text := range []string{"{one", "{two"} {
		if err := os.WriteFile(filepath.Join(a.Dir, "project.json"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil {
			t.Fatal(i, err)
		}
	}
	one, _ := os.ReadFile(filepath.Join(a.Dir, "project.json.bad"))
	two, _ := os.ReadFile(filepath.Join(a.Dir, "project.json.bad.1"))
	if string(one) != "{one" || string(two) != "{two" {
		t.Fatalf("bad files = %q, %q", one, two)
	}
}
