package numbering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/docs/docfile"
)

var taskPattern = regexp.MustCompile(`^TASK-(\d{3})-`)

// createTask returns a create function that makes TASK-NNN-x.md in dir with exclusive create.
func createTask(dir string) func(int) error {
	return func(n int) error {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("TASK-%03d-x.md", n)), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		return f.Close()
	}
}

func TestNextUsesHighestFileAndCounter(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	if err := createTask(docs)(7); err != nil {
		t.Fatal(err)
	}
	n, err := a.Next("task", docs, taskPattern, createTask(docs))
	if err != nil || n != 8 {
		t.Fatalf("Next = %d, %v; want 8 from the highest file", n, err)
	}
	other := t.TempDir() // A worktree with fewer files.
	n, err = a.Next("task", other, taskPattern, createTask(other))
	if err != nil || n != 9 {
		t.Fatalf("Next = %d, %v; want 9 from the counter", n, err)
	}
}

// TestNextTakesNoLock checks project-data#R-1.
func TestNextTakesNoLock(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	if _, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(a.Dir, "lock")); err == nil {
		t.Fatal("Next made the numbering lock file")
	}
}

func TestNextKeepsCounterWhenCreateFails(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	docs := t.TempDir()
	_, err := a.Next("adr", docs, taskPattern, func(int) error { return fmt.Errorf("disk full") })
	if err == nil {
		t.Fatal("Next did not return the create error")
	}
	n, err := a.Next("adr", docs, taskPattern, createTask(docs))
	if err != nil || n != 1 {
		t.Fatalf("Next = %d, %v; want 1", n, err)
	}
}

func TestHighestMissingDir(t *testing.T) {
	n, err := Highest(filepath.Join(t.TempDir(), "none"), taskPattern)
	if err != nil || n != 0 {
		t.Fatalf("Highest = %d, %v; want 0, nil", n, err)
	}
}

// TestHelperAllocate runs in a child process (see TestNextParallelProcesses).
func TestHelperAllocate(t *testing.T) {
	cache, docs := os.Getenv("NUMBERING_CACHE"), os.Getenv("NUMBERING_DOCS")
	if cache == "" {
		t.Skip("helper process only")
	}
	// The docs tools call Next with the docs lock (project-data#R-1).
	var n int
	err := docfile.Locked(filepath.Join(cache, "docs.lock"), func() error {
		var err error
		n, err = Allocator{Dir: cache}.Next("task", docs, taskPattern, createTask(docs))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n)
}

func TestNextParallelProcesses(t *testing.T) {
	cache := t.TempDir()
	worktrees := []string{t.TempDir(), t.TempDir()}
	const count = 20
	cmds := make([]*exec.Cmd, count)
	outs := make([][]byte, count)
	errs := make(chan error, count)
	for i := range cmds {
		cmds[i] = exec.Command(os.Args[0], "-test.run=^TestHelperAllocate$", "-test.count=1")
		cmds[i].Env = append(os.Environ(), "NUMBERING_CACHE="+cache, "NUMBERING_DOCS="+worktrees[i%2])
	}
	for i, cmd := range cmds {
		go func() {
			out, err := cmd.Output()
			outs[i] = out
			errs <- err
		}()
	}
	for range cmds {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	var numbers []int
	for _, out := range outs {
		var n int
		if _, err := fmt.Sscan(string(out), &n); err != nil {
			t.Fatalf("bad helper output %q", out)
		}
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	for i, n := range numbers {
		if n != i+1 {
			t.Fatalf("numbers = %v, want 1 to %d with no duplicates", numbers, count)
		}
	}
}

// TestProjectFile checks project-data#R-2, R-3: the counter is in project.json, and an old counters.json is
// not read.
func TestProjectFile(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	if err := os.WriteFile(filepath.Join(a.Dir, "counters.json"), []byte(`{"task":50}`), 0o644); err != nil {
		t.Fatal(err)
	}
	docs := t.TempDir()
	if n, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil || n != 1 {
		t.Fatalf("Next = %d, %v; want 1", n, err)
	}
	data, err := os.ReadFile(filepath.Join(a.Dir, "project.json"))
	if err != nil || string(data) != `{"counters":{"task":1}}` {
		t.Fatalf("project.json = %q, %v", data, err)
	}
}

// A write changes only the key "counters" and keeps the other keys of project.json (ADR-007).
func TestNextKeepsOtherKeys(t *testing.T) {
	a := Allocator{Dir: t.TempDir()}
	if err := os.WriteFile(a.path(), []byte(`{"path":"/p","last_used":"x","extra":{"a":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	docs := t.TempDir()
	if _, err := a.Next("task", docs, taskPattern, createTask(docs)); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.path())
	want := `{"counters":{"task":1},"extra":{"a":1},"last_used":"x","path":"/p"}`
	if string(data) != want {
		t.Fatalf("project.json = %s, want %s", data, want)
	}
}
