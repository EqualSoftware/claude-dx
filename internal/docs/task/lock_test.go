package task

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
)

// TestHoldDocsLock is a helper process: it holds the docs lock in DX_DOCS_LOCK until stdin closes.
// The lock is exclusive if DX_LOCK_MODE is "exclusive", else shared.
func TestHoldDocsLock(t *testing.T) {
	path := os.Getenv("DX_DOCS_LOCK")
	if path == "" {
		t.Skip("helper process")
	}
	try := filelock.TryRLock
	if os.Getenv("DX_LOCK_MODE") == "exclusive" {
		try = filelock.TryLock
	}
	f, err := try(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	fmt.Println("ready")
	_, _ = io.ReadAll(os.Stdin)
}

// holdDocsLock holds the docs lock of s in a helper process with mode. release ends the helper
// process, and with it the lock.
func holdDocsLock(t *testing.T, s Store, mode string) (release func()) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHoldDocsLock$")
	cmd.Env = append(os.Environ(), "DX_DOCS_LOCK="+s.Lock, "DX_LOCK_MODE="+mode)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	release = func() { _ = stdin.Close(); _ = cmd.Wait() }
	t.Cleanup(release)
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper output = %q, %v", line, err)
	}
	return release
}

// callAsync calls the tool name of s in a goroutine and returns a channel that gets the end time.
func callAsync(t *testing.T, s Store, name string, args map[string]any) <-chan time.Time {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: "test"}, nil)
	Register(server, s)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	done := make(chan time.Time, 1)
	go func() {
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil || res.IsError {
			t.Errorf("%s: %v, %+v", name, err, res)
		}
		done <- time.Now()
	}()
	return done
}

// waitsFor checks that the call ends only after release.
func waitsFor(t *testing.T, done <-chan time.Time, release func()) {
	t.Helper()
	time.Sleep(100 * time.Millisecond)
	released := time.Now()
	release()
	if (<-done).Before(released) {
		t.Fatal("the tool did not wait for the docs lock")
	}
}

// create makes the task general/a.
func create(t *testing.T, s Store) {
	t.Helper()
	if _, err := s.Create(New{Spec: "general", Slug: "a", Title: "A task", Priority: "P1", Brief: "Do it.", Plan: []string{"Do it."}}); err != nil {
		t.Fatal(err)
	}
}

// TestReadWaitsForWrite checks docs-common#R-10.
func TestReadWaitsForWrite(t *testing.T) {
	s := newStore(t)
	create(t, s)
	release := holdDocsLock(t, s, "exclusive")
	waitsFor(t, callAsync(t, s, "task_list", map[string]any{}), release)
}

// TestReadDoesNotWaitForRead checks docs-common#R-10.
func TestReadDoesNotWaitForRead(t *testing.T) {
	s := newStore(t)
	create(t, s)
	holdDocsLock(t, s, "shared")
	select {
	case <-callAsync(t, s, "task_get", map[string]any{"id": "general/a"}):
	case <-time.After(10 * time.Second):
		t.Fatal("task_get waited for a shared lock")
	}
}

// TestWriteWaitsForRead checks docs-common#R-10.
func TestWriteWaitsForRead(t *testing.T) {
	s := newStore(t)
	create(t, s)
	release := holdDocsLock(t, s, "shared")
	done := callAsync(t, s, "task_update_section", map[string]any{"id": "general/a", "section": "Notes", "body": "A note."})
	waitsFor(t, done, release)
}

// TestTaskWriter is a helper process: it makes the tasks a/<DX_PREFIX>-1 to -10 in the stores of
// DX_ROOT and DX_DATA, each one covers one requirement, and sets each one to Done.
func TestTaskWriter(t *testing.T) {
	root, prefix := os.Getenv("DX_ROOT"), os.Getenv("DX_PREFIX")
	if root == "" {
		t.Skip("helper process")
	}
	s := storeAt(root, os.Getenv("DX_DATA"))
	first, _ := strconv.Atoi(os.Getenv("DX_FIRST"))
	for i := range 10 {
		id := fmt.Sprintf("a/%s-%d", prefix, i)
		_, err := s.Create(New{Spec: "a", Slug: id[2:], Title: "Task " + id[2:], Priority: "P1", Brief: "Do it.", Plan: []string{"Do it."},
			Covers: []string{fmt.Sprintf("R-%d", first+i)}})
		must(t, err)
		must(t, s.CheckStep(id, 1))
		must(t, s.CheckDoD(id, 1, "Test."))
		must(t, s.CheckDoD(id, 2, "Test."))
		_, err = s.SetStatus(id, "In Progress", "")
		must(t, err)
		_, err = s.SetStatus(id, "Done", "")
		must(t, err)
	}
}

// TestTwoProcesses is R-11: two processes change tasks and their spec at the same time. No write is lost
// and each file stays valid.
func TestTwoProcesses(t *testing.T) {
	s := newStore(t)
	addSpec(t, s, "a", "Approved", 20)
	var cmds []*exec.Cmd
	for i, prefix := range []string{"x", "y"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestTaskWriter$")
		cmd.Env = append(os.Environ(), "DX_ROOT="+s.Specs.Root, "DX_DATA="+filepath.Dir(s.Lock), "DX_PREFIX="+prefix,
			fmt.Sprintf("DX_FIRST=%d", 1+10*i))
		must(t, cmd.Start())
		cmds = append(cmds, cmd)
	}
	for _, cmd := range cmds {
		must(t, cmd.Wait())
	}
	tasks, _, err := s.List("a", "Done")
	must(t, err)
	if len(tasks) != 20 {
		t.Errorf("%d Done tasks, want 20", len(tasks))
	}
	_, rids, err := s.Specs.RequirementsLocked("a")
	must(t, err)
	sp, err := s.Specs.Get("a", "")
	must(t, err)
	if n := strings.Count(sp, "| Done |"); n != 20 || len(rids) != 20 {
		t.Errorf("%d Done requirements of %d, want 20:\n%s", n, len(rids), sp)
	}
}
