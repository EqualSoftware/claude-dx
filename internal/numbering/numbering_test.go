package numbering

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
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
	n, err := Allocator{Dir: cache}.Next("task", docs, taskPattern, createTask(docs))
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
