package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/EqualSoftware/claude-dx/internal/cache"
)

// tools is the full list of tool names. A new tool needs a change of this list.
var tools = []string{
	"adr_create", "adr_get", "adr_list", "adr_supersede", "adr_update",
	"glossary_delete", "glossary_list", "glossary_lookup", "glossary_upsert",
	"learning_create", "learning_delete", "learning_get", "learning_list", "learning_update",
	"spec_create", "spec_delete", "spec_get", "spec_list", "spec_question_add", "spec_question_answer", "spec_req_add", "spec_req_set",
	"spec_set_status", "spec_update_fields", "spec_update_section",
	"task_create", "task_delete", "task_dod_add", "task_dod_check", "task_dod_remove", "task_dod_uncheck", "task_get", "task_list", "task_move",
	"task_set_status", "task_step_check", "task_update_fields", "task_update_section",
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
	return serve(t, repo)
}

// serve starts the dx server for the project that contains dir and returns a client session over memory.
func serve(t *testing.T, dir string) *mcp.ClientSession {
	t.Helper()
	server, err := New("test", dir)
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
		{"task_create", map[string]any{"spec": "general", "slug": "add-a-cache", "title": "Add a cache", "priority": "P1",
			"brief": "Add a cache.", "plan": []string{"Add it. Check: go test ./..."}}, "Created task general/add-a-cache."},
		{"task_list", map[string]any{}, "general/add-a-cache  Todo  P1  Add a cache\n"},
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

// callText calls a tool, checks that it succeeds and returns its text.
func callText(ctx context.Context, session *mcp.ClientSession, name string, args map[string]any) (string, error) {
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	if res.IsError {
		return "", fmt.Errorf("%s: tool error: %s", name, text)
	}
	return text, nil
}

// TestToolsParallel checks mcp-server#R-9: parallel tool calls give no error, no data race and no lost write.
func TestToolsParallel(t *testing.T) {
	session := connect(t)
	ctx := context.Background()
	const workers = 8
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := strconv.Itoa(i)
			calls := []struct {
				name string
				args map[string]any
			}{
				{"task_create", map[string]any{"spec": "general", "slug": "task-" + n, "title": "Task " + n, "priority": "P1",
					"brief": "Do " + n + ".", "plan": []string{"Do it. Check: go test ./..."}}},
				{"adr_create", map[string]any{"title": "Decision " + n, "context": "C.", "decision": "D.", "evidence": "E.",
					"alternatives": []map[string]any{{"option": "A", "reason": "more work"}},
					"consequences": map[string]any{"positive": "P.", "negative": "N.", "follow_up": "None."}}},
				{"glossary_upsert", map[string]any{"term": "Term" + n, "definition": "A term.", "area": "process", "source": "`docs/`"}},
				{"learning_create", map[string]any{"topic": "topic-" + n, "description": "About " + n + ".", "body": "Lesson."}},
				{"task_list", map[string]any{}},
				{"adr_list", map[string]any{}},
			}
			for _, c := range calls {
				if _, err := callText(ctx, session, c.name, c.args); err != nil {
					t.Errorf("worker %d: %v", i, err)
				}
			}
		}()
	}
	wg.Wait()
	for name, want := range map[string]int{"task_list": workers, "adr_list": workers, "glossary_list": workers, "learning_list": workers} {
		text, err := callText(ctx, session, name, map[string]any{})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Count(text, "\n"); got != want {
			t.Errorf("%s has %d lines, want %d:\n%s", name, got, want, text)
		}
	}
	text, err := callText(ctx, session, "adr_list", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= workers; n++ {
		if !strings.Contains(text, fmt.Sprintf("ADR-%03d ", n)) {
			t.Errorf("adr_list has no ADR-%03d:\n%s", n, text)
		}
	}
}

