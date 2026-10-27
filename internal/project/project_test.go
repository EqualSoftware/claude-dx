package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// newRepo makes a git repository with one commit and returns its resolved path.
func newRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	run(t, dir, "init", "-q", "-b", "main")
	// No signature: the global git config of a developer can sign commits with an external agent.
	run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "init")
	return dir
}

func TestOpenFindsRootFromSubdir(t *testing.T) {
	repo := newRepo(t)
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := Open(sub)
	if err != nil {
		t.Fatal(err)
	}
	if p.Root != repo {
		t.Fatalf("Root = %q, want %q", p.Root, repo)
	}
	if len(p.Key) != 16 {
		t.Fatalf("Key = %q, want 16 hex characters", p.Key)
	}
}

func TestOpenOutsideGit(t *testing.T) {
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	p, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Root != dir {
		t.Fatalf("Root = %q, want %q", p.Root, dir)
	}
}

func TestWorktreesShareKey(t *testing.T) {
	repo := newRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	run(t, repo, "worktree", "add", "-q", "-b", "other", wt)
	main, err := Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	other, err := Open(wt)
	if err != nil {
		t.Fatal(err)
	}
	if main.Key != other.Key {
		t.Fatalf("keys differ: %q and %q", main.Key, other.Key)
	}
	if main.Root == other.Root {
		t.Fatal("worktrees must have different roots")
	}
}

func TestResolve(t *testing.T) {
	repo := newRepo(t)
	p, err := Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	ok := []string{"docs/x.md", "new/dir/file", ".", filepath.Join(repo, "a")}
	for _, rel := range ok {
		if _, err := p.Resolve(rel); err != nil {
			t.Errorf("Resolve(%q) = %v, want no error", rel, err)
		}
	}
	bad := []string{"..", "../x", "a/../../x", filepath.Dir(repo)}
	for _, rel := range bad {
		if _, err := p.Resolve(rel); err == nil {
			t.Errorf("Resolve(%q) = no error, want outside-project error", rel)
		}
	}
}

func TestResolveRefusesSymlinkOutside(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need extra privileges on Windows")
	}
	repo := newRepo(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	p, err := Open(repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"link", "link/new-file"} {
		if _, err := p.Resolve(rel); err == nil {
			t.Errorf("Resolve(%q) through a symlink to outside = no error", rel)
		}
	}
}
