package mcpserver

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/EqualSoftware/claude-dx/internal/cache"
)

// writeProjectFile with a file that has the JSON value null, invalid JSON or a value that is not an
// object replaces the file with path and last_used.
func TestWriteProjectFileBadContent(t *testing.T) {
	for _, content := range []string{"null", "{bad", "[1]", "5", `"x"`, ""} {
		dir := t.TempDir()
		path := filepath.Join(dir, "project.json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := writeProjectFile(path, "/root", filepath.Join(dir, "docs.lock")); err != nil {
			t.Fatalf("content %q: %v", content, err)
		}
		var f struct {
			Path     string `json:"path"`
			LastUsed string `json:"last_used"`
		}
		data, _ := os.ReadFile(path)
		if err := json.Unmarshal(data, &f); err != nil || f.Path != "/root" || f.LastUsed == "" {
			t.Fatalf("content %q: project.json = %s, %v", content, data, err)
		}
	}
}

// A session in a linked worktree writes the root of the main worktree, so the repository is not
// missing when the linked worktree is removed.
func TestProjectFileHasMainRoot(t *testing.T) {
	connect(t)
	repo := filepath.Join(os.Getenv("HOME"), "repo") // connect sets HOME to its temporary directory
	other := filepath.Join(os.Getenv("HOME"), "other")
	for _, args := range [][]string{
		{"-C", repo, "-c", "commit.gpgsign=false", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "first"},
		{"-C", repo, "worktree", "add", "-q", other},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	serve(t, other)
	cacheDir, _ := cache.Dir()
	files, _ := filepath.Glob(filepath.Join(cacheDir, "projects", "*", "project.json"))
	if len(files) != 1 {
		t.Fatalf("Project files (cache) = %v, want 1", files)
	}
	var f struct {
		Path string `json:"path"`
	}
	data, _ := os.ReadFile(files[0])
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	root, _ := filepath.EvalSymlinks(repo)
	if f.Path != root {
		t.Fatalf("path = %q, want the main root %q", f.Path, root)
	}
}
