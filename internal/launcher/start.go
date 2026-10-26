package launcher

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// drop gets the ignored signals while start runs. Nobody reads it, so the package drops them. A test reads
// it to know that start got a signal.
var drop = make(chan os.Signal, 1)

// start runs claude as a child process with the same terminal and environment, waits for it and returns
// its exit code (ADR-015). So the locks of this process live as long as the session, and no child of claude
// gets them. This process ignores the signals that the terminal also sends to claude, and sends the
// forwarded signals to claude, so a kill of this process ends the session.
func start(claude string, args []string) (int, error) {
	cmd := exec.Command(claude, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// Notify and drop, not signal.Ignore: an ignored signal stays ignored in claude after exec, a caught
	// signal does not. cmd/go does the same for "go run".
	signal.Notify(drop, ignored...)
	defer signal.Stop(drop)
	if err := cmd.Start(); err != nil {
		return 1, err
	}
	for _, s := range forwarded {
		defer forward(cmd.Process, s)()
	}
	return exitCode(cmd.Wait())
}

// forward sends each s that this process gets to p, and returns the function that stops it.
// Each signal has its own channel with space for one signal, so the package drops a signal only when
// the same signal is not sent yet, as the OS does with a pending signal.
func forward(p *os.Process, s os.Signal) func() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, s)
	go func() {
		for range c {
			_ = p.Signal(s)
		}
	}()
	// Stop sends no more signals when it returns, so the close is safe.
	return func() { signal.Stop(c); close(c) }
}

// exitCode returns the exit code of a finished command error. A command that a signal ended gives
// 128 plus the signal number, as a shell does.
func exitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return 1, err
	}
	if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal()), nil
	}
	return exitErr.ExitCode(), nil
}
