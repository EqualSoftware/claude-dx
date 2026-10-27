package spec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newStore returns a store in a new git repository with ADR-008.
func newStore(t *testing.T) Store {
	t.Helper()
	root := t.TempDir()
	git(t, root, "init", "-q")
	adrs := filepath.Join(root, "docs", "decisions")
	if err := os.MkdirAll(adrs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(adrs, "ADR-008-specs.md"), []byte("# ADR-008\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return Store{
		Dir:    filepath.Join(root, "docs", "specs"),
		ADRDir: adrs,
		Root:   root,
		Now:    func() time.Time { return time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) },
		Author: func() string { return "Tester" },
		Lock:   filepath.Join(t.TempDir(), "docs.lock"),
	}
}

// git runs git in dir with a test identity and no commit signature.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	args = append([]string{"-C", dir, "-c", "user.name=Tester", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// connect returns a client session to a server with the spec tools of s, over memory.
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

// step is one tool call and its expected reply. An error reply has want with the prefix "ERR ".
type step struct {
	name string
	args map[string]any
	want string
}

func run(t *testing.T, session *mcp.ClientSession, steps []step) {
	t.Helper()
	for i, s := range steps {
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: s.name, Arguments: s.args})
		if err != nil {
			t.Fatalf("step %d %s: %v", i, s.name, err)
		}
		got := res.Content[0].(*mcp.TextContent).Text
		if res.IsError {
			got = "ERR " + got
		}
		if got != s.want {
			t.Fatalf("step %d %s = %q, want %q", i, s.name, got, s.want)
		}
	}
}

// filled are bodies for all sections that In Review needs.
var filled = map[string]string{"1": "S.", "2": "P.", "3": "- G-1: G.", "4": "- O.", "6": "D.", "7": "- R-1: T."}

func createArgs(slug, title string) map[string]any {
	return map[string]any{"slug": slug, "title": title, "milestone": "v0.1.0", "priority": "P1", "area": []string{"docs"}, "related": []string{"ADR-008"}}
}

func TestCreateAndList(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	long := strings.Repeat("a", 41)
	run(t, session, []step{
		{"spec_list", map[string]any{}, "No specs. docs/specs does not exist."},
		{"spec_get", map[string]any{"slug": "x"}, "No specs. docs/specs does not exist."},
		{"spec_create", createArgs("Docs_Lock", "Docs lock"), `ERR Slug "Docs_Lock" is not kebab-case. Example: docs-lock.`},
		{"spec_create", createArgs("general", "Docs lock"), `ERR Slug "general" is reserved.`},
		{"spec_create", createArgs(long, "Docs lock"), "ERR The slug has 41 characters. The limit is 40. Write a shorter slug."},
		{"spec_create", map[string]any{"slug": "a", "title": "A", "milestone": "1.0", "priority": "P1", "area": []string{}},
			`ERR Milestone "1.0" is not valid. Use vMAJOR.MINOR.PATCH, example v0.1.0.`},
		{"spec_create", map[string]any{"slug": "a", "title": "A", "milestone": "v0.1.0", "priority": "P9", "area": []string{}},
			`ERR Priority "P9" is not valid. Use one of: P0, P1, P2, P3.`},
		{"spec_create", map[string]any{"slug": "a", "title": "A", "milestone": "v0.1.0", "priority": "P1", "area": []string{}, "related": []string{"ADR-9"}},
			`ERR Related "ADR-9": ADR-009 does not exist.`},
		{"spec_create", createArgs("docs-lock", "Docs lock"), "Created docs-lock."},
		{"spec_create", createArgs("docs-lock", "Other"), `ERR Spec "docs-lock" exists.`},
		{"spec_create", createArgs("other", "DOCS LOCK"), "ERR Docs-lock has the same title. Give a different title, or change docs-lock."},
		{"spec_create", map[string]any{"slug": "cache", "title": "Cache", "milestone": "v0.2.0", "priority": "P2", "area": []string{"cli"}, "split_from": "Old"},
			`ERR Slug "Old" is not kebab-case. Example: docs-lock.`},
		{"spec_create", map[string]any{"slug": "cache", "title": "Cache", "milestone": "v0.2.0", "priority": "P2", "area": []string{"cli"}, "split_from": "old-cache"}, "Created cache."},
		{"spec_list", map[string]any{}, "cache      Draft  P2  v0.2.0  Cache      [cli]   0/0\ndocs-lock  Draft  P1  v0.1.0  Docs lock  [docs]  0/0\n"},
		{"spec_list", map[string]any{"related": "adr-008"}, "docs-lock  Draft  P1  v0.1.0  Docs lock  [docs]  0/0\n"},
		{"spec_list", map[string]any{"status": "Approved"}, "No specs match.\n"},
		{"spec_get", map[string]any{"slug": "nope"}, `ERR Spec "nope" does not exist.`},
		{"spec_get", map[string]any{"slug": "docs-lock", "section": "99"}, `ERR Section "99" does not exist. Use a number (example: 5) or a title.`},
		{"spec_get", map[string]any{"slug": "docs-lock", "section": "Goals"}, "- G-1: [Measurable result.]"},
	})
	doc, err := s.Get("docs-lock", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"slug: docs-lock\n", "title: Docs lock\n", "status: Draft\n", "related: [ADR-008]\n", "\n# Docs lock\n",
		"| 2026-10-26 | Tester | Initial draft specification |"} {
		if !strings.Contains(doc, want) {
			t.Errorf("SPEC.md has no %q:\n%s", want, doc)
		}
	}
	if text, _ := s.Get("cache", "9"); !strings.Contains(text, "| 2026-10-26 | Tester | Initial draft specification. Split from old-cache. |") {
		t.Errorf("cache history =\n%s", text)
	}
	index, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	rows := "| [docs-lock](docs-lock/SPEC.md) | Docs lock | `Draft` | v0.1.0 | `P1` | `docs` | 0/0 |\n" +
		"| [cache](cache/SPEC.md) | Cache | `Draft` | v0.2.0 | `P2` | `cli` | 0/0 |\n"
	if string(index) != indexHeader+rows {
		t.Errorf("README.md =\n%s", index)
	}
}

