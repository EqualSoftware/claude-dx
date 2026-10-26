package task

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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

// TestReadWaitsForWrite is AC-1 of TASK-062.
func TestReadWaitsForWrite(t *testing.T) {
	s := newStore(t)
	create(t, s, "A task")
	release := holdDocsLock(t, s, "exclusive")
	waitsFor(t, callAsync(t, s, "task_list", map[string]any{}), release)
}

// TestReadDoesNotWaitForRead is AC-2 of TASK-062.
func TestReadDoesNotWaitForRead(t *testing.T) {
	s := newStore(t)
	create(t, s, "A task")
	holdDocsLock(t, s, "shared")
	select {
	case <-callAsync(t, s, "task_get", map[string]any{"id": "TASK-001"}):
	case <-time.After(10 * time.Second):
		t.Fatal("task_get waited for a shared lock")
	}
}

// TestWriteWaitsForRead is AC-3 of TASK-062.
func TestWriteWaitsForRead(t *testing.T) {
	s := newStore(t)
	create(t, s, "A task")
	release := holdDocsLock(t, s, "shared")
	done := callAsync(t, s, "task_update_section", map[string]any{"id": "TASK-001", "section": "1", "body": "New summary."})
	waitsFor(t, done, release)
}
