package clean

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EqualSoftware/claude-dx/internal/filelock"
	"github.com/EqualSoftware/claude-dx/internal/project"
)

// write makes the file path with data.
func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeInfo writes the Project file (cache) of the folder key.
func writeInfo(t *testing.T, cacheDir, key, path, lastUsed string) {
	t.Helper()
	data, _ := json.Marshal(map[string]string{"path": path, "last_used": lastUsed})
	write(t, filepath.Join(cacheDir, "projects", key, "project.json"), string(data))
}

// writeTask writes a task file with a status.
func writeTask(t *testing.T, dataDir, spec, slug, status string) string {
	t.Helper()
	path := filepath.Join(dataDir, "tasks", spec, slug+".md")
	write(t, path, fmt.Sprintf("---\nid: %s/%s\ntitle: T\nstatus: %s\n---\nbody\n", spec, slug, status))
	return path
}

// fixture is a cache with old plugin versions and projects.
type fixture struct {
	cache   string
	root    string // the root of a project that exists
	missing string // the root of a project that does not exist
	data    string // the data directory of the project that exists
	tasks   []string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	f := fixture{cache: filepath.Join(t.TempDir(), "claude-dx")}
	write(t, filepath.Join(f.cache, "plugin", "own", "a.txt"), "own")
	write(t, filepath.Join(f.cache, "plugin", "old1", "a.txt"), "old")
	write(t, filepath.Join(f.cache, "plugin", "old2.tmp-1", "a.txt"), "old")
	f.missing = filepath.Join(t.TempDir(), "gone")
	p, err := project.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f.root = p.Root
	f.data = p.DataDir(f.cache)
	writeInfo(t, f.cache, p.Key, p.Root, "2026-01-02T00:00:00Z")
	f.tasks = []string{
		writeTask(t, f.data, "alpha", "one", "Todo"),
		writeTask(t, f.data, "alpha", "two", "Todo"),
		writeTask(t, f.data, "alpha", "three", "Done"),
		writeTask(t, f.data, "beta", "four", "In Progress"),
	}
	write(t, filepath.Join(f.data, "docs.lock"), "")
	write(t, filepath.Join(f.cache, "cache.lock"), "")
	writeInfo(t, f.cache, "aaaaaaaaaaaaaaaa", f.missing, "2026-01-03T00:00:00Z")
	write(t, filepath.Join(f.cache, "projects", "aaaaaaaaaaaaaaaa", "other.txt"), "x")
	write(t, filepath.Join(f.cache, "projects", "aaaaaaaaaaaaaaaa", "docs.lock"), "")
	write(t, filepath.Join(f.cache, "projects", "bbbbbbbbbbbbbbbb", "docs.lock"), "") // only a lock file
	return f
}

// labels returns "group | label" of each item.
func labels(groups []Group) []string {
	var out []string
	for _, g := range groups {
		for _, it := range g.Items {
			out = append(out, g.Name+" | "+it.Label)
		}
	}
	return out
}

func scan(t *testing.T, f fixture) []Group {
	t.Helper()
	groups, err := Scan(f.cache, "own")
	if err != nil {
		t.Fatal(err)
	}
	return groups
}

// all selects every item.
func all(groups []Group) ([]Item, error) {
	var out []Item
	for _, g := range groups {
		out = append(out, g.Items...)
	}
	return out, nil
}

func ui(sel func([]Group) ([]Item, error), answer bool, prompts *[]string) UI {
	return UI{
		Select: sel,
		Confirm: func(p string) (bool, error) {
			*prompts = append(*prompts, p)
			return answer, nil
		},
	}
}

func TestScanTree(t *testing.T) {
	f := newFixture(t)
	groups := scan(t, f)
	want := []string{
		"Global | Old plugin versions",
		f.missing + " (missing) | All data",
		f.root + " | alpha Todo: 2 tasks",
		f.root + " | alpha Done: 1 task",
		f.root + " | beta In Progress: 1 task",
	}
	if got := labels(groups); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("items:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if groups[0].Items[0].Size != 6 {
		t.Fatalf("size of old versions = %d, want 6", groups[0].Items[0].Size)
	}
	if groups[2].Items[0].Size == 0 {
		t.Fatal("a task group has no size")
	}
}

