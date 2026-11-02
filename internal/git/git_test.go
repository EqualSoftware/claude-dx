package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/report"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=Tester", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)
	if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// repo makes a repository with one commit that has the file a.txt.
func repo(t *testing.T) Runner {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-b", "main")
	write(t, dir, "a.txt", "one\ntwo\n")
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "feat: add a")
	return Runner{Root: dir}
}

func contains(t *testing.T, got string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("output =\n%s\nwant it to contain %q", got, w)
		}
	}
}

func TestStatus(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	got, err := r.Status(ctx, r.Root)
	if err != nil || got != "## main\nClean.\n" {
		t.Fatalf("Status = %q, %v", got, err)
	}
	write(t, r.Root, "a.txt", "one\ntwo\nthree\n")
	write(t, r.Root, "b.txt", "b\n")
	write(t, r.Root, "c d.txt", "c\n")
	git(t, r.Root, "add", "b.txt")
	got, err = r.Status(ctx, r.Root)
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "## main\n", "Staged:\n  A b.txt\n", "Unstaged:\n  M a.txt\n", "Untracked:\n  ? c d.txt\n")
}

func TestStatusConflict(t *testing.T) {
	r := repo(t)
	git(t, r.Root, "checkout", "-b", "other")
	write(t, r.Root, "a.txt", "other\n")
	git(t, r.Root, "commit", "-a", "-m", "feat: other")
	git(t, r.Root, "checkout", "main")
	write(t, r.Root, "a.txt", "main\n")
	git(t, r.Root, "commit", "-a", "-m", "feat: main")
	_ = exec.Command("git", "-C", r.Root, "-c", "user.name=T", "-c", "user.email=t@e.c", "merge", "other").Run()
	got, err := r.Status(context.Background(), r.Root)
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "Conflicted:\n  UU a.txt\n")
}

func TestFormatStatusUpstream(t *testing.T) {
	got := formatStatus("# branch.oid abc\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +1 -2\n2 R. N... 100644 100644 100644 a b R100 new.txt\told.txt\n")
	want := "## main...origin/main +1 -2\nStaged:\n  R new.txt <- old.txt\n"
	if got != want {
		t.Fatalf("formatStatus = %q, want %q", got, want)
	}
}

