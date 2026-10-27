//go:build !windows

package resume

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeGit puts a script named git first in PATH. The script writes its arguments to a file and runs body.
func fakeGit(t *testing.T, body string) string {
	t.Helper()
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "args")
	script := "#!/bin/sh\necho \"$*\" > '" + log + "'\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return log
}

func TestGitNoOptionalLocks(t *testing.T) {
	log := fakeGit(t, "echo ok")
	if out, err := git(t.TempDir(), "status"); err != nil || out != "ok" {
		t.Fatalf("git = %q, %v", out, err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "--no-optional-locks ") {
		t.Fatalf("arguments = %q, want --no-optional-locks first", data)
	}
}

func TestGitTimeout(t *testing.T) {
	old := gitTimeout
	gitTimeout = 100 * time.Millisecond
	t.Cleanup(func() { gitTimeout = old })
	fakeGit(t, "exec sleep 5")
	start := time.Now()
	if _, err := git(t.TempDir(), "status"); err == nil {
		t.Fatal("the command returned no error after the timeout")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("the command took %v", d)
	}
}