// TestDataDirAndProjectFile checks project-data#R-2, R-3: the server uses the docs lock and the
// counter in <git-common-dir>/dx, and writes the Project file (cache).
func TestDataDirAndProjectFile(t *testing.T) {
	session := connect(t)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "learning_create", Arguments: map[string]any{
		"topic": "x", "description": "d", "body": "b",
	}})
	if err != nil || res.IsError {
		t.Fatalf("learning_create = %v, %+v", err, res)
	}
	cacheDir, _ := cache.Dir()
	repo := filepath.Join(os.Getenv("HOME"), "repo") // connect sets HOME to its temporary directory
	if _, err := os.Stat(filepath.Join(repo, ".git", "dx", "docs.lock")); err != nil {
		t.Fatalf("docs.lock is not in .git/dx: %v", err)
	}
	files, _ := filepath.Glob(filepath.Join(cacheDir, "projects", "*", "project.json"))
	if len(files) != 1 {
		t.Fatalf("Project files (cache) = %v, want 1", files)
	}
	var f struct {
		Path     string `json:"path"`
		LastUsed string `json:"last_used"`
	}
	data, _ := os.ReadFile(files[0])
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.EvalSymlinks(repo)
	if _, err := time.Parse(time.RFC3339, f.LastUsed); err != nil || f.Path != root || !strings.HasSuffix(f.LastUsed, "Z") {
		t.Fatalf("Project file (cache) = %+v, %v; want path %q and an RFC 3339 UTC date", f, err, root)
	}
}

// TestWorktreesShareTasks checks docs-task: two worktrees of one repository see the same tasks.
func TestWorktreesShareTasks(t *testing.T) {
	main := connect(t)
	repo := filepath.Join(os.Getenv("HOME"), "repo") // connect sets HOME to its temporary directory
	other := filepath.Join(os.Getenv("HOME"), "other")
	for _, args := range [][]string{
		{"-C", repo, "-c", "commit.gpgsign=false", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "first"},
		{"-C", repo, "worktree", "add", "-q", other},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	res, err := main.CallTool(context.Background(), &mcp.CallToolParams{Name: "task_create", Arguments: map[string]any{
		"spec": "general", "slug": "shared", "title": "Shared", "priority": "P1", "brief": "B.", "plan": []string{"S."},
	}})
	if err != nil || res.IsError {
		t.Fatalf("task_create = %v, %+v", err, res)
	}
	res, err = serve(t, other).CallTool(context.Background(), &mcp.CallToolParams{Name: "task_list", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if text := res.Content[0].(*mcp.TextContent).Text; text != "general/shared  Todo  P1  Shared\n" {
		t.Fatalf("task_list in the other worktree = %q", text)
	}
	if _, err := os.Stat(filepath.Join(repo, ".git", "dx", "tasks", "general", "shared.md")); err != nil {
		t.Fatalf("the task is not in .git/dx/tasks: %v", err)
	}
}

// TestPluginTextsNameRegisteredTools checks docs-task: each task tool name in the plugin texts is a registered tool.
func TestPluginTextsNameRegisteredTools(t *testing.T) {
	root := filepath.Join("..", "plugin", "assets")
	session := connect(t)
	registered := map[string]bool{}
	for tool, err := range session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		registered[tool.Name] = true
	}
	name := regexp.MustCompile(`task_[a-z_]*[a-z]`)
	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		files++
		data, err := os.ReadFile(path)
		for _, m := range name.FindAllString(string(data), -1) {
			if !registered[m] {
				t.Errorf("%s names the tool %q, which the server does not register", path, m)
			}
		}
		return err
	})
	if err != nil || files == 0 {
		t.Fatalf("walk %s: %d files, error %v", root, files, err)
	}
}

// TestProjectFileKeepsCountersOutsideGit checks ADR-007: outside git, the Project file (cache) and the
// Project file (data) are one file. After a start, an allocation and a second start, it has all keys.
func TestProjectFileKeepsCountersOutsideGit(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "cache"))
	dir := filepath.Join(tmp, "plain")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	session := serve(t, dir)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "adr_create", Arguments: map[string]any{
		"title": "Use stdio", "context": "C.", "decision": "D.", "evidence": "E.",
		"alternatives": []map[string]any{{"option": "HTTP", "reason": "more work"}},
		"consequences": map[string]any{"positive": "P.", "negative": "N.", "follow_up": "None."},
	}})
	if err != nil || res.IsError {
		t.Fatalf("adr_create = %v, %+v", err, res)
	}
	serve(t, dir) // a second start writes path and last_used again
	cacheDir, _ := cache.Dir()
	files, _ := filepath.Glob(filepath.Join(cacheDir, "projects", "*", "project.json"))
	if len(files) != 1 {
		t.Fatalf("project files = %v, want 1", files)
	}
	var f map[string]json.RawMessage
	data, _ := os.ReadFile(files[0])
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"path", "last_used", "counters"} {
		if _, ok := f[key]; !ok {
			t.Errorf("project.json has no %q: %s", key, data)
		}
	}
}