func TestDiff(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	if got, _ := r.Diff(ctx, r.Root, DiffOptions{}); got != "No changes." {
		t.Fatalf("Diff = %q", got)
	}
	write(t, r.Root, "a.txt", "one\ntwo\nthree\n")
	got, err := r.Diff(ctx, r.Root, DiffOptions{Paths: []string{"a.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, "+three")
	if strings.Contains(got, "\x1b") {
		t.Error("diff has color")
	}
	got, _ = r.Diff(ctx, r.Root, DiffOptions{Stat: true})
	contains(t, got, "a.txt", "1 insertion")
	git(t, r.Root, "add", "a.txt")
	if got, _ = r.Diff(ctx, r.Root, DiffOptions{}); got != "No changes." {
		t.Fatalf("unstaged Diff = %q", got)
	}
	got, _ = r.Diff(ctx, r.Root, DiffOptions{Staged: true})
	contains(t, got, "+three")
	if _, err = r.Diff(ctx, r.Root, DiffOptions{Ref: "--output=x"}); err == nil {
		t.Error("Diff accepted a ref that starts with a dash")
	}
}

func TestLog(t *testing.T) {
	r := repo(t)
	write(t, r.Root, "b.txt", "b\n")
	git(t, r.Root, "add", "b.txt")
	git(t, r.Root, "commit", "-m", "feat: add b")
	ctx := context.Background()
	got, err := r.Log(ctx, r.Root, LogOptions{})
	if err != nil || len(strings.Split(strings.TrimSpace(got), "\n")) != 2 {
		t.Fatalf("Log = %q, %v", got, err)
	}
	first := strings.Split(got, "\n")[0]
	if !strings.HasSuffix(first, " Tester feat: add b") || len(strings.Fields(first)[0]) < 7 || len(strings.Fields(first)[1]) != 10 {
		t.Errorf("first line = %q", first)
	}
	got, _ = r.Log(ctx, r.Root, LogOptions{N: 1})
	if strings.Count(got, "\n") != 1 {
		t.Errorf("Log n=1 = %q", got)
	}
	got, _ = r.Log(ctx, r.Root, LogOptions{Paths: []string{"a.txt"}})
	contains(t, got, "feat: add a")
	if strings.Contains(got, "add b") {
		t.Errorf("Log with a path = %q", got)
	}
}

func TestShow(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	got, err := r.Show(ctx, r.Root, "", false)
	if err != nil {
		t.Fatal(err)
	}
	contains(t, got, " Tester feat: add a\n", "+one")
	got, _ = r.Show(ctx, r.Root, "HEAD", true)
	contains(t, got, "a.txt", "2 insertions")
	if strings.Contains(got, "+one") {
		t.Errorf("Show stat has a diff: %q", got)
	}
}

func TestBlame(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	got, err := r.Blame(ctx, r.Root, BlameOptions{File: "a.txt"})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 || !strings.HasSuffix(lines[0], " Tester 1: one") || !strings.HasSuffix(lines[1], " Tester 2: two") {
		t.Fatalf("Blame = %q", got)
	}
	got, _ = r.Blame(ctx, r.Root, BlameOptions{File: "a.txt", Start: 2})
	if strings.Count(got, "\n") != 1 || !strings.HasSuffix(strings.TrimSpace(got), "2: two") {
		t.Errorf("Blame from line 2 = %q", got)
	}
	got, _ = r.Blame(ctx, r.Root, BlameOptions{File: "a.txt", End: 1})
	if strings.Count(got, "\n") != 1 {
		t.Errorf("Blame to line 1 = %q", got)
	}
	if _, err = r.Blame(ctx, r.Root, BlameOptions{}); err == nil {
		t.Error("Blame accepted no file")
	}
}

func TestErrorText(t *testing.T) {
	r := repo(t)
	_, err := r.Show(context.Background(), r.Root, "nosuchref", false)
	if err == nil || !strings.HasPrefix(err.Error(), "git show failed: ") || strings.Count(err.Error(), "\n") > errorLines-1 {
		t.Fatalf("error = %v", err)
	}
}

func TestEnvironment(t *testing.T) {
	r := repo(t)
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	_, err := r.Show(context.Background(), r.Root, "nosuchref", false)
	if err == nil || !strings.Contains(err.Error(), "bad revision") {
		t.Fatalf("error = %v, want English text", err)
	}
}

func TestDir(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	if dir, err := r.Dir(ctx, ""); err != nil || dir != r.Root {
		t.Fatalf("Dir empty = %q, %v", dir, err)
	}
	if dir, err := r.Dir(ctx, r.Root); err != nil || dir != r.Root {
		t.Fatalf("Dir root = %q, %v", dir, err)
	}
	wt := filepath.Join(filepath.Dir(r.Root), "wt")
	git(t, r.Root, "worktree", "add", "-b", "feature", wt)
	if dir, err := r.Dir(ctx, wt); err != nil || dir != wt {
		t.Fatalf("Dir worktree = %q, %v", dir, err)
	}
	if got, err := r.Status(ctx, wt); err != nil || !strings.HasPrefix(got, "## feature") {
		t.Fatalf("Status in worktree = %q, %v", got, err)
	}
}

func TestDirRefused(t *testing.T) {
	r := repo(t)
	ctx := context.Background()
	for _, p := range []string{t.TempDir(), filepath.Join(r.Root, "nodir"), filepath.Join(r.Root, ".git")} {
		_, err := r.Dir(ctx, p)
		if err == nil || !strings.Contains(err.Error(), "is not the project root or a worktree of this repository") {
			t.Errorf("Dir(%q) error = %v", p, err)
		}
	}
}

func TestRunNetworkTimeout(t *testing.T) {
	r := repo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.RunNetwork(ctx, r.Root, "status"); err == nil {
		t.Fatal("RunNetwork ignored a canceled context")
	}
	if NetworkTimeout.Seconds() != 120 {
		t.Errorf("NetworkTimeout = %v", NetworkTimeout)
	}
}

// TestExec checks git-tools#R-16: git_exec runs git and logs the call.
func TestExec(t *testing.T) {
	r := repo(t)
	log := filepath.Join(t.TempDir(), "cache", report.LogName)
	out, err := r.Exec(context.Background(), r.Root, log, execInput{Args: []string{"rev-parse", "--abbrev-ref", "HEAD"}, Reason: "no tool"})
	if err != nil || !strings.HasPrefix(out, "main\n") {
		t.Fatalf("Exec = %q, %v", out, err)
	}
	if !strings.Contains(out, "claude-dx report") || !strings.Contains(out, report.IssueURL) {
		t.Errorf("Exec has no notice: %q", out)
	}
	checkExecLog(t, r, log)
}

// checkExecLog runs a failing command, and checks that the log has both calls.
func checkExecLog(t *testing.T, r Runner, log string) {
	t.Helper()
	if _, err := r.Exec(context.Background(), r.Root, log, execInput{Args: []string{"nope"}, Reason: "bad"}); err == nil || !strings.Contains(err.Error(), "git nope failed") {
		t.Fatalf("Exec error = %v", err)
	}
	entries, err := report.Read(log)
	if err != nil || len(entries) != 2 || entries[1].Sub != "nope" {
		t.Fatalf("log = %+v, %v", entries, err)
	}
	if entries[0].Sub != "rev-parse" || entries[0].Reason != "no tool" {
		t.Errorf("first entry = %+v", entries[0])
	}
}

func TestExecNeedsArgsAndReason(t *testing.T) {
	r := repo(t)
	log := filepath.Join(t.TempDir(), report.LogName)
	for _, in := range []execInput{{Reason: "x"}, {Args: []string{"status"}}} {
		if _, err := r.Exec(context.Background(), r.Root, log, in); err == nil {
			t.Errorf("Exec(%+v) gave no error", in)
		}
	}
	if _, err := os.Stat(log); err == nil {
		t.Error("an invalid call was logged")
	}
}
