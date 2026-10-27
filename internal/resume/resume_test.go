package resume

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// env is a temporary repository with a temporary cache directory.
type env struct {
	t     *testing.T
	repo  string
	cache string
	data  string // the data directory of the project
}

// runGit runs a git command in dir.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=T", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newEnv makes a repository on the branch main with one commit.
func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{t: t, repo: t.TempDir(), cache: t.TempDir()}
	runGit(t, e.repo, "init", "-q", "-b", "main")
	runGit(t, e.repo, "commit", "-q", "--allow-empty", "-m", "chore: first")
	p, err := project.Open(e.repo)
	if err != nil {
		t.Fatal(err)
	}
	e.data = p.DataDir(e.cache)
	if err := os.MkdirAll(filepath.Join(e.data, project.Tasks), 0o755); err != nil {
		t.Fatal(err)
	}
	return e
}

// writeFile writes a file and creates its folder.
func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// task writes the task id. body is the text after the title.
func (e *env) task(id, status, priority, deps, body string, age time.Duration) {
	e.t.Helper()
	sp, slug, _ := strings.Cut(id, "/")
	path := filepath.Join(e.data, project.Tasks, sp, slug+".md")
	text := fmt.Sprintf("---\nid: %s\ntitle: Title %s\nstatus: %s\npriority: %s\ncovers: []\ndepends_on: [%s]\n---\n# Title %s\n\n%s",
		id, slug, status, priority, deps, slug, body)
	writeFile(e.t, path, text)
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		e.t.Fatal(err)
	}
}

// spec writes a spec with the status.
func (e *env) spec(slug, status string) {
	e.t.Helper()
	writeFile(e.t, filepath.Join(e.repo, "docs", "specs", slug, "SPEC.md"),
		fmt.Sprintf("---\nslug: %s\ntitle: T\nstatus: %s\nmilestone: v1\npriority: P1\narea: []\nrelated: []\n---\n# T\n", slug, status))
}

// run returns the output of Run.
func (e *env) run() string {
	e.t.Helper()
	var out bytes.Buffer
	Run(&out, e.repo, e.cache)
	return out.String()
}

func check(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("output:\n%s\nwant:\n%s", got, want)
	}
}

// TestClean checks R-4: no active task and a clean tree. Done tasks do not count.
func TestClean(t *testing.T) {
	e := newEnv(t)
	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare")
	runGit(t, e.repo, "remote", "add", "origin", remote)
	runGit(t, e.repo, "push", "-q", "-u", "origin", "main")
	e.task("general/old", "Done", "P1", "", "## Plan\n\n- [ ] Step\n", 0)
	check(t, e.run(), "dx: branch main, clean tree, no active tasks.\n")
}

// TestCleanWithUnpushed checks R-4: unpushed commits and nothing else show the header and Unpushed.
func TestCleanWithUnpushed(t *testing.T) {
	e := newEnv(t)
	hash := runGit(t, e.repo, "log", "-1", "--format=%h")
	check(t, e.run(), "dx: branch main, no upstream, 0 changed files\nUnpushed:\n- "+hash+" chore: first\n")
}

// TestStateWithUpstream checks R-1, R-5 to R-8: each section, in order.
func TestStateWithUpstream(t *testing.T) {
	e := newEnv(t)
	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare")
	runGit(t, e.repo, "remote", "add", "origin", remote)
	runGit(t, e.repo, "push", "-q", "-u", "origin", "main")
	runGit(t, e.repo, "commit", "-q", "--allow-empty", "-m", "feat(cli): add resume\n\nRefs: resume-command#R-1\nRefs: resume-command#R-2")
	runGit(t, e.repo, "commit", "-q", "--allow-empty", "-m", "fix: other")
	writeFile(t, filepath.Join(e.repo, "a.go"), "package a\n")
	e.spec("alpha", "Approved")
	e.spec("beta", "Draft")
	e.task("alpha/first", "In Progress", "P1", "", "## Plan\n\n- [x] Done step\n- [ ] Add the hook file\n- [ ] Later\n\n## Notes\n\n- one\n- two\n  - three\n- four\n", time.Hour)
	e.task("alpha/second", "In Progress", "P2", "", "## Plan\n\n- [x] All\n\n## Notes\n\n", 2*time.Hour)
	e.task("alpha/todo", "Todo", "P1", "alpha/blocker", "", 0)
	e.task("alpha/ready", "Todo", "P2", "alpha/finished", "", 0)
	e.task("alpha/finished", "Done", "P2", "", "", 0)
	e.task("alpha/blocker", "Blocked", "P0", "", "## Notes\n\n- first\n- Waits for the API\n", 0)
	e.task("beta/draft", "Todo", "P0", "", "", 0)
	e.task("general/loose", "Todo", "P3", "", "", 0)
	check(t, e.run(), `dx: branch main, 2 ahead of origin/main, 2 changed files
Changed: ?? a.go, ?? docs/
In Progress:
- alpha/first (P1) Title first
  Next step: Add the hook file
  Notes: two
  Notes: three
  Notes: four
- alpha/second (P2) Title second
Next:
- alpha/ready (P2) Title ready
- general/loose (P3) Title loose
Blocked:
- alpha/blocker (P0) Title blocker: Waits for the API
Unpushed:
- `+runGit(t, e.repo, "log", "-1", "--format=%h", "HEAD")+` fix: other
- `+runGit(t, e.repo, "log", "-1", "--format=%h", "HEAD~1")+` feat(cli): add resume [resume-command#R-1, resume-command#R-2]
`)
}

