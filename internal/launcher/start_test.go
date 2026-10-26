package launcher

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// TestHelperClaude is the fake claude of the start tests. It runs only in a child process with DX_FAKE_CLAUDE.
// "exit3" exits with code 3. "wait" exits with code 8 if it got SIGINT as ignored, else makes the file
// DX_READY, then exits with code 7 on SIGTERM.
// "kill" ends itself with SIGKILL.
func TestHelperClaude(_ *testing.T) {
	switch os.Getenv("DX_FAKE_CLAUDE") {
	case "exit3":
		os.Exit(3)
	case "wait":
		if signal.Ignored(os.Interrupt) {
			os.Exit(8)
		}
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		if err := os.WriteFile(os.Getenv("DX_READY"), nil, 0o644); err != nil {
			os.Exit(1)
		}
		<-sigs
		os.Exit(7)
	case "kill":
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Kill()
		time.Sleep(10 * time.Second)
	}
}

// fakeClaude runs the test binary as claude in mode, and returns the exit code of start.
func fakeClaude(t *testing.T, mode string) (int, error) {
	t.Helper()
	t.Setenv("DX_FAKE_CLAUDE", mode)
	return start(os.Args[0], []string{"-test.run=^TestHelperClaude$"})
}

// TestStartExitCode is AC-1 of TASK-060.
func TestStartExitCode(t *testing.T) {
	if code, err := fakeClaude(t, "exit3"); err != nil || code != 3 {
		t.Fatalf("start = %d, %v; want 3", code, err)
	}
}
