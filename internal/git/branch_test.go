package git

import (
	"context"
	"strings"
	"testing"
)

func TestBranch(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	steps := []struct {
		o    BranchOptions
		want string
	}{
		{BranchOptions{Action: "create", Name: "topic"}, "Created branch topic."},
		{BranchOptions{Action: "switch", Name: "topic"}, "Switched to branch topic."},
		{BranchOptions{Action: "rename", Name: "topic", NewName: "work"}, "Renamed branch topic to work."},
		{BranchOptions{Action: "switch", Name: "main"}, "Switched to branch main."},
		{BranchOptions{Action: "create", Name: "other", Start: "main"}, "Created branch other."},
	}
	for _, s := range steps {
		if got, err := r.Branch(ctx, r.Root, s.o, nil); err != nil || got != s.want {
			t.Fatalf("%v = %q, %v; want %q", s.o, got, err, s.want)
		}
	}
	got, err := r.Branch(ctx, r.Root, BranchOptions{}, nil)
	contains(t, got, "* main ", "  work ", "  other ", "feat: add a")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := r.Branch(ctx, r.Root, BranchOptions{Action: "delete", Name: "other"}, nil); err != nil || got != "Deleted branch other." {
		t.Fatalf("delete = %q, %v", got, err)
	}
}

func TestBranchRefusals(t *testing.T) {
	r := local(t)
	ctx := context.Background()
	tests := []BranchOptions{
		{Action: "create", Name: "-x"},
		{Action: "create"},
		{Action: "rename", Name: "main"},
		{Action: "delete", Name: "main"},
		{Action: "merge", Name: "main"},
	}
	for _, o := range tests {
		if _, err := r.Branch(ctx, r.Root, o, nil); err == nil || !strings.HasPrefix(err.Error(), "git_branch refused: ") {
			t.Errorf("%v: err = %v, want a refusal", o, err)
		}
	}
}

// unmerged makes the branch "wip" with one commit that main does not have.
func unmerged(t *testing.T) Runner {
	t.Helper()
	r := local(t)
	git(t, r.Root, "switch", "-c", "wip")
	write(t, r.Root, "w.txt", "w\n")
	git(t, r.Root, "add", "w.txt")
	git(t, r.Root, "commit", "-m", "feat: add w")
	git(t, r.Root, "switch", "main")
	return r
}

func TestBranchDeleteNotMerged(t *testing.T) {
	ctx := context.Background()
	del := BranchOptions{Action: "delete", Name: "wip"}
	yes := func(context.Context, string) (bool, error) { return true, nil }
	no := func(context.Context, string) (bool, error) { return false, nil }
	tests := []struct {
		name    string
		confirm Confirm
		want    string
		deleted bool
	}{
		{"nil", nil, "cannot ask", false},
		{"no", no, "user said no", false},
		{"yes", yes, "", true},
	}
	for _, tt := range tests {
		r := unmerged(t)
		got, err := r.Branch(ctx, r.Root, del, tt.confirm)
		if tt.want == "" && (err != nil || got != "Deleted branch wip.") || tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
			t.Errorf("%s: %q, %v; want %q", tt.name, got, err, tt.want)
		}
		list, _ := r.Branch(ctx, r.Root, BranchOptions{}, nil)
		if strings.Contains(list, " wip ") == tt.deleted {
			t.Errorf("%s: branch list =\n%s", tt.name, list)
		}
	}
}