// TestNoUpstream checks R-1, R-8: the commits that no remote branch has.
func TestNoUpstream(t *testing.T) {
	e := newEnv(t)
	e.task("general/work", "In Progress", "P1", "", "", 0)
	hash := runGit(t, e.repo, "log", "-1", "--format=%h")
	check(t, e.run(), "dx: branch main, no upstream, 0 changed files\nIn Progress:\n- general/work (P1) Title work\nUnpushed:\n- "+hash+" chore: first\n")
}

// TestDetached checks R-1: no upstream and no Unpushed section.
func TestDetached(t *testing.T) {
	e := newEnv(t)
	hash := runGit(t, e.repo, "rev-parse", "--short", "HEAD")
	runGit(t, e.repo, "checkout", "-q", "--detach")
	e.task("general/work", "Blocked", "P1", "", "## Notes\n\n- Waits\n", 0)
	check(t, e.run(), "dx: detached at "+hash+"\nBlocked:\n- general/work (P1) Title work: Waits\n")
}

// TestLimits checks R-5, R-7, R-8: the lists end with "+N more".
func TestLimits(t *testing.T) {
	e := newEnv(t)
	for i := range 12 {
		writeFile(t, filepath.Join(e.repo, fmt.Sprintf("f%02d.txt", i)), "x")
	}
	for i := range 5 {
		e.task(fmt.Sprintf("general/t%d", i), "Todo", "P1", "", "", 0)
	}
	for i := range 7 {
		e.task(fmt.Sprintf("general/b%d", i), "Blocked", "P1", "", "", 0)
	}
	for range 12 {
		runGit(t, e.repo, "commit", "-q", "--allow-empty", "-m", "chore: more")
	}
	lines := strings.Split(e.run(), "\n")
	for _, want := range []string{
		"Changed: ?? f00.txt, ?? f01.txt, ?? f02.txt, ?? f03.txt, ?? f04.txt, ?? f05.txt, ?? f06.txt, ?? f07.txt, ?? f08.txt, ?? f09.txt, +2 more",
		"  +2 more", "  +2 more", "  +3 more",
	} {
		found := false
		for _, line := range lines {
			found = found || line == want
		}
		if !found {
			t.Fatalf("no line %q in:\n%s", want, strings.Join(lines, "\n"))
		}
	}
}

// TestCutNotesFirst checks R-9: the output has 40 lines or less, and the Notes lines go first.
func TestCutNotesFirst(t *testing.T) {
	e := newEnv(t)
	for i := range 10 {
		e.task(fmt.Sprintf("general/w%02d", i), "In Progress", "P1", "", "## Plan\n\n- [ ] Step\n\n## Notes\n\n- a\n- b\n- c\n", time.Duration(i)*time.Minute)
	}
	out := e.run()
	if lines := strings.Count(out, "\n"); lines > maxLines || !strings.Contains(out, "Next step: Step") || strings.Contains(out, "Notes: a") || !strings.Contains(out, "Notes: c") {
		t.Fatalf("%d lines:\n%s", lines, out)
	}
	for i := 10; i < 30; i++ {
		e.task(fmt.Sprintf("general/w%02d", i), "In Progress", "P1", "", "", time.Duration(i)*time.Minute)
	}
	out = e.run()
	if lines := strings.Count(out, "\n"); lines != maxLines || !strings.HasSuffix(out, "+5 more\n") {
		t.Fatalf("%d lines:\n%s", lines, out)
	}
}

// TestNoGit checks R-10: with no repository, the header and the task sections only.
func TestNoGit(t *testing.T) {
	dir, cache := t.TempDir(), t.TempDir()
	p, err := project.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	e := &env{t: t, repo: dir, cache: cache, data: p.DataDir(cache)}
	e.task("general/work", "In Progress", "P1", "", "", 0)
	check(t, e.run(), "dx: no git repository\nIn Progress:\n- general/work (P1) Title work\n")
}

// TestErrors checks R-10, R-11: a part that fails shows one line, and nothing is created.
func TestErrors(t *testing.T) {
	e := newEnv(t)
	e.task("general/good", "In Progress", "P1", "", "", 0)
	writeFile(t, filepath.Join(e.data, project.Tasks, "general", "bad.md"), "---\nid: [\n---\n")
	writeFile(t, filepath.Join(e.data, project.Tasks, "general", "wrong.md"), "---\nid: general/other\n---\n")
	out := e.run()
	for _, want := range []string{"dx: skipped general/bad: ", "dx: skipped general/wrong: the frontmatter id is \"general/other\"", "- general/good (P1) Title good"} {
		if !strings.Contains(out, want) {
			t.Fatalf("no %q in:\n%s", want, out)
		}
	}

	if err := os.RemoveAll(e.data); err != nil {
		t.Fatal(err)
	}
	// A missing data directory means no tasks, with no error line.
	check(t, e.run(), "dx: branch main, no upstream, 0 changed files\nUnpushed:\n- "+runGit(t, e.repo, "log", "-1", "--format=%h")+" chore: first\n")
	if _, err := os.Stat(e.data); !os.IsNotExist(err) {
		t.Fatalf("the data directory exists: %v", err)
	}
}

// TestNoCacheDir checks that a missing cache directory skips the tasks.
func TestNoCacheDir(t *testing.T) {
	e := newEnv(t)
	var out bytes.Buffer
	Run(&out, e.repo, "")
	if !strings.Contains(out.String(), "dx: skipped tasks: no cache directory\n") {
		t.Fatalf("output:\n%s", out.String())
	}
}
