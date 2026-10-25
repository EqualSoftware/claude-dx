package adr

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect returns a client session to a server with the ADR tools of s, over memory.
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

func TestToolsFlow(t *testing.T) {
	session := connect(t, newStore(t))

	text, isErr := call(t, session, "adr_list", map[string]any{})
	if isErr || text != "No ADRs. docs/decisions does not exist." {
		t.Fatalf("adr_list = %q, error %v", text, isErr)
	}
	create := map[string]any{
		"title": "Use stdio", "context": "C.", "decision": "D.", "evidence": "E.",
		"alternatives": []map[string]any{{"option": "HTTP", "reason": "more work"}},
		"consequences": map[string]any{"positive": "P.", "negative": "N.", "follow_up": "None."},
	}
	text, isErr = call(t, session, "adr_create", create)
	if isErr || text != "Created docs/decisions/ADR-001-use-stdio.md (Proposed)." {
		t.Fatalf("adr_create = %q, error %v", text, isErr)
	}
	text, _ = call(t, session, "adr_list", map[string]any{})
	if text != "ADR-001  Proposed  Use stdio\n" {
		t.Fatalf("adr_list = %q", text)
	}
	text, isErr = call(t, session, "adr_update", map[string]any{"id": "ADR-001", "status": "Done"})
	if !isErr || !strings.HasPrefix(text, "Status \"Done\" is not valid.") {
		t.Fatalf("adr_update with a bad status = %q, error %v", text, isErr)
	}
	text, _ = call(t, session, "adr_get", map[string]any{"id": "ADR-001", "section": "Alternatives"})
	if text != "- HTTP: more work" {
		t.Fatalf("adr_get = %q", text)
	}
}

func TestToolSchemaRequiresFields(t *testing.T) {
	session := connect(t, newStore(t))
	text, isErr := call(t, session, "adr_create", map[string]any{"title": "X"})
	if !isErr || !strings.Contains(text, "missing properties") {
		t.Fatalf("adr_create with missing fields = %q, error %v", text, isErr)
	}
}
