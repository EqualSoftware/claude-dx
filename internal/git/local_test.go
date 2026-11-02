package git

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// local makes a repository with a committer, for the tools that commit.
func local(t *testing.T) Runner {
	t.Helper()
	r := repo(t)
	git(t, r.Root, "config", "user.name", "Tester")
	git(t, r.Root, "config", "user.email", "t@example.com")
	git(t, r.Root, "config", "commit.gpgsign", "false")
	return r
}

func head(t *testing.T, r Runner) string {
	t.Helper()
	out, err := r.Run(context.Background(), r.Root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func TestMessageRefusals(t *testing.T) {
	ok := Message{Type: "feat", Summary: "add a thing"}
	tests := []struct {
		name string
		edit func(*Message)
		want string
	}{
		{"type", func(m *Message) { m.Type = "feature" }, "is not one of"},
		{"scope", func(m *Message) { m.Scope = "A b" }, "scope"},
		{"period", func(m *Message) { m.Summary = "add a thing." }, "period"},
		{"upper", func(m *Message) { m.Summary = "Add a thing" }, "lowercase"},
		{"empty", func(m *Message) { m.Summary = " " }, "not empty"},
		{"past", func(m *Message) { m.Summary = "added a thing" }, "imperative"},
		{"ing", func(m *Message) { m.Summary = "adding a thing" }, "imperative"},
		{"third", func(m *Message) { m.Summary = "fixes a thing" }, "imperative"},
		{"long", func(m *Message) { m.Summary = "add " + strings.Repeat("x", 80) }, "maximum is 72"},
		{"coauthor", func(m *Message) { m.CoAuthors = []string{"nobody"} }, "Name <email>"},
	}
	for _, tt := range tests {
		m := ok
		tt.edit(&m)
		if _, err := m.Args(); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want %q", tt.name, err, tt.want)
		}
	}
	if _, err := (Message{Type: "fix", Summary: "embed the file"}).Args(); err != nil {
		t.Errorf("embed: %v", err)
	}
}

func TestMessageArgs(t *testing.T) {
	m := Message{Type: "feat", Scope: "cli", Summary: "add x", Body: "why", Breaking: true, CoAuthors: []string{"A B <a@b.c>", "C D <c@d.e>"}}
	args, err := m.Args()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-m", "feat(cli)!: add x", "-m", "why", "-m",
		"BREAKING CHANGE: add x\nCo-Authored-By: A B <a@b.c>\nCo-Authored-By: C D <c@d.e>"}
	if strings.Join(args, "|") != strings.Join(want, "|") {
		t.Errorf("args = %q, want %q", args, want)
	}
}

func TestStageUnstageDiscard(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	write(t, r.Root, "a.txt", "changed\n")
	write(t, r.Root, "new.txt", "n\n")
	if got, err := r.Stage(ctx, r.Root, []string{"a.txt"}, false); err != nil || got != "Staged 1 paths." {
		t.Fatalf("Stage = %q, %v", got, err)
	}
	if got, err := r.Unstage(ctx, r.Root, []string{"a.txt"}); err != nil || got != "Unstaged 1 paths." {
		t.Fatalf("Unstage = %q, %v", got, err)
	}
	if _, err := r.Discard(ctx, r.Root, []string{"new.txt"}); err == nil || !strings.Contains(err.Error(), "not tracked") {
		t.Fatalf("Discard untracked err = %v", err)
	}
	if got, err := r.Discard(ctx, r.Root, []string{"a.txt"}); err != nil || got != "Discarded 1 paths." {
		t.Fatalf("Discard = %q, %v", got, err)
	}
	st, _ := r.Status(ctx, r.Root)
	if strings.Contains(st, "a.txt") {
		t.Errorf("a.txt is still changed: %s", st)
	}
}

