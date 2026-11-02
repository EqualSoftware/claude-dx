package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/project"
)

// newTools makes a project in a temporary directory. With git true it is a git repository.
func newTools(t *testing.T, git bool) Tools {
	t.Helper()
	dir := t.TempDir()
	if git {
		if out, err := exec.Command("git", "init", dir).CombinedOutput(); err != nil {
			t.Fatalf("git init: %v\n%s", err, out)
		}
	}
	p, err := project.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return Tools{Project: p}
}

func (t Tools) write(tb testing.TB, rel, text string) {
	tb.Helper()
	path := filepath.Join(t.Project.Root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		tb.Fatal(err)
	}
}

func (t Tools) read(tb testing.TB, rel string) string {
	tb.Helper()
	data, err := os.ReadFile(filepath.Join(t.Project.Root, filepath.FromSlash(rel)))
	if err != nil {
		tb.Fatal(err)
	}
	return string(data)
}

func (t Tools) exists(rel string) bool {
	_, err := os.Lstat(filepath.Join(t.Project.Root, filepath.FromSlash(rel)))
	return err == nil
}

func wantErr(t *testing.T, err error, part string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), part) {
		t.Fatalf("err = %v, want it to contain %q", err, part)
	}
}

func TestOutsideRefused(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "a.txt", "a")
	out := filepath.Join(t.TempDir(), "x")
	for _, call := range map[string]func() error{
		"move source": func() error { _, err := tl.Move(out, "b", false); return err },
		"move dest":   func() error { _, err := tl.Move("a.txt", out, false); return err },
		"move dots":   func() error { _, err := tl.Move("a.txt", "../b", false); return err },
		"copy dest":   func() error { _, err := tl.Copy("a.txt", out, false); return err },
		"delete":      func() error { _, err := tl.Delete([]string{out}, true); return err },
		"mkdir":       func() error { _, err := tl.Mkdir([]string{out}); return err },
		"tree":        func() error { _, err := tl.Tree(context.Background(), out, 1); return err },
	} {
		wantErr(t, call(), "outside the project")
	}
	if !tl.exists("a.txt") {
		t.Fatal("a.txt is gone")
	}
}

func TestMove(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "a.txt", "a")
	tl.write(t, "d/x.txt", "x")
	if _, err := tl.Move("a.txt", "b.txt", false); err != nil {
		t.Fatal(err)
	}
	if tl.exists("a.txt") || tl.read(t, "b.txt") != "a" {
		t.Fatal("file not moved")
	}
	if _, err := tl.Move("d", "e", false); err != nil {
		t.Fatal(err)
	}
	if tl.read(t, "e/x.txt") != "x" || tl.exists("d") {
		t.Fatal("directory not moved")
	}
	_, err := tl.Move("nothing", "z", false)
	wantErr(t, err, "does not exist")
}

func TestOverwrite(t *testing.T) {
	for _, op := range []string{"move", "copy"} {
		tl := newTools(t, false)
		do := func(src, dst string, ow bool) error {
			var err error
			if op == "move" {
				_, err = tl.Move(src, dst, ow)
			} else {
				_, err = tl.Copy(src, dst, ow)
			}
			return err
		}
		tl.write(t, "a.txt", "new")
		tl.write(t, "b.txt", "old")
		tl.write(t, "d/new.txt", "n")
		tl.write(t, "e/old.txt", "o")
		wantErr(t, do("a.txt", "b.txt", false), "exists")
		wantErr(t, do("d", "e", false), "exists")
		if tl.read(t, "b.txt") != "old" || !tl.exists("e/old.txt") {
			t.Fatalf("%s: destination changed", op)
		}
		if err := do("a.txt", "b.txt", true); err != nil || tl.read(t, "b.txt") != "new" {
			t.Fatalf("%s file overwrite: %v", op, err)
		}
		if err := do("d", "e", true); err != nil || !tl.exists("e/new.txt") || tl.exists("e/old.txt") {
			t.Fatalf("%s directory overwrite: %v", op, err)
		}
	}
}

