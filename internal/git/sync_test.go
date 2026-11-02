package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// push2 makes a commit in a second clone and pushes it to the remote of r. The commit sets file to text.
func push2(t *testing.T, r Runner, file, text string) {
	t.Helper()
	remote, err := r.Run(context.Background(), r.Root, "remote", "get-url", "origin")
	if err != nil {
		t.Fatal(err)
	}
	clone := filepath.Join(t.TempDir(), "clone")
	git(t, filepath.Dir(clone), "clone", strings.TrimSpace(remote), clone)
	write(t, clone, file, text)
	git(t, clone, "add", file)
	git(t, clone, "commit", "-m", "feat: change "+file)
	git(t, clone, "push", "origin", "main")
}

func TestSyncKeepsLocalChanges(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	push2(t, r, "b.txt", "b\n")
	write(t, r.Root, "a.txt", "one\ntwo\nlocal\n")
	got, err := r.Sync(ctx, r.Root)
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "Synced main onto origin/main.", "HEAD is now", "feat: change b.txt")
	if data, err := os.ReadFile(filepath.Join(r.Root, "a.txt")); err != nil || string(data) != "one\ntwo\nlocal\n" {
		t.Errorf("a.txt = %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(r.Root, "b.txt")); err != nil {
		t.Errorf("b.txt: %v", err)
	}
}

func TestSyncRefusals(t *testing.T) {
	r := local(t)
	if _, err := r.Sync(context.Background(), r.Root); err == nil || !strings.Contains(err.Error(), "git_sync refused: the branch main has no upstream") {
		t.Errorf("no upstream err = %v", err)
	}
	git(t, r.Root, "checkout", "--detach")
	if _, err := r.Sync(context.Background(), r.Root); err == nil || !strings.Contains(err.Error(), "detached") {
		t.Errorf("detached err = %v", err)
	}
}

// diverged makes a local commit and a remote commit that change a.txt in different ways.
func diverged(t *testing.T) Runner {
	t.Helper()
	r := pushed(t)
	push2(t, r, "a.txt", "remote\n")
	write(t, r.Root, "a.txt", "local\n")
	git(t, r.Root, "commit", "-am", "feat: change a locally")
	return r
}

func TestSyncConflictAndRebase(t *testing.T) {
	r := diverged(t)
	ctx := context.Background()
	got, err := r.Sync(ctx, r.Root)
	if err != nil {
		t.Fatalf("Sync err = %v, want a normal result", err)
	}
	contains(t, got, "git_sync stopped: conflicts. Resolve them, stage the files, then run git_rebase action=continue.",
		"a.txt: hunk 1 lines 1-5", "<<<<<<<", "remote", "=======", "local", ">>>>>>>")
	if _, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "start", Onto: "main"}, nil); err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Errorf("start during rebase err = %v", err)
	}
	write(t, r.Root, "a.txt", "both\n")
	git(t, r.Root, "add", "a.txt")
	got, err = r.Rebase(ctx, r.Root, RebaseOptions{Action: "continue"}, nil)
	if err != nil || !strings.HasPrefix(got, "Rebase finished. HEAD is now ") || !strings.HasSuffix(got, "feat: change a locally.") {
		t.Fatalf("continue = %q, %v", got, err)
	}
	if _, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "continue"}, nil); err == nil || !strings.Contains(err.Error(), "no rebase is in progress") {
		t.Errorf("continue with no rebase err = %v", err)
	}
}

func TestRebaseAbortSkip(t *testing.T) {
	ctx := context.Background()
	r := diverged(t)
	before := head(t, r)
	git(t, r.Root, "fetch")
	if got, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "start", Onto: "origin/main"}, nil); err != nil || !strings.Contains(got, "stopped: conflicts") {
		t.Fatalf("start = %q, %v", got, err)
	}
	if got, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "abort"}, nil); err != nil || got != "Rebase aborted." || head(t, r) != before {
		t.Fatalf("abort = %q, %v", got, err)
	}
	r.Rebase(ctx, r.Root, RebaseOptions{Action: "start", Onto: "origin/main"}, nil)
	got, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "skip"}, nil)
	if err != nil || !strings.HasPrefix(got, "Rebase finished. HEAD is now ") {
		t.Fatalf("skip = %q, %v", got, err)
	}
}

func TestRebaseRefusals(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	for _, o := range []RebaseOptions{{Action: "start"}, {Action: "start", Onto: "-x"}, {Action: "squash"}, {Action: "abort"}} {
		if _, err := r.Rebase(ctx, r.Root, o, nil); err == nil || !strings.HasPrefix(err.Error(), "git_rebase refused: ") {
			t.Errorf("%v: err = %v", o, err)
		}
	}
}

func TestRebaseStartGuardsRemote(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	git(t, r.Root, "branch", "base", "HEAD")
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	git(t, r.Root, "push", "origin", "main")
	// The new commit is on the remote, and rebase onto base would rewrite it.
	git(t, r.Root, "checkout", "-q", "base")
	git(t, r.Root, "commit", "--allow-empty", "-m", "feat: move base")
	git(t, r.Root, "checkout", "-q", "main")
	before := head(t, r)
	if _, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "start", Onto: "base"}, nil); err == nil || !strings.Contains(err.Error(), "git_rebase refused") || head(t, r) != before {
		t.Fatalf("start with no confirm err = %v", err)
	}
	yes := func(context.Context, string) (bool, error) { return true, nil }
	if got, err := r.Rebase(ctx, r.Root, RebaseOptions{Action: "start", Onto: "base"}, yes); err != nil || head(t, r) == before {
		t.Fatalf("start with yes = %q, %v", got, err)
	}
}

func TestConflictsNone(t *testing.T) {
	r := local(t)
	if got, err := r.Conflicts(context.Background(), r.Root); err != nil || got != "No conflicts." {
		t.Errorf("Conflicts = %q, %v", got, err)
	}
}

func TestHunksOfDiff3(t *testing.T) {
	text := "x\n<<<<<<< HEAD\nours\n||||||| base\nold\n=======\ntheirs\n>>>>>>> topic\ny\n<<<<<<< HEAD\na\n=======\nb\n>>>>>>> topic\n"
	h := hunksOf(text)
	if len(h) != 2 || h[0].start != 2 || h[0].end != 8 || h[1].start != 10 || h[1].end != 14 {
		t.Fatalf("hunks = %+v", h)
	}
	contains(t, h[0].text, "||||||| base", "old", ">>>>>>> topic")
}
