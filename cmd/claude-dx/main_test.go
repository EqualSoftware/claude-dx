package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if got, want := stdout.String(), "claude-dx dev\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunHelpDoesNotStartClaude(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "claude-dx -- --help") {
		t.Fatalf("help does not show the -- form:\n%s", stdout.String())
	}
}

// launchEnv holds the binaries and the environment of one launcher test.
type launchEnv struct {
	dx    string // path of claude-dx
	bin   string // directory with the fake claude
	cache string // cache base directory
}

// setup builds claude-dx and the fake claude into temp directories.
func setup(t *testing.T) launchEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("builds binaries")
	}
	root := t.TempDir()
	env := launchEnv{
		dx:    filepath.Join(root, "dx", exe("claude-dx")),
		bin:   filepath.Join(root, "bin"),
		cache: filepath.Join(root, "cache"),
	}
	build(t, env.dx, "-ldflags=-X main.version=v9.9.9", ".")
	build(t, filepath.Join(env.bin, exe("claude")), "./testdata/fakeclaude")
	return env
}

func build(t *testing.T, out string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"build", "-o", out}, args...)...)
	if msg, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %v: %v\n%s", args, err, msg)
	}
}

// launch runs claude-dx with PATH set to path and the cache in env.cache.
func (env launchEnv) launch(t *testing.T, path string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(env.dx, args...)
	cmd.Env = append(os.Environ(), "PATH="+path,
		"HOME="+env.cache, "XDG_CACHE_HOME="+env.cache, "LocalAppData="+env.cache)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatal(err)
	}
	return cmd.ProcessState.ExitCode(), stdout.String(), stderr.String()
}

func TestLaunchPassesArgsAndExitCode(t *testing.T) {
	env := setup(t)
	tests := [][]string{{"--resume"}, {"--", "--help"}}
	for _, args := range tests {
		code, stdout, stderr := env.launch(t, env.bin, args...)
		if code != 3 {
			t.Fatalf("%q: exit code = %d, want 3; stderr: %s", args, code, stderr)
		}
		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		if len(lines) != 5 || lines[0] != "--plugin-dir" || lines[2] != "--settings" || lines[4] != args[len(args)-1] {
			t.Fatalf("%q: claude args = %q", args, lines)
		}
		manifest := filepath.Join(strings.TrimSpace(lines[1]), ".claude-plugin", "plugin.json")
		if _, err := os.Stat(manifest); err != nil {
			t.Fatalf("%q: plugin is not extracted: %v", args, err)
		}
	}
}

func TestLaunchWithoutClaude(t *testing.T) {
	env := setup(t)
	code, _, stderr := env.launch(t, t.TempDir(), "--resume")
	if code != 1 || !strings.Contains(stderr, "claude is not in PATH") {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr)
	}
}

func TestLaunchDeletesOldBinary(t *testing.T) {
	env := setup(t)
	old := env.dx + ".old"
	if err := os.WriteFile(old, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	env.launch(t, env.bin, "--resume")
	if _, err := os.Stat(old); err == nil {
		t.Fatal(".old file was not deleted")
	}
}

func TestMCPServerInitialize(t *testing.T) {
	env := setup(t)
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.Command(env.dx, "mcp")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	t.Logf("initialize took %v", time.Since(start))
	info := session.InitializeResult().ServerInfo
	if info.Name != "dx" || info.Version != "v9.9.9" {
		t.Fatalf("server info = %+v, want dx v9.9.9", info)
	}
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