// session#R-4: a missing cache gives an empty tree.
func TestScanMissingCache(t *testing.T) {
	groups, err := Scan(filepath.Join(t.TempDir(), "claude-dx"), "own")
	if err != nil || len(groups) != 0 {
		t.Fatalf("Scan = %v, %v", groups, err)
	}
}

func TestNoSelection(t *testing.T) {
	f := newFixture(t)
	var out bytes.Buffer
	var prompts []string
	none := func([]Group) ([]Item, error) { return nil, nil }
	if err := Run(&out, f.cache, "own", ui(none, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Nothing deleted.\n" || len(prompts) != 0 || !exists(f.tasks[0]) {
		t.Fatalf("out = %q, prompts = %q", out.String(), prompts)
	}
}

func TestAnswerNo(t *testing.T) {
	f := newFixture(t)
	var out bytes.Buffer
	var prompts []string
	if err := Run(&out, f.cache, "own", ui(all, false, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), "Nothing deleted.\n") || !exists(f.tasks[0]) || !exists(filepath.Join(f.cache, "plugin", "old1")) {
		t.Fatalf("out = %q", out.String())
	}
	if len(prompts) != 1 || prompts[0] != "Delete 5 items ("+FormatSize(sum(scan(t, f)))+")?" {
		t.Fatalf("prompts = %q", prompts)
	}
	if !strings.Contains(out.String(), "Global: Old plugin versions (6 B)\n") {
		t.Fatalf("the selected items have no sizes: %q", out.String())
	}
}

// session#R-4 and project-data#R-5: Yes deletes the items, and the lock files stay.
func TestAnswerYes(t *testing.T) {
	f := newFixture(t)
	var out bytes.Buffer
	var prompts []string
	want := "Deleted 5 items (" + FormatSize(sum(scan(t, f))) + ").\n"
	if err := Run(&out, f.cache, "own", ui(all, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(out.String(), want) {
		t.Fatalf("out = %q", out.String())
	}
	gone := []string{"plugin/old1", "plugin/old2.tmp-1", "projects/aaaaaaaaaaaaaaaa/project.json", "projects/aaaaaaaaaaaaaaaa/other.txt"}
	for _, g := range gone {
		if exists(filepath.Join(f.cache, filepath.FromSlash(g))) {
			t.Errorf("%s exists", g)
		}
	}
	for _, task := range f.tasks {
		if exists(task) {
			t.Errorf("%s exists", task)
		}
	}
	keep := []string{
		filepath.Join(f.cache, "plugin", "own", "a.txt"),
		filepath.Join(f.cache, "cache.lock"),
		filepath.Join(f.cache, "projects", "aaaaaaaaaaaaaaaa", "docs.lock"),
		filepath.Join(f.cache, "projects", "bbbbbbbbbbbbbbbb", "docs.lock"),
		filepath.Join(f.data, "docs.lock"),
	}
	for _, path := range keep {
		if !exists(path) {
			t.Errorf("%s was deleted", path)
		}
	}
	if groups := scan(t, f); len(groups) != 0 {
		t.Fatalf("items after the delete: %v", labels(groups))
	}
}

// session#R-4: no lock file is an item, and the size has no lock file.
func TestNoLockFiles(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.cache, "plugin", "old1", "docs.lock"), "123456789")
	groups := scan(t, f)
	for _, g := range groups {
		for _, it := range g.Items {
			if strings.Contains(it.Text(), ".lock") {
				t.Errorf("item %q shows a lock file", it.Text())
			}
		}
	}
	if got := groups[0].Items[0].Size; got != 6 {
		t.Fatalf("size = %d, want 6 with no lock file", got)
	}
}

// session#R-4: the shared cache lock disables the global item. The other items stay enabled.
func TestGlobalItemDisabled(t *testing.T) {
	f := newFixture(t)
	release, err := filelock.RLock(filepath.Join(f.cache, "cache.lock"))
	if err != nil {
		t.Fatal(err)
	}
	groups := scan(t, f)
	if it := groups[0].Items[0]; it.Reason != "(in use by a running session)" {
		t.Fatalf("reason = %q", it.Reason)
	}
	if groups[1].Items[0].Reason != "" {
		t.Fatal("a project item is disabled")
	}
	var out bytes.Buffer
	List(&out, groups)
	if !strings.Contains(out.String(), "Old plugin versions  6 B  (in use by a running session)\n") {
		t.Fatalf("list:\n%s", out.String())
	}
	// A disabled item that a UI returns is not deleted.
	var prompts []string
	if err := Run(&out, f.cache, "own", ui(all, true, &prompts)); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(f.cache, "plugin", "old1")) || !strings.HasPrefix(prompts[0], "Delete 4 items") {
		t.Fatalf("a disabled item was deleted, prompts = %q", prompts)
	}
	_ = release()
	if scan(t, f)[0].Items[0].Reason != "" {
		t.Fatal("the item is disabled after the release")
	}
}

