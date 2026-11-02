package git

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// pushed makes a repository whose main branch is pushed to a bare remote.
func pushed(t *testing.T) Runner {
	t.Helper()
	r := repo(t)
	bare := t.TempDir()
	git(t, bare, "init", "--bare", "-b", "main")
	git(t, r.Root, "remote", "add", "origin", bare)
	git(t, r.Root, "push", "-u", "origin", "main")
	return r
}

func TestOnRemote(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	if on, err := r.OnRemote(ctx, r.Root, "HEAD"); err != nil || !on {
		t.Fatalf("OnRemote(pushed) = %v, %v", on, err)
	}
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	if on, err := r.OnRemote(ctx, r.Root, "HEAD"); err != nil || on {
		t.Fatalf("OnRemote(local) = %v, %v", on, err)
	}
}

func TestGuardRemote(t *testing.T) {
	r := pushed(t)
	ctx := context.Background()
	yes := func(context.Context, string) (bool, error) { return true, nil }
	no := func(context.Context, string) (bool, error) { return false, nil }
	fail := func(context.Context, string) (bool, error) { return false, errors.New("no elicitation") }
	tests := []struct {
		name    string
		confirm Confirm
		want    string
	}{
		{"yes", yes, ""},
		{"no", no, "user said no"},
		{"unsupported", nil, "cannot ask"},
		{"error", fail, "was not asked"},
	}
	for _, tt := range tests {
		err := r.GuardRemote(ctx, r.Root, "HEAD", "git_undo", "q?", tt.confirm)
		if tt.want == "" && err != nil || tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("%s: err = %v, want %q", tt.name, err, tt.want)
		}
	}
}