func TestGates(t *testing.T) {
	s := newStore(t)
	session := connect(t, s)
	approved := "docs-lock is Approved."
	review := " The scope changed, so docs-lock is In Review. The user must approve it again."
	steps := []step{
		{"spec_create", createArgs("docs-lock", "Docs lock"), "Created docs-lock."},
		{"spec_update_section", map[string]any{"slug": "docs-lock", "section": "5", "body": "x"}, `ERR Section "5" changes only with spec_req_add and spec_req_set.`},
		{"spec_update_section", map[string]any{"slug": "docs-lock", "section": "open questions", "body": "x"}, `ERR Section "open questions" changes only with spec_question_add and spec_question_answer.`},
		{"spec_update_section", map[string]any{"slug": "docs-lock", "section": "9", "body": "x"}, `ERR Section "9" changes only with spec_set_status.`},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Done"}, `ERR Status "Done" is not valid. Use one of: Draft, In Review, Approved.`},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Approved"}, "ERR Cannot set Approved from Draft. Allowed: In Review."},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "In Review"}, "ERR Cannot set In Review: section 1 has the template text. Fill it or write Not applicable; " +
			"section 2 has the template text. Fill it or write Not applicable; section 3 has the template text. Fill it or write Not applicable; " +
			"section 4 has the template text. Fill it or write Not applicable; section 6 has the template text. Fill it or write Not applicable; " +
			"section 7 has the template text. Fill it or write Not applicable; section 5 has no requirement. Add one with spec_req_add."},
	}
	for _, ref := range []string{"1", "2", "3", "4", "6", "7"} {
		steps = append(steps, step{"spec_update_section", map[string]any{"slug": "docs-lock", "section": ref, "body": filled[ref]}, "Updated section " + ref + " of docs-lock."})
	}
	steps = append(steps, []step{
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": " "}, "ERR Give the text of the requirement."},
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": "Each write holds\nthe lock | always."}, "Added R-1."},
		{"spec_question_add", map[string]any{"slug": "docs-lock", "question": "Which lock?"}, "Added Q-1."},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "In Review"}, "ERR Cannot set In Review: Q-1 is Open."},
		{"spec_question_answer", map[string]any{"slug": "docs-lock", "qid": "Q-7", "answer": "A."}, "ERR Question Q-7 does not exist."},
		{"spec_question_answer", map[string]any{"slug": "docs-lock", "qid": "q-1", "answer": "The docs lock."}, "Answered q-1."},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "In Review"}, "docs-lock is In Review."},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Approved", "note": "The user approved it."}, approved},
		{"spec_update_section", map[string]any{"slug": "docs-lock", "section": "1", "body": "New summary."}, "Updated section 1 of docs-lock."},
		{"spec_update_section", map[string]any{"slug": "docs-lock", "section": "Goals", "body": "- G-1: New."}, "Updated section Goals of docs-lock." + review},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Approved"}, approved},
		{"spec_req_set", map[string]any{"slug": "docs-lock", "rid": "R-1"}, "ERR Give text, status or both."},
		{"spec_req_set", map[string]any{"slug": "docs-lock", "rid": "R-1", "status": "Gone"}, `ERR Status "Gone" is not valid. Use one of: Planned, Done.`},
		{"spec_req_set", map[string]any{"slug": "docs-lock", "rid": "R-9", "status": "Done"}, "ERR Requirement R-9 does not exist."},
		{"spec_req_set", map[string]any{"slug": "docs-lock", "rid": "r-1", "status": "Done"}, "Updated R-1."},
		{"spec_req_set", map[string]any{"slug": "docs-lock", "rid": "R-1", "text": "Each write holds the docs lock."}, "Updated R-1." + review},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Approved"}, approved},
		{"spec_req_add", map[string]any{"slug": "docs-lock", "text": "Each read holds the shared lock."}, "Added R-2." + review},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Approved"}, approved},
		{"spec_question_add", map[string]any{"slug": "docs-lock", "question": "Q?", "owner": "Equal"}, "Added Q-2. Q-2 is Open, so docs-lock is In Review. The user must approve it again."},
		{"spec_set_status", map[string]any{"slug": "docs-lock", "status": "Draft"}, "docs-lock is Draft."},
		{"spec_list", map[string]any{}, "docs-lock  Draft  P1  v0.1.0  Docs lock  [docs]  1/2\n"},
	}...)
	run(t, session, steps)
	text, err := s.Get("docs-lock", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"| R-1 | Each write holds the docs lock. | Done |\n| R-2 | Each read holds the shared lock. | Planned |\n",
		"| Q-1 | Which lock? | Tester | The docs lock. |\n| Q-2 | Q? | Equal | Open |\n",
		"| Added R-1. |", "| Status: In Review. |", "| The user approved it. |", "| Status: In Review. The scope changed in section Goals. |",
		"| R-1: status Done. |", "| R-1: text changed. |", "| Status: In Review. The scope changed in section 5. |",
		"| Added R-2. |", "| Status: In Review. Q-2 is Open. |", "| Status: Draft. |",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("SPEC.md has no %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "None.") {
		t.Errorf("SPEC.md has a placeholder row:\n%s", text)
	}
}

// TestIndexMatchesRepository makes README.md for the specs of this repository and compares it with
// the committed README.md.
func TestIndexMatchesRepository(t *testing.T) {
	repo := filepath.Join("..", "..", "..", "docs", "specs")
	want, err := os.ReadFile(filepath.Join(repo, "README.md"))
	if err != nil {
		t.Skip("no repository specs")
	}
	s := Store{Dir: repo}
	specs, _, err := s.List(nil)
	if err != nil {
		t.Fatal(err)
	}
	s.Dir = t.TempDir()
	for _, sp := range specs {
		data, _ := os.ReadFile(filepath.Join(repo, sp.Front.Slug, fileName))
		writeFile(t, s.path(sp.Front.Slug), string(data))
	}
	if err := s.WriteIndex(); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(s.Dir, "README.md"))
	if string(got) != string(want) {
		t.Fatalf("README.md differs from the committed file.\ngot:\n%s", got)
	}
}
