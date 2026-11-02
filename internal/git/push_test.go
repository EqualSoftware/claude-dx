package git

import (
	"context"
	"strings"
	"testing"
)

func remoteHead(t *testing.T, r Runner, branch string) string {
	t.Helper()
	out, err := r.Run(context.Background(), r.Root, "ls-remote", "origin", "refs/heads/"+branch)
	if err != nil {
		t.Fatal(err)
	}
	hash, _, _ := strings.Cut(out, "\t")
	return hash
}

func TestPushNewBranchSetsUpstream(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	git(t, r.Root, "switch", "-c", "topic")
	write(t, r.Root, "t.txt", "t\n")
	git(t, r.Root, "add", "t.txt")
	git(t, r.Root, "commit", "-m", "feat: add t")
	got, err := r.Push(ctx, r.Root, PushOptions{}, nil)
	if err != nil || got != "Pushed topic to origin. Set the upstream to origin/topic." {
		t.Fatalf("Push = %q, %v", got, err)
	}
	if remoteHead(t, r, "topic") != head(t, r) {
		t.Error("the remote branch topic is not at HEAD")
	}
	write(t, r.Root, "u.txt", "u\n")
	git(t, r.Root, "add", "u.txt")
	git(t, r.Root, "commit", "-m", "feat: add u")
	if got, err := r.Push(ctx, r.Root, PushOptions{}, nil); err != nil || got != "Pushed topic to origin." {
		t.Fatalf("second Push = %q, %v", got, err)
	}
}

func TestPushForce(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	git(t, r.Root, "switch", "-c", "topic")
	write(t, r.Root, "t.txt", "t\n")
	git(t, r.Root, "add", "t.txt")
	git(t, r.Root, "commit", "-m", "feat: add t")
	if _, err := r.Push(ctx, r.Root, PushOptions{}, nil); err != nil {
		t.Fatal(err)
	}
	git(t, r.Root, "commit", "--amend", "-m", "feat: add t file")
	got, err := r.Push(ctx, r.Root, PushOptions{}, nil)
	if err != nil || got != "Pushed topic to origin with force-with-lease." || remoteHead(t, r, "topic") != head(t, r) {
		t.Fatalf("force Push = %q, %v", got, err)
	}
}

func TestPushForceMain(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	git(t, r.Root, "commit", "--amend", "-m", "feat: add a file")
	before := remoteHead(t, r, "main")
	if _, err := r.Push(ctx, r.Root, PushOptions{}, nil); err == nil || !strings.Contains(err.Error(), "git_push refused: main is the main branch and the client cannot ask the user") {
		t.Fatalf("no confirm err = %v", err)
	}
	no := func(context.Context, string) (bool, error) { return false, nil }
	if _, err := r.Push(ctx, r.Root, PushOptions{}, no); err == nil || !strings.Contains(err.Error(), "the user said no") {
		t.Fatalf("no err = %v", err)
	}
	if remoteHead(t, r, "main") != before {
		t.Error("a refused push changed the remote")
	}
	var asked string
	yes := func(_ context.Context, q string) (bool, error) { asked = q; return true, nil }
	got, err := r.Push(ctx, r.Root, PushOptions{}, yes)
	if err != nil || got != "Pushed main to origin with force-with-lease." || remoteHead(t, r, "main") != head(t, r) {
		t.Fatalf("yes Push = %q, %v", got, err)
	}
	contains(t, asked, "The branch main is the main branch. git_push must force push it with --force-with-lease. Do you allow this?")
}

// A remote commit that the branch does not include is never overwritten, also after the user says yes.
func TestPushKeepsRemoteCommits(t *testing.T) {
	r := pushed(t)
	push2(t, r, "b.txt", "b\n")
	write(t, r.Root, "c.txt", "c\n")
	git(t, r.Root, "add", "c.txt")
	git(t, r.Root, "commit", "-m", "feat: add c")
	before := remoteHead(t, r, "main")
	yes := func(context.Context, string) (bool, error) { return true, nil }
	if _, err := r.Push(context.Background(), r.Root, PushOptions{}, yes); err == nil || !strings.Contains(err.Error(), "git push failed") {
		t.Fatalf("err = %v, want a git push failure", err)
	}
	if remoteHead(t, r, "main") != before {
		t.Error("the push overwrote a remote commit")
	}
}

func TestPushRefusals(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	if _, err := r.Push(ctx, r.Root, PushOptions{Remote: "-x"}, nil); err == nil || !strings.HasPrefix(err.Error(), "git_push refused: ") {
		t.Errorf("dash err = %v", err)
	}
	git(t, r.Root, "checkout", "--detach")
	if _, err := r.Push(ctx, r.Root, PushOptions{}, nil); err == nil || !strings.Contains(err.Error(), "detached") {
		t.Errorf("detached err = %v", err)
	}
}