func TestStageRefusal(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	if _, err := r.Stage(ctx, r.Root, nil, false); err == nil {
		t.Error("Stage with no paths must refuse")
	}
	if got, err := r.Stage(ctx, r.Root, nil, true); err != nil || got != "Staged all changes." {
		t.Errorf("Stage all = %q, %v", got, err)
	}
}

func TestCommit(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	m := Message{Type: "fix", Scope: "cli", Summary: "change a", Breaking: true, CoAuthors: []string{"A B <a@b.c>"}}
	if _, err := r.Commit(ctx, r.Root, m, nil); err == nil || !strings.Contains(err.Error(), "nothing is staged") {
		t.Fatalf("Commit with nothing staged err = %v", err)
	}
	write(t, r.Root, "a.txt", "changed\n")
	got, err := r.Commit(ctx, r.Root, m, []string{"a.txt"})
	if err != nil || !strings.HasSuffix(got, " fix(cli)!: change a") {
		t.Fatalf("Commit = %q, %v", got, err)
	}
	msg, _ := r.Run(ctx, r.Root, "log", "-1", "--format=%B")
	contains(t, msg, "BREAKING CHANGE: change a\nCo-Authored-By: A B <a@b.c>")
}

func TestAmendUndoFixup(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	first := head(t, r)
	got, err := r.Amend(ctx, r.Root, Message{Type: "feat", Summary: "add b file"}, nil, nil)
	if err != nil || !strings.HasSuffix(got, " feat: add b file.") || head(t, r) == first {
		t.Fatalf("Amend = %q, %v", got, err)
	}
	if _, err := r.Fixup(ctx, r.Root, "HEAD~1", nil); err == nil || !strings.Contains(err.Error(), "nothing is staged") {
		t.Fatalf("Fixup with nothing staged err = %v", err)
	}
	write(t, r.Root, "b.txt", "bb\n")
	git(t, r.Root, "add", "b.txt")
	if got, err := r.Fixup(ctx, r.Root, "HEAD", nil); err != nil || !strings.Contains(got, "fixup! feat: add b file") {
		t.Fatalf("Fixup = %q, %v", got, err)
	}
	undoTwo(t, r)
}

func undoTwo(t *testing.T, r Runner) {
	t.Helper()
	ctx := context.Background()
	got, err := r.Undo(ctx, r.Root, 2, nil)
	if err != nil || !strings.HasSuffix(got, " feat: add a.") {
		t.Fatalf("Undo = %q, %v", got, err)
	}
	if _, err := r.Undo(ctx, r.Root, -1, nil); err == nil {
		t.Error("Undo with n < 1 must refuse")
	}
}

func TestRevertCherryPickTagStash(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	git(t, r.Root, "switch", "-c", "side")
	write(t, r.Root, "s.txt", "s\n")
	git(t, r.Root, "add", "s.txt")
	git(t, r.Root, "commit", "-m", "feat: add s")
	pick := head(t, r)
	git(t, r.Root, "switch", "main")
	if got, err := r.CherryPick(ctx, r.Root, []string{pick}); err != nil || !strings.Contains(got, "feat: add s") {
		t.Fatalf("CherryPick = %q, %v", got, err)
	}
	if got, err := r.Revert(ctx, r.Root, "HEAD", false); err != nil || !strings.Contains(got, `Revert "feat: add s"`) {
		t.Fatalf("Revert = %q, %v", got, err)
	}
	if got, err := r.Revert(ctx, r.Root, "HEAD~1", true); err != nil || got != "Reverted HEAD~1 in the index." {
		t.Fatalf("Revert no commit = %q, %v", got, err)
	}
	if _, err := r.Revert(ctx, r.Root, "--abort", false); err == nil {
		t.Error("Revert must refuse a dash")
	}
	tagAndStash(t, r)
}

