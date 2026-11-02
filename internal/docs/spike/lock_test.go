package spike

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

func waitsFor(t *testing.T, done <-chan time.Time, release func()) {
	t.Helper()
	time.Sleep(100 * time.Millisecond)
	released := time.Now()
	release()
	if (<-done).Before(released) {
		t.Fatal("the tool did not wait for the docs lock")
	}
}

// R-9: a read waits for a write, a write waits for a read, and reads do not wait for reads.
func TestToolsHoldDocsLock(t *testing.T) {
	s := newStore(t)
	must(t, s.Create("a", "A", "Q?"))
	release := holdDocsLock(t, s, "exclusive")
	waitsFor(t, callAsync(t, s, "spike_list", map[string]any{}), release)

	release = holdDocsLock(t, s, "shared")
	waitsFor(t, callAsync(t, s, "spike_note", map[string]any{"slug": "a", "text": "x"}), release)

	holdDocsLock(t, s, "shared")
	select {
	case <-callAsync(t, s, "spike_get", map[string]any{"slug": "a"}):
	case <-time.After(10 * time.Second):
		t.Fatal("spike_get waited for a shared lock")
	}
}