func TestCopyDirectory(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "d/x.txt", "x")
	tl.write(t, "d/s/y.txt", "y")
	if _, err := tl.Copy("d", "c", false); err != nil {
		t.Fatal(err)
	}
	if tl.read(t, "c/x.txt") != "x" || tl.read(t, "c/s/y.txt") != "y" || !tl.exists("d/x.txt") {
		t.Fatal("directory not copied")
	}
	wantErr(t, func() error { _, err := tl.Copy("d", "d/in", false); return err }(), "contain each other")
}

func TestProtected(t *testing.T) {
	tl := newTools(t, true)
	tl.write(t, "a.txt", "a")
	tl.write(t, ".git/note", "n")
	for _, p := range []string{".", ".git", ".git/note", "sub/.."} {
		_, err := tl.Delete([]string{p}, true)
		wantErr(t, err, "cannot be deleted")
		_, err = tl.Move("a.txt", p, true)
		wantErr(t, err, "not allowed")
		_, err = tl.Copy("a.txt", p, true)
		wantErr(t, err, "not allowed")
		_, err = tl.Move(p, "z", true)
		wantErr(t, err, "not allowed")
		_, err = tl.Copy(p, "z", true)
		wantErr(t, err, "not allowed")
	}
	if !tl.exists("a.txt") || !tl.exists(".git/note") {
		t.Fatal("protected path changed")
	}
}

func TestDelete(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "a.txt", "a")
	tl.write(t, "b.txt", "b")
	tl.write(t, "d/x.txt", "x")
	if err := os.Mkdir(filepath.Join(tl.Project.Root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := tl.Delete([]string{"a.txt", "d"}, false)
	wantErr(t, err, "recursive")
	if !tl.exists("a.txt") {
		t.Fatal("a.txt deleted although the call failed")
	}
	if _, err := tl.Delete([]string{"a.txt", "b.txt", "empty"}, false); err != nil {
		t.Fatal(err)
	}
	if tl.exists("a.txt") || tl.exists("b.txt") || tl.exists("empty") {
		t.Fatal("paths not deleted")
	}
	if _, err := tl.Delete([]string{"d"}, true); err != nil || tl.exists("d") {
		t.Fatalf("recursive delete: %v", err)
	}
	_, err = tl.Delete([]string{"d"}, true)
	wantErr(t, err, "does not exist")
}

func TestMkdir(t *testing.T) {
	tl := newTools(t, false)
	if _, err := tl.Mkdir([]string{"a/b/c", "x"}); err != nil {
		t.Fatal(err)
	}
	if !tl.exists("a/b/c") || !tl.exists("x") {
		t.Fatal("directories not made")
	}
	if _, err := tl.Mkdir([]string{"a/b/c"}); err != nil {
		t.Fatalf("existing directory: %v", err)
	}
	tl.write(t, "f", "f")
	if _, err := tl.Mkdir([]string{"f"}); err == nil {
		t.Fatal("mkdir over a file = no error")
	}
}

func TestSymlinkActsOnLink(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "target/x.txt", "x")
	link := filepath.Join(tl.Project.Root, "link")
	if err := os.Symlink(filepath.Join(tl.Project.Root, "target"), link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("cannot make a symbolic link")
		}
		t.Fatal(err)
	}
	checkLinkCopyAndMove(t, tl)
	if _, err := tl.Delete([]string{"link", "link3"}, false); err != nil {
		t.Fatal(err)
	}
	if tl.exists("link") || tl.exists("link3") || tl.read(t, "target/x.txt") != "x" {
		t.Fatal("link not deleted or target changed")
	}
}

func TestTreeGit(t *testing.T) {
	tl := newTools(t, true)
	tl.write(t, ".gitignore", "ignored.txt\n")
	tl.write(t, "ignored.txt", "i")
	tl.write(t, "a.txt", "a")
	tl.write(t, "d/b.txt", "b")
	tl.write(t, "d/e/f/deep.txt", "d")
	got, err := tl.Tree(context.Background(), "", 3)
	if err != nil {
		t.Fatal(err)
	}
	want := ".gitignore\na.txt\nd/\n  b.txt\n  e/\n    f/"
	if got != want {
		t.Fatalf("tree =\n%s\nwant\n%s", got, want)
	}
	got, err = tl.Tree(context.Background(), "d", 1)
	if err != nil || got != "b.txt\ne/" {
		t.Fatalf("tree d depth 1 = %q, %v", got, err)
	}
}