func tagAndStash(t *testing.T, r Runner) {
	t.Helper()
	ctx := context.Background()
	if got, _ := r.Tag(ctx, r.Root, TagOptions{}); got != "No tags." {
		t.Errorf("Tag list = %q", got)
	}
	if got, err := r.Tag(ctx, r.Root, TagOptions{Action: "create", Name: "v1", Message: "one"}); err != nil || got != "Created tag v1." {
		t.Errorf("Tag create = %q, %v", got, err)
	}
	if got, _ := r.Tag(ctx, r.Root, TagOptions{Action: "list"}); got != "v1\n" {
		t.Errorf("Tag list = %q", got)
	}
	if got, err := r.Tag(ctx, r.Root, TagOptions{Action: "delete", Name: "v1"}); err != nil || got != "Deleted tag v1." {
		t.Errorf("Tag delete = %q, %v", got, err)
	}
	if _, err := r.Tag(ctx, r.Root, TagOptions{Action: "create", Name: "-x"}); err == nil {
		t.Error("Tag must refuse a dash")
	}
	stash(t, r)
}

func stash(t *testing.T, r Runner) {
	t.Helper()
	ctx := context.Background()
	git(t, r.Root, "reset", "--hard", "HEAD")
	write(t, r.Root, "a.txt", "stashed\n")
	if _, err := r.Stash(ctx, r.Root, StashOptions{Action: "push", Message: "wip"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.Stash(ctx, r.Root, StashOptions{}); !strings.Contains(got, "wip") {
		t.Errorf("Stash list = %q", got)
	}
	if _, err := r.Stash(ctx, r.Root, StashOptions{Action: "pop"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.Stash(ctx, r.Root, StashOptions{}); got != "No stash entries." {
		t.Errorf("Stash list = %q", got)
	}
}

func TestRewriteOnRemote(t *testing.T) {
	r := pushed(t)
	git(t, r.Root, "config", "user.name", "Tester")
	git(t, r.Root, "config", "user.email", "t@example.com")
	before := head(t, r)
	ctx := context.Background()
	no := func(_ context.Context, q string) (bool, error) {
		contains(t, q, "is on the remote. git_", " rewrites it. Do you allow this?")
		return false, nil
	}
	for name, confirm := range map[string]Confirm{"no": no, "none": nil} {
		if _, err := r.Undo(ctx, r.Root, 1, confirm); err == nil || !strings.Contains(err.Error(), "refused") {
			t.Errorf("%s: Undo err = %v", name, err)
		}
		if _, err := r.Amend(ctx, r.Root, Message{}, nil, confirm); err == nil {
			t.Errorf("%s: Amend must refuse", name)
		}
		if _, err := r.Fixup(ctx, r.Root, "HEAD", confirm); err == nil {
			t.Errorf("%s: Fixup must refuse", name)
		}
		if head(t, r) != before {
			t.Fatalf("%s: HEAD changed", name)
		}
	}
	yes := func(context.Context, string) (bool, error) { return true, nil }
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	git(t, r.Root, "push", "origin", "main")
	if got, err := r.Undo(ctx, r.Root, 1, yes); err != nil || !strings.HasSuffix(got, " feat: add a.") {
		t.Errorf("Undo with yes = %q, %v", got, err)
	}
}

// TestElicitation runs git_undo through MCP with a client that declines, and with a client with no elicitation.
func TestElicitation(t *testing.T) {
	r := pushed(t)
	before := head(t, r)
	decline := func(context.Context, *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		return &mcp.ElicitResult{Action: "decline"}, nil
	}
	for name, opts := range map[string]*mcp.ClientOptions{"decline": {ElicitationHandler: decline}, "none": nil} {
		text, isErr := callUndo(t, r, opts)
		if !isErr || !strings.Contains(text, "refused") {
			t.Errorf("%s: text = %q, error %v", name, text, isErr)
		}
		if head(t, r) != before {
			t.Fatalf("%s: HEAD changed", name)
		}
	}
}

func callUndo(t *testing.T, r Runner, opts *mcp.ClientOptions) (string, bool) {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	Register(server, r.Root)
	st, ct := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, opts).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "git_undo", Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	return res.Content[0].(*mcp.TextContent).Text, res.IsError
}
