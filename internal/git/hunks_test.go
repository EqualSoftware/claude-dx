package git

import (
	"context"
	"strings"
	"testing"
)

// twoHunks makes a.txt with two far hunks changed in the working tree.
func twoHunks(t *testing.T) Runner {
	t.Helper()
	r := repo(t)
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, "line"+strings.Repeat("x", i))
	}
	write(t, r.Root, "a.txt", strings.Join(lines, "\n")+"\n")
	git(t, r.Root, "commit", "-am", "chore: grow a")
	lines[1] = "changed top"
	lines[28] = "changed bottom"
	write(t, r.Root, "a.txt", strings.Join(lines, "\n")+"\n")
	return r
}

func tokenOf(t *testing.T, out string) string {
	t.Helper()
	line, _, _ := strings.Cut(out, "\n")
	tok, ok := strings.CutPrefix(line, "Token: ")
	if !ok {
		t.Fatalf("no token in %q", out)
	}
	return tok
}

func TestDiffHunks(t *testing.T) {
	r := twoHunks(t)
	ctx := context.Background()
	plain, err := r.Diff(ctx, r.Root, DiffOptions{Hunks: false})
	if err != nil || strings.Contains(plain, "Token:") || strings.Contains(plain, "[1]") {
		t.Fatalf("plain Diff = %q, %v", plain, err)
	}
	got, err := r.Diff(ctx, r.Root, DiffOptions{Hunks: true})
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "Token: ", "[1] @@ ", "[2] @@ ", "changed top", "changed bottom")
	again, _ := r.Diff(ctx, r.Root, DiffOptions{Hunks: true})
	if got != again {
		t.Error("the token is not stable for the same diff")
	}
	if _, err := r.Diff(ctx, r.Root, DiffOptions{Hunks: true, Staged: true}); err == nil {
		t.Error("hunks with staged was accepted")
	}
}

func TestStageHunks(t *testing.T) {
	r := twoHunks(t)
	ctx := context.Background()
	out, _ := r.Diff(ctx, r.Root, DiffOptions{Hunks: true})
	tok := tokenOf(t, out)
	if _, err := r.StageHunks(ctx, r.Root, "000000000000", []HunkSelection{{Path: "a.txt", Numbers: []int{1}}}); err == nil || !strings.Contains(err.Error(), "token") {
		t.Errorf("wrong token err = %v", err)
	}
	if _, err := r.StageHunks(ctx, r.Root, tok, []HunkSelection{{Path: "a.txt", Numbers: []int{3}}}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("missing number err = %v", err)
	}
	if staged, _ := r.Run(ctx, r.Root, "diff", "--cached"); staged != "" {
		t.Fatalf("a refusal changed the index:\n%s", staged)
	}
	if _, err := r.StageHunks(ctx, r.Root, tok, []HunkSelection{{Path: "a.txt", Numbers: []int{2}}}); err != nil {
		t.Fatal(err)
	}
	staged, _ := r.Run(ctx, r.Root, "diff", "--cached")
	contains(t, staged, "changed bottom")
	if strings.Contains(staged, "changed top") {
		t.Errorf("hunk 1 was staged:\n%s", staged)
	}
	rest, _ := r.Run(ctx, r.Root, "diff")
	contains(t, rest, "changed top")
	if _, err := r.StageHunks(ctx, r.Root, tok, []HunkSelection{{Path: "a.txt", Numbers: []int{1}}}); err == nil {
		t.Error("the old token was accepted after the diff changed")
	}
}

func TestStageHunksUntracked(t *testing.T) {
	r := twoHunks(t)
	write(t, r.Root, "new.txt", "x\n")
	out, _ := r.Diff(context.Background(), r.Root, DiffOptions{Hunks: true})
	_, err := r.StageHunks(context.Background(), r.Root, tokenOf(t, out), []HunkSelection{{Path: "new.txt", Numbers: []int{1}}})
	if err == nil || !strings.Contains(err.Error(), "git_stage") {
		t.Errorf("untracked err = %v", err)
	}
}