func TestTreeWithoutGit(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "a.txt", "a")
	tl.write(t, "d/b.txt", "b")
	tl.write(t, ".git/config", "c")
	got, err := tl.Tree(context.Background(), "", 0)
	if err != nil || got != "a.txt\nd/\n  b.txt" {
		t.Fatalf("tree = %q, %v", got, err)
	}
}

func TestTreeLimit(t *testing.T) {
	tl := newTools(t, false)
	for i := 0; i < 503; i++ {
		tl.write(t, filepath.Join("d", "f"+string(rune('a'+i/26/26))+string(rune('a'+i/26%26))+string(rune('a'+i%26))), "")
	}
	got, err := tl.Tree(context.Background(), "", 3)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(got, "\n")
	// 504 entries: the directory line and 503 files.
	if len(lines) != 501 || !strings.HasSuffix(lines[500], "lines not shown.") {
		t.Fatalf("lines = %d, last = %q", len(lines), lines[len(lines)-1])
	}
}

func checkLinkCopyAndMove(t *testing.T, tl Tools) {
	t.Helper()
	if _, err := tl.Copy("link", "link2", false); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(filepath.Join(tl.Project.Root, "link2")); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("copy of a link is not a link: %v", err)
	}
	if _, err := tl.Move("link2", "link3", false); err != nil {
		t.Fatal(err)
	}
}

// caseInsensitive reports whether the volume of the project treats names with another case as the same.
func caseInsensitive(tl Tools) bool {
	tl.write(nopTB{}, "CaseProbe", "")
	return tl.exists("caseprobe")
}

// nopTB lets the probe write a file without a test.
type nopTB struct{ testing.TB }

func (nopTB) Helper()      {}
func (nopTB) Fatal(...any) {}

// TestGitCaseVariant checks that .GIT is the protected .git on a case-insensitive volume.
func TestGitCaseVariant(t *testing.T) {
	tl := newTools(t, true)
	if !caseInsensitive(tl) {
		t.Skip("the volume is case sensitive")
	}
	tl.write(t, "a.txt", "a")
	for _, p := range []string{".GIT", ".Git/config", "SUB/../.GIT"} {
		_, err := tl.Delete([]string{p}, true)
		wantErr(t, err, "cannot be deleted")
		_, err = tl.Move("a.txt", p, true)
		wantErr(t, err, "not allowed")
	}
	if !tl.exists(".git/HEAD") {
		t.Fatal(".git changed")
	}
}

func TestCaseOnlyRename(t *testing.T) {
	tl := newTools(t, false)
	if !caseInsensitive(tl) {
		t.Skip("the volume is case sensitive")
	}
	tl.write(t, "Foo.txt", "x")
	if _, err := tl.Move("Foo.txt", "foo.txt", true); err != nil {
		t.Fatal(err)
	}
	names, _ := filepath.Glob(filepath.Join(tl.Project.Root, "*.txt"))
	if len(names) != 1 || filepath.Base(names[0]) != "foo.txt" || tl.read(t, "foo.txt") != "x" {
		t.Fatalf("files = %v", names)
	}
}

func TestSameFileRefused(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "a.txt", "a")
	if err := os.Link(filepath.Join(tl.Project.Root, "a.txt"), filepath.Join(tl.Project.Root, "b.txt")); err != nil {
		t.Skip("cannot make a hard link")
	}
	for _, op := range []func() error{
		func() error { _, err := tl.Move("a.txt", "b.txt", true); return err },
		func() error { _, err := tl.Copy("a.txt", "b.txt", true); return err },
		func() error { _, err := tl.Move("a.txt", "a.txt", true); return err },
	} {
		wantErr(t, op(), "same file")
	}
	if tl.read(t, "a.txt") != "a" || tl.read(t, "b.txt") != "a" {
		t.Fatal("a file is lost")
	}
}

