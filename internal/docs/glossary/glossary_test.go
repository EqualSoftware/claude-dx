package glossary

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newStore(t *testing.T) Store {
	t.Helper()
	return Store{Path: filepath.Join(t.TempDir(), "docs", "GLOSSARY.md")}
}

// connect returns a client session to a server with the glossary tools of s, over memory.
func connect(t *testing.T, s Store) *mcp.ClientSession {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "dx", Version: "test"}, nil)
	Register(server, s)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func call(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return res.Content[0].(*mcp.TextContent).Text, res.IsError
}

func upsert(t *testing.T, s Store, term string, doNotUse ...string) {
	t.Helper()
	if _, err := s.Upsert(Row{Term: term, Definition: "About " + term + ".", Area: "process", DoNotUse: doNotUse, Source: "x"}); err != nil {
		t.Fatal(err)
	}
}

// expect calls the tool and stops the test if the result is not want, or if it is an error.
func expect(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, want string) {
	t.Helper()
	if text, isErr := call(t, session, name, args); isErr || text != want {
		t.Fatalf("%s(%v) = %q, error %v; want %q", name, args, text, isErr, want)
	}
}

func TestToolsFlow(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	expect(t, session, "glossary_lookup", map[string]any{"term": "Task"}, "No glossary. docs/GLOSSARY.md does not exist.")
	args := map[string]any{"term": "Task", "definition": "A spec file.", "area": "process", "do_not_use": []string{"ticket", "issue"}, "source": "`docs/tasks/`"}
	expect(t, session, "glossary_upsert", args, `Added "Task".`)
	// AC-1: a Do not use word gives the approved term.
	expect(t, session, "glossary_lookup", map[string]any{"term": "TICKET"}, `Use "Task", not "ticket".`)
	expect(t, session, "glossary_lookup", map[string]any{"term": "task"}, "| Task | A spec file. | process | ticket, issue | `docs/tasks/` |")
	args["definition"] = "A spec file for one unit of work."
	expect(t, session, "glossary_upsert", args, `Changed "Task".`)
	expect(t, session, "glossary_lookup", map[string]any{"term": "tas"}, "No match.")
	expect(t, session, "glossary_delete", map[string]any{"term": "task"}, `Deleted "task".`)
	data, _ := os.ReadFile(s.Path)
	if string(data) != templateText {
		t.Fatalf("file after delete =\n%s", data)
	}
}

func TestUpsertSortsAndEscapes(t *testing.T) {
	s := newStore(t)
	upsert(t, s, "dx server")
	upsert(t, s, "Cache")
	upsert(t, s, "ADR")
	if _, err := s.Upsert(Row{Term: "Pipe", Definition: "The | character.\nIt splits cells.", Area: "build", Source: "x"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(s.Path)
	want := templateText +
		"| ADR | About ADR. | process | — | x |\n" +
		"| Cache | About Cache. | process | — | x |\n" +
		"| dx server | About dx server. | process | — | x |\n" +
		"| Pipe | The \\| character. It splits cells. | build | — | x |\n"
	if string(data) != want {
		t.Fatalf("file =\n%s\nwant\n%s", data, want)
	}
	text, _, _ := s.Lookup("pipe")
	if text != "| Pipe | The \\| character. It splits cells. | build | — | x |" {
		t.Fatalf("Lookup(pipe) = %q", text)
	}
}

func TestUpsertConflict(t *testing.T) {
	s := newStore(t)
	upsert(t, s, "Task", "ticket")
	if _, err := s.Upsert(Row{Term: "Ticket", Definition: "D.", Area: "a", Source: "x"}); err == nil || !strings.Contains(err.Error(), `term "Task"`) {
		t.Fatalf("Upsert of a Do not use word = %v", err)
	}
	if _, err := s.Upsert(Row{Term: "Issue", Definition: "D.", Area: "a", DoNotUse: []string{"task"}, Source: "x"}); err == nil || !strings.Contains(err.Error(), `"Task" is a term`) {
		t.Fatalf("Upsert with a term in Do not use = %v", err)
	}
	upsert(t, s, "task", "ticket", "card") // the same term with other text is not a conflict
}

func TestParseError(t *testing.T) {
	s := newStore(t)
	upsert(t, s, "ADR")
	data, _ := os.ReadFile(s.Path)
	bad := strings.Replace(string(data), "| ADR | About ADR. |", "| ADR |", 1)
	if err := os.WriteFile(s.Path, []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := s.Lookup("ADR")
	if err == nil || !strings.Contains(err.Error(), "line 13: a row must have 5 cells, it has 4") {
		t.Fatalf("Lookup = %v", err)
	}
}

func TestDeleteMissing(t *testing.T) {
	s := newStore(t)
	if err := s.Delete("x"); !errors.Is(err, ErrNoFile) {
		t.Fatalf("Delete with no file = %v", err)
	}
	if _, err := os.Stat(filepath.Dir(s.Path)); err == nil {
		t.Fatal("Delete made docs/")
	}
	upsert(t, s, "ADR")
	if err := s.Delete("x"); err == nil {
		t.Fatal("Delete of a missing term did not fail")
	}
}

// TestRepositoryRoundTrip reads and writes the glossary of this repository. The file must not change.
func TestRepositoryRoundTrip(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "GLOSSARY.md"))
	if err != nil {
		t.Skip("no repository glossary")
	}
	s := newStore(t)
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.Path, want, 0o644); err != nil {
		t.Fatal(err)
	}
	upsert(t, s, "Zzz")
	must(t, s.Delete("Zzz"))
	got, _ := os.ReadFile(s.Path)
	if string(got) != strings.ReplaceAll(string(want), "\r\n", "\n") {
		t.Fatalf("GLOSSARY.md changed:\n%s", got)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// BenchmarkUpsert500 measures NFR-1: glossary operations on 500 rows take less than 20 ms.
func BenchmarkUpsert500(b *testing.B) {
	var sb strings.Builder
	sb.WriteString(templateText)
	for i := range 500 {
		fmt.Fprintf(&sb, "| Term %03d | A definition of one or two sentences for term %d. | process | word %d, other %d | `docs/x.md` |\n", i, i, i, i)
	}
	s := Store{Path: filepath.Join(b.TempDir(), "GLOSSARY.md")}
	if err := os.WriteFile(s.Path, []byte(sb.String()), 0o644); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, err := s.Upsert(Row{Term: "Term 250", Definition: "Changed.", Area: "process", Source: "x"}); err != nil {
			b.Fatal(err)
		}
	}
}
