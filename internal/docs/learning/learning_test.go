package learning

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newStore(t *testing.T) Store {
	t.Helper()
	return Store{Dir: filepath.Join(t.TempDir(), "docs", "learnings"), Lock: filepath.Join(t.TempDir(), "docs.lock")}
}

// connect returns a client session to a server with the learning tools of s, over memory.
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

const wantFile = `---
topic: go-embed-paths
description: 'go: embed paths use forward slashes. Use when you embed files.'
---

Use forward slashes in go:embed paths on all platforms.
`

func TestCreate(t *testing.T) {
	s := newStore(t)
	if err := s.Create("go-embed-paths", "go: embed paths use forward slashes.\nUse when you embed files.", "Use forward slashes in go:embed paths on all platforms.\n\n"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(s.Dir, "go-embed-paths.md"))
	if string(data) != wantFile {
		t.Fatalf("file =\n%s\nwant\n%s", data, wantFile)
	}
	// AC-2: a topic that exists is refused.
	if err := s.Create("go-embed-paths", "D.", "B."); err == nil || !strings.Contains(err.Error(), "learning_update") {
		t.Fatalf("second Create = %v", err)
	}
	if err := s.Create("Go Embed", "D.", "B."); err == nil {
		t.Fatal("Create accepted a topic that is not kebab-case")
	}
}

// TestMissingDir is AC-3: with no docs/learnings, the tools give a message and make no directory.
func TestMissingDir(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	for name, args := range map[string]map[string]any{"learning_list": {}, "learning_delete": {"topic": "x"}} {
		text, isErr := call(t, session, name, args)
		if isErr || text != "No learnings. docs/learnings does not exist." {
			t.Errorf("%s = %q, error %v", name, text, isErr)
		}
	}
	if _, err := os.Stat(s.Dir); err == nil {
		t.Fatal("a read or delete made docs/learnings")
	}
}

// expect calls the tool and stops the test if the result is not want, or if isErr is not wantErr.
func expect(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, want string, wantErr bool) {
	t.Helper()
	if text, isErr := call(t, session, name, args); isErr != wantErr || text != want {
		t.Fatalf("%s(%v) = %q, error %v; want %q", name, args, text, isErr, want)
	}
}

func TestToolsFlow(t *testing.T) {
	session := connect(t, newStore(t))
	expect(t, session, "learning_create", map[string]any{"topic": "x", "description": "About x.", "body": "Lesson one."}, "Created docs/learnings/x.md.", false)
	expect(t, session, "learning_create", map[string]any{"topic": "x", "description": "D.", "body": "B."}, `Learning "x" exists. Use learning_update to change it.`, true)
	expect(t, session, "learning_update", map[string]any{"topic": "x", "body": "Lesson one.\n\nLesson two."}, "Updated learning x.", false)
	expect(t, session, "learning_update", map[string]any{"topic": "x", "description": "About x and y."}, "Updated learning x.", false)
	expect(t, session, "learning_list", map[string]any{}, "x: About x and y.\n", false)
	expect(t, session, "learning_get", map[string]any{"topic": "x"}, "---\ntopic: x\ndescription: About x and y.\n---\n\nLesson one.\n\nLesson two.\n", false)
	expect(t, session, "learning_delete", map[string]any{"topic": "x"}, "Deleted learning x.", false)
	expect(t, session, "learning_get", map[string]any{"topic": "x"}, `Learning "x" does not exist. Use learning_list.`, true)
	expect(t, session, "learning_list", map[string]any{}, "No learnings in docs/learnings.", false)
}

func TestUpdateNeedsAChange(t *testing.T) {
	s := newStore(t)
	if err := s.Create("x", "D.", "B."); err != nil {
		t.Fatal(err)
	}
	if err := s.Update("x", "", ""); err == nil {
		t.Fatal("Update with no change did not fail")
	}
}

// TestListBadFrontmatter checks docs-adr#R-9.
func TestListBadFrontmatter(t *testing.T) {
	s := newStore(t)
	for _, topic := range []string{"a-topic", "b-topic"} {
		if err := s.Create(topic, "About "+topic+".", "Lesson."); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(s.Dir, "bad.md"), []byte("---\ndescription: [\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	text, err := s.listText()
	if err != nil || !strings.HasPrefix(text, "a-topic: About a-topic.\nb-topic: About b-topic.\nWarning: bad.md: the frontmatter is not valid") {
		t.Fatalf("learning_list = %q, %v", text, err)
	}
	if err := s.WriteIndex(); err != nil {
		t.Fatalf("WriteIndex = %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(s.Dir, indexFile))
	if strings.Contains(string(data), "bad") {
		t.Fatalf("README.md has the bad file:\n%s", data)
	}
}

func TestIndex(t *testing.T) {
	s := newStore(t)
	for _, topic := range []string{"b-topic", "a-topic"} {
		if err := s.Create(topic, "About "+topic+".", "Lesson."); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Update("b-topic", "About b.", ""); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(s.Dir, "README.md")
	data, _ := os.ReadFile(readme)
	want := indexHeader + "| [a-topic](a-topic.md) | About a-topic. |\n| [b-topic](b-topic.md) | About b. |\n"
	if string(data) != want {
		t.Fatalf("README.md =\n%s\nwant\n%s", data, want)
	}
	list, _, _ := s.List()
	if len(list) != 2 {
		t.Fatalf("List has README.md: %v", list)
	}
	if err := s.Delete("a-topic"); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(readme)
	if strings.Contains(string(data), "a-topic") {
		t.Fatalf("README.md after delete =\n%s", data)
	}
}

// TestIndexEscapesPipes checks docs-common#R-6 for learnings.
func TestIndexEscapesPipes(t *testing.T) {
	s := newStore(t)
	if err := s.Create("a-topic", "About a|b.", "Lesson."); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if want := "| [a-topic](a-topic.md) | About a\\|b. |\n"; !strings.HasSuffix(string(data), want) {
		t.Fatalf("README.md =\n%s\nwant suffix %s", data, want)
	}
}

func TestCreateRefusesReadme(t *testing.T) {
	s := newStore(t)
	for _, topic := range []string{"readme", "README"} {
		if err := s.Create(topic, "D.", "B."); err == nil || !strings.Contains(err.Error(), "index") {
			t.Fatalf("Create(%q) = %v", topic, err)
		}
	}
}