// fifo makes a named pipe, a file type that fs_copy refuses.
func fifo(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("no named pipes")
	}
	if out, err := exec.Command("mkfifo", path).CombinedOutput(); err != nil {
		t.Skipf("mkfifo: %v %s", err, out)
	}
}

// TestCopyFailureKeepsDestination checks that a failed copy keeps the old destination and leaves no temporary name.
func TestCopyFailureKeepsDestination(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "src/ok.txt", "ok")
	fifo(t, filepath.Join(tl.Project.Root, "src", "pipe"))
	tl.write(t, "dst/old.txt", "old")
	_, err := tl.Copy("src", "dst", true)
	wantErr(t, err, "not a file")
	if tl.read(t, "dst/old.txt") != "old" {
		t.Fatal("destination lost")
	}
	_, err = tl.Copy("src/pipe", "p2", false)
	wantErr(t, err, "not a file")
	assertNoTemp(t, tl)
}

func assertNoTemp(t *testing.T, tl Tools) {
	t.Helper()
	entries, err := os.ReadDir(tl.Project.Root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".fs-") {
			t.Fatalf("temporary name %s is left", e.Name())
		}
	}
}

func TestReplaceLeavesNoTemp(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "d/new.txt", "n")
	tl.write(t, "e/old.txt", "o")
	tl.write(t, "f.txt", "f")
	tl.write(t, "g.txt", "g")
	for _, call := range []func() error{
		func() error { _, err := tl.Move("d", "e", true); return err },
		func() error { _, err := tl.Copy("f.txt", "g.txt", true); return err },
		func() error { _, err := tl.Copy("e", "h", true); return err },
	} {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	assertNoTemp(t, tl)
	if tl.read(t, "g.txt") != "f" || !tl.exists("h/new.txt") {
		t.Fatal("content wrong")
	}
}

func TestTreeThroughLinkOutside(t *testing.T) {
	tl := newTools(t, false)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(tl.Project.Root, "link")); err != nil {
		t.Skip("cannot make a symbolic link")
	}
	_, err := tl.Tree(context.Background(), "link", 1)
	wantErr(t, err, "outside the project")
}

func TestDriveRelativeRefused(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("drive letters exist only on Windows")
	}
	tl := newTools(t, false)
	_, err := tl.Delete([]string{"C:foo"}, true)
	wantErr(t, err, "drive")
}

func TestTreeGitLimit(t *testing.T) {
	tl := newTools(t, true)
	for i := 0; i < 503; i++ {
		tl.write(t, fmt.Sprintf("d/f%03d", i), "")
	}
	got, err := tl.Tree(context.Background(), "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, "\n4 lines not shown.") {
		t.Fatalf("tail = %q", got[len(got)-30:])
	}
}

// TestReplaceKeepsBackupWhenRestoreFails checks that the old destination is not deleted when it cannot come back.
func TestReplaceKeepsBackupWhenRestoreFails(t *testing.T) {
	tl := newTools(t, false)
	tl.write(t, "d/new.txt", "n")
	tl.write(t, "e/old.txt", "o")
	calls := 0
	osRename = func(src, dst string) error {
		calls++
		if calls == 1 {
			return os.Rename(src, dst) // dst to the backup name
		}
		return errors.New("rename failed")
	}
	t.Cleanup(func() { osRename = os.Rename })
	_, err := tl.Move("d", "e", true)
	if err == nil || !strings.Contains(err.Error(), "saved in") {
		t.Fatalf("err = %v, want the backup path", err)
	}
	path := err.Error()[strings.Index(err.Error(), "saved in ")+len("saved in "):]
	if data, rerr := os.ReadFile(filepath.Join(path, "old.txt")); rerr != nil || string(data) != "o" {
		t.Fatalf("backup lost: %v", rerr)
	}
}

func TestMkdirProtected(t *testing.T) {
	tl := newTools(t, true)
	paths := []string{".", ".git/x"}
	if caseInsensitive(tl) {
		paths = append(paths, ".GIT/x")
	}
	for _, p := range paths {
		_, err := tl.Mkdir([]string{"ok", p})
		wantErr(t, err, "not allowed")
	}
	if tl.exists("ok") || tl.exists(".git/x") {
		t.Fatal("a directory was made")
	}
}
