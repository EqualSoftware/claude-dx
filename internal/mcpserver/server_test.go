package mcpserver

import (
	"context"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/cache"
)

// tools is the full list of tool names. A new tool needs a change of this list.
var tools = []string{
	"adr_create", "adr_get", "adr_list", "adr_supersede", "adr_update",
	"glossary_delete", "glossary_list", "glossary_lookup", "glossary_upsert",
	"learning_create", "learning_delete", "learning_get", "learning_list", "learning_update",
	"task_create", "task_dod_check", "task_get", "task_list", "task_question_add", "task_question_answer",
	"task_renumber", "task_set_status", "task_update_fields", "task_update_section",
}

// connect starts the dx server in a new git repository with a temporary cache directory,
// and returns a client session over memory.
func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "cache"))
	cacheDir, err := cache.Dir()
	if err != nil || !strings.HasPrefix(cacheDir, tmp) {
		t.Fatalf("cache.Dir() = %q, %v; want a path under %q", cacheDir, err, tmp)
	}
	repo := filepath.Join(tmp, "repo")
	for _, args := range [][]string{{"init", repo}, {"-C", repo, "config", "user.name", "Tester"}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	server, err := New("test", repo)
	if err != nil {
		t.Fatal(err)
	}
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

func TestToolNames(t *testing.T) {
	session := connect(t)
	var names []string
	for tool, err := range session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	if !slices.Equal(names, tools) {
		t.Fatalf("tools =\n%v\nwant\n%v", names, tools)
	}
}

func TestToolsFlow(t *testing.T) {
	session := connect(t)
	steps := []struct {
		name string
		args map[string]any
		want string
	}{
		{"task_create", map[string]any{"title": "Add a cache", "priority": "P1", "type": "feat", "area": []string{"cli"}},
			"Created docs/tasks/TASK-001-add-a-cache.md (Draft)."},
		{"task_list", map[string]any{}, "TASK-001  Draft  P1  Add a cache  [cli]  deps:None\n"},
		{"adr_create", map[string]any{
			"title": "Use stdio", "context": "C.", "decision": "D.", "evidence": "E.",
			"alternatives": []map[string]any{{"option": "HTTP", "reason": "more work"}},
			"consequences": map[string]any{"positive": "P.", "negative": "N.", "follow_up": "None."},
		}, "Created docs/decisions/ADR-001-use-stdio.md (Proposed)."},
		{"adr_list", map[string]any{}, "ADR-001  Proposed  Use stdio\n"},
		{"glossary_upsert", map[string]any{"term": "Task", "definition": "A spec file.", "area": "process", "source": "`docs/tasks/`"},
			`Added "Task".`},
		{"glossary_list", map[string]any{}, "| Task | A spec file. | process | — | `docs/tasks/` |\n"},
		{"learning_create", map[string]any{"topic": "x", "description": "About x.", "body": "Lesson one."},
			"Created docs/learnings/x.md."},
		{"learning_list", map[string]any{}, "x: About x.\n"},
	}
	for _, s := range steps {
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: s.name, Arguments: s.args})
		if err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		text := res.Content[0].(*mcp.TextContent).Text
		if res.IsError || text != s.want {
			t.Fatalf("%s = %q, error %v; want %q", s.name, text, res.IsError, s.want)
		}
	}
}
