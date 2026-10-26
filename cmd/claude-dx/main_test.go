package main

import (
	"bufio"
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

	"github.com/EqualSoftware/claude-dx/internal/cache"
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

// environ returns the environment of a child process with PATH set to path and the cache in env.cache.
func (env launchEnv) environ(path string) []string {
	return append(os.Environ(), "PATH="+path, "HOME="+env.cache, "XDG_CACHE_HOME="+env.cache, "LocalAppData="+env.cache)
}

// dxCache returns the claude-dx cache directory of the child processes.
func (env launchEnv) dxCache(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"HOME", "XDG_CACHE_HOME", "LocalAppData"} {
		t.Setenv(name, env.cache)
	}
	dir, err := cache.Dir()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeState writes update.json with latest and the current time.
func (env launchEnv) writeState(t *testing.T, latest string) {
	t.Helper()
	dir := env.dxCache(t)
	state := `{"checked_at":"` + time.Now().UTC().Format(time.RFC3339) + `","latest":"` + latest + `"}`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "update.json"), []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
}

// launch runs claude-dx with PATH set to path and the cache in env.cache.
func (env launchEnv) launch(t *testing.T, path string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(env.dx, args...)
	cmd.Env = env.environ(path)
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
		if len(lines) != 7 || lines[0] != "--plugin-dir" || lines[2] != "--settings" || lines[4] != "--append-system-prompt-file" || lines[6] != args[len(args)-1] {
			t.Fatalf("%q: claude args = %q", args, lines)
		}
		for _, file := range []string{filepath.Join(lines[1], ".claude-plugin", "plugin.json"), lines[5]} {
			if _, err := os.Stat(strings.TrimSpace(file)); err != nil {
				t.Fatalf("%q: plugin is not extracted: %v", args, err)
			}
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
	env.writeState(t, "v9.9.9") // a new update.json: the server does no network request
	cmd := exec.Command(env.dx, "mcp")
	cmd.Env = env.environ(os.Getenv("PATH"))
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
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

// TestCacheCleanRefusesMCPServer is AC-6 of TASK-061: the MCP server holds a shared cache lock
// while it runs.
func TestCacheCleanRefusesMCPServer(t *testing.T) {
	env := setup(t)
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	env.writeState(t, "v9.9.9") // a new update.json: the server does no network request
	cmd := exec.Command(env.dx, "mcp")
	cmd.Env = env.environ(os.Getenv("PATH"))
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	code, _, stderr := env.launch(t, env.bin, "cache", "clean")
	if code != 1 || stderr != "claude-dx: a claude-dx session uses the cache. Close it and try again\n" {
		t.Fatalf("cache clean with a server: exit code = %d, stderr = %q", code, stderr)
	}
	_ = session.Close()
	if code, _, stderr = env.launch(t, env.bin, "cache", "clean"); code != 0 {
		t.Fatalf("cache clean after the server: exit code = %d, stderr = %q", code, stderr)
	}
}

func TestLaunchShowsUpdateNotice(t *testing.T) {
	env := setup(t)
	env.writeState(t, "v10.0.0")
	_, _, stderr := env.launch(t, env.bin, "--resume")
	if stderr != "claude-dx v10.0.0 is available. Run: claude-dx update\n" {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestDoctor(t *testing.T) {
	env := setup(t)
	code, stdout, _ := env.launch(t, env.bin, "doctor")
	if code != 1 || !strings.HasPrefix(stdout, "OK    claude 2.1.308 in PATH\n") || !strings.Contains(stdout, "FAIL  plugin v9.9.9-") {
		t.Fatalf("exit code = %d, stdout:\n%s", code, stdout)
	}
	env.launch(t, env.bin, "--resume") // extracts the plugin
	if code, stdout, _ = env.launch(t, env.bin, "doctor"); code != 0 {
		t.Fatalf("exit code = %d, stdout:\n%s", code, stdout)
	}
}

func TestCacheClean(t *testing.T) {
	env := setup(t)
	env.launch(t, env.bin, "--resume") // extracts the plugin
	dir := env.dxCache(t)
	code, stdout, stderr := env.launch(t, env.bin, "cache", "clean")
	if code != 0 || stdout != "Removed the data in "+dir+".\n" {
		t.Fatalf("exit code = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "plugin")); err == nil {
		t.Fatal("the plugin exists")
	}
	if code, _, stderr = env.launch(t, env.bin, "cache"); code != 1 || stderr != "Usage: claude-dx cache clean\n" {
		t.Fatalf("cache with no subcommand: exit code = %d, stderr = %q", code, stderr)
	}
}

// TestCacheCleanRefusesSession checks AC-2 of TASK-041 with a fake claude that runs until its
// stdin closes.
func TestCacheCleanRefusesSession(t *testing.T) {
	env := setup(t)
	dir := env.dxCache(t)
	cmd := exec.Command(env.dx, "--resume")
	cmd.Env = append(env.environ(env.bin), "FAKECLAUDE_WAIT=1")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("fake claude output = %q, %v", line, err)
	}
	code, _, stderr := env.launch(t, env.bin, "cache", "clean")
	want := "claude-dx: a claude-dx session uses the cache. Close it and try again\n"
	if code != 1 || stderr != want {
		t.Fatalf("exit code = %d, stderr = %q, want %q", code, stderr, want)
	}
	if _, err := os.Stat(filepath.Join(dir, "plugin")); err != nil {
		t.Fatalf("cache clean removed the plugin of a running session: %v", err)
	}
	_ = stdin.Close()
	_ = cmd.Wait()
	if code, _, stderr = env.launch(t, env.bin, "cache", "clean"); code != 0 {
		t.Fatalf("cache clean after the session: exit code = %d, stderr = %q", code, stderr)
	}
}

func TestUpdateRefusesDev(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"update"}, &stdout, &stderr); code != 1 || stderr.String() != "claude-dx: dev builds do not update. Install a release\n" {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
