package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreeAddListRemove(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	wt := filepath.Join(t.TempDir(), "wt")
	got, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "add", Worktree: wt, Branch: "feature"})
	if err != nil || got != "Added worktree "+wt+" on the new branch feature from HEAD." {
		t.Fatalf("add = %q, %v", got, err)
	}
	list, err := r.Worktree(ctx, r.Root, WorktreeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	contains(t, list, r.Root+" ", " main\n", " feature\n")
	if _, err := r.Dir(ctx, wt); err != nil {
		t.Errorf("the new worktree is refused by Dir: %v", err)
	}
	got, err = r.Worktree(ctx, r.Root, WorktreeOptions{Action: "remove", Worktree: wt, DeleteBranch: true})
	if err != nil || got != "Removed worktree "+wt+" and deleted branch feature." {
		t.Fatalf("remove = %q, %v", got, err)
	}
	if _, err := os.Stat(wt); err == nil {
		t.Error("the worktree directory is still there")
	}
	if b, _ := r.Branch(ctx, r.Root, BranchOptions{}, nil); strings.Contains(b, "feature") {
		t.Errorf("the branch feature is still there:\n%s", b)
	}
}

func TestWorktreeAddStartPoint(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	first := head(t, r)
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	wt := filepath.Join(t.TempDir(), "wt")
	if _, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "add", Worktree: wt, Branch: "old", Start: first}); err != nil {
		t.Fatal(err)
	}
	out, _ := r.Run(ctx, wt, "rev-parse", "HEAD")
	if strings.TrimSpace(out) != first {
		t.Errorf("worktree HEAD = %s, want %s", out, first)
	}
}

func TestWorktreeRemoveRefusals(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	wt := filepath.Join(t.TempDir(), "wt")
	if _, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "add", Worktree: wt, Branch: "feature"}); err != nil {
		t.Fatal(err)
	}
	write(t, wt, "new.txt", "n\n")
	if _, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "remove", Worktree: wt}); err == nil || !strings.Contains(err.Error(), "has changes that are not committed") {
		t.Fatalf("remove with untracked file err = %v", err)
	}
	if _, err := os.Stat(wt); err != nil {
		t.Errorf("a refused remove deleted the worktree: %v", err)
	}
	for _, o := range []WorktreeOptions{
		{Action: "remove", Worktree: r.Root},
		{Action: "remove", Worktree: t.TempDir()},
		{Action: "add", Worktree: wt},
		{Action: "add", Worktree: wt, Branch: "-x"},
		{Action: "add", Branch: "x"},
		{Action: "move", Worktree: wt},
	} {
		if _, err := r.Worktree(ctx, r.Root, o); err == nil || !strings.HasPrefix(err.Error(), "git_worktree refused: ") {
			t.Errorf("%v: err = %v, want a refusal", o, err)
		}
	}
}

func TestWorktreeRemoveKeepsUnmergedBranch(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	wt := filepath.Join(t.TempDir(), "wt")
	if _, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "add", Worktree: wt, Branch: "feature"}); err != nil {
		t.Fatal(err)
	}
	write(t, wt, "w.txt", "w\n")
	git(t, wt, "add", "w.txt")
	git(t, wt, "commit", "-m", "feat: add w")
	got, err := r.Worktree(ctx, r.Root, WorktreeOptions{Action: "remove", Worktree: wt, DeleteBranch: true})
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "Removed worktree "+wt+". The branch feature was not deleted: ", "not fully merged")
}