// A lock that comes after the scan makes the delete fail, and the files stay.
func TestGlobalDeleteLockedAfterScan(t *testing.T) {
	f := newFixture(t)
	groups := scan(t, f)
	release, err := filelock.RLock(filepath.Join(f.cache, "cache.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = release() }()
	var out bytes.Buffer
	var prompts []string
	pick := func([]Group) ([]Item, error) { return groups[0].Items, nil }
	err = Run(&out, f.cache, "own", ui(pick, true, &prompts))
	if err == nil || !strings.Contains(err.Error(), "in use by a running session") || !exists(filepath.Join(f.cache, "plugin", "old1")) {
		t.Fatalf("err = %v", err)
	}
}

func TestListDeletesNothing(t *testing.T) {
	f := newFixture(t)
	var out bytes.Buffer
	List(&out, scan(t, f))
	for _, want := range []string{"Global\n  Old plugin versions  6 B\n", f.missing + " (missing)\n  All data  ", "  alpha Todo: 2 tasks  "} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("list has no %q:\n%s", want, out.String())
		}
	}
	if !exists(f.tasks[0]) || !exists(filepath.Join(f.cache, "plugin", "old1")) {
		t.Fatal("List deleted data")
	}
}

func TestFormatSize(t *testing.T) {
	tests := map[int64]string{0: "0 B", 512: "512 B", 1023: "1023 B", 1024: "1.0 KB", 12595: "12.3 KB", 4404019: "4.2 MB", 3 << 29: "1.5 GB"}
	for n, want := range tests {
		if got := FormatSize(n); got != want {
			t.Errorf("FormatSize(%d) = %q, want %q", n, got, want)
		}
	}
}

// sum returns the size of all items.
func sum(groups []Group) int64 {
	var n int64
	for _, g := range groups {
		for _, it := range g.Items {
			n += it.Size
		}
	}
	return n
}

// TestHoldDocsLock is a helper process: it holds the exclusive lock in DX_DOCS_LOCK until stdin closes.
func TestHoldDocsLock(t *testing.T) {
	path := os.Getenv("DX_DOCS_LOCK")
	if path == "" {
		t.Skip("helper process")
	}
	f, err := filelock.TryLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	fmt.Println("ready")
	_, _ = io.ReadAll(os.Stdin)
}

// project-data#R-5: a task group delete waits while another process holds the docs lock.
func TestTaskDeleteWaitsForDocsLock(t *testing.T) {
	f := newFixture(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestHoldDocsLock$")
	cmd.Env = append(os.Environ(), "DX_DOCS_LOCK="+filepath.Join(f.data, "docs.lock"))
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	release := func() { _ = stdin.Close(); _ = cmd.Wait() }
	t.Cleanup(release)
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper output = %q, %v", line, err)
	}
	pick := func(groups []Group) ([]Item, error) { return groups[2].Items[:1], nil }
	done := make(chan error, 1)
	var out bytes.Buffer
	var prompts []string
	go func() { done <- Run(&out, f.cache, "own", ui(pick, true, &prompts)) }()
	select {
	case err := <-done:
		t.Fatalf("Run did not wait for the docs lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	if !exists(f.tasks[0]) {
		t.Fatal("a task was deleted while the lock was held")
	}
	release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if exists(f.tasks[0]) {
		t.Fatal("the task exists after the lock was released")
	}
}
