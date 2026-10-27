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
	"slices"
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
	env.writeState(t, "v9.9.9") // a new update.json: the server does no network request
	const runs = 3
	starts := make([]time.Duration, 0, runs)
	costs := make([]time.Duration, 0, runs)
	for range runs {
		costs = append(costs, processCost(t, env))
		starts = append(starts, startMCPServer(t, env))
	}
	// mcp-server#R-9: the server starts in less than 100 ms, minus the cost of a process start.
	// The fastest runs filter out machine noise.
	diff := slices.Min(starts) - slices.Min(costs)
	t.Logf("initialize took %v; process cost (--version) took %v; difference = %v", starts, costs, diff)
	if diff >= 100*time.Millisecond {
		t.Fatalf("start minus process cost = %v, want less than 100ms; starts: %v; costs: %v", diff, starts, costs)
	}
}

// processCost returns the time that claude-dx --version needs, from start to exit.
func processCost(t *testing.T, env launchEnv) time.Duration {
	t.Helper()
	cmd := exec.Command(env.dx, "--version")
	cmd.Env = env.environ(os.Getenv("PATH"))
	start := time.Now()
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return time.Since(start)
}

// startMCPServer starts claude-dx mcp, checks the server info and returns the time of the initialize.
func startMCPServer(t *testing.T, env launchEnv) time.Duration {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.Command(env.dx, "mcp")
	cmd.Env = env.environ(os.Getenv("PATH"))
	start := time.Now()
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	info := session.InitializeResult().ServerInfo
	if info.Name != "dx" || info.Version != "v9.9.9" {
		t.Fatalf("server info = %+v, want dx v9.9.9", info)
	}
	return elapsed
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

// TestCleanListMCPServer checks mcp-server#R-10: while an MCP server runs, the global item is disabled.
func TestCleanListMCPServer(t *testing.T) {
	env := setup(t)
	dir := env.dxCache(t)
	if err := os.MkdirAll(filepath.Join(dir, "plugin", "old"), 0o755); err != nil {
		t.Fatal(err)
	}
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
	code, stdout, stderr := env.launch(t, env.bin, "clean", "--list")
	if code != 0 || !strings.Contains(stdout, "Old plugin versions") || !strings.Contains(stdout, "(in use by a running session)") {
		t.Fatalf("clean --list with a server: exit code = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}
	_ = session.Close()
	if _, stdout, _ = env.launch(t, env.bin, "clean", "--list"); strings.Contains(stdout, "in use") {
		t.Fatalf("clean --list after the server: stdout = %q", stdout)
	}
}

// TestCleanListRefusesSession checks session#R-1 with a fake claude that runs until its stdin closes.
func TestCleanListRefusesSession(t *testing.T) {
	env := setup(t)
	dir := env.dxCache(t)
	if err := os.MkdirAll(filepath.Join(dir, "plugin", "old"), 0o755); err != nil {
		t.Fatal(err)
	}
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
	// A session removes the old versions at its start, so the old version comes after it.
	if err := os.MkdirAll(filepath.Join(dir, "plugin", "old2"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, out, _ := env.launch(t, env.bin, "clean", "--list")
	if code != 0 || !strings.Contains(out, "(in use by a running session)") {
		t.Fatalf("exit code = %d, stdout = %q", code, out)
	}
	_ = stdin.Close()
	_ = cmd.Wait()
	if _, out, _ = env.launch(t, env.bin, "clean", "--list"); strings.Contains(out, "in use") {
		t.Fatalf("clean --list after the session: stdout = %q", out)
	}
}

// TestCleanNeedsTerminal checks cli-clean#R-7: the test process has no terminal on stdin.
func TestCleanNeedsTerminal(t *testing.T) {
	env := setup(t)
	code, stdout, stderr := env.launch(t, env.bin, "clean")
	want := "claude-dx clean needs an interactive terminal. Run claude-dx clean --list to see the data.\n"
	if code != 1 || stdout != "" || stderr != want {
		t.Fatalf("exit code = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}
}

// TestHelpHasNoCacheClean checks cli-clean#R-8. The parser test checks that "cache" starts claude.
func TestHelpHasNoCacheClean(t *testing.T) {
	if strings.Contains(help, "cache clean") || !strings.Contains(help, "claude-dx clean --list") {
		t.Fatalf("help:\n%s", help)
	}
}

// TestHelpHasResume checks resume-command#R-12.
func TestHelpHasResume(t *testing.T) {
	want := "claude-dx resume                 Print the work state (tasks and git) for a new session.\n"
	if !strings.Contains(help, want) {
		t.Fatalf("help:\n%s", help)
	}
}

// TestResumeExitsZero checks resume-command#R-10, R-11: the exit code is 0 with no repository and no
// data directory (no tasks, no error line), and the command creates nothing.
func TestResumeExitsZero(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"HOME", "XDG_CACHE_HOME", "LocalAppData"} {
		t.Setenv(name, home)
	}
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := run([]string{"resume"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr: %s", code, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "dx: no git repository\n") {
		t.Fatalf("stdout:\n%s", stdout.String())
	}
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Fatalf("resume created %d entries in the cache", len(entries))
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
