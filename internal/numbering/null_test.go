package numbering

import (
	"os"
	"testing"
)

// A project.json with the JSON value null has no keys, and Next does not panic.
func TestNextWithNullFile(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	if err := os.WriteFile(a.path(), []byte("null"), 0o644); err != nil {
		t.Fatal(err)
	}
	docs := t.TempDir()
	if n, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil || n != 1 {
		t.Fatalf("Next = %d, %v", n, err)
	}
	if data, _ := os.ReadFile(a.path()); string(data) != `{"counters":{"task":1}}` {
		t.Fatalf("project.json = %s", data)
	}
}
