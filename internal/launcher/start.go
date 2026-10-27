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
// its exit code (ADR-006). So the locks of this process live as long as the session, and no child of claude
// gets them. This process ignores the signals that the terminal also sends to claude, and sends the
// forwarded signals to claude, so a kill of this process ends the session.
func start(claude string, args []string) (int, error) {
	cmd := exec.Command(claude, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// Notify and drop, not signal.Ignore: an ignored signal stays ignored in claude after exec, a caught
	// signal does not. cmd/go does the same for "go run".
	signal.Notify(drop, ignored...)
	defer signal.Stop(drop)
	// Catch the forwarded signals before the start, so a signal that comes during the start is not lost.
	// The channel has space for one signal of each kind, so the package drops a signal only when the same
	// signal is not sent yet, as the OS does with a pending signal.
	fwd, stop := catch(forwarded)
	defer stop()
	if err := cmd.Start(); err != nil {
		return 1, err
	}
	if fwd != nil {
		done := make(chan struct{})
		defer close(done)
		go forward(cmd.Process, fwd, done)
	}
	return exitCode(cmd.Wait())
}

// catch makes the channel for the signals of list and starts to catch them. With an empty list it does
// nothing and returns a nil channel: Notify with no signal would catch all signals (Windows has no
// forwarded signal).
func catch(list []os.Signal) (chan os.Signal, func()) {
	if len(list) == 0 {
		return nil, func() {}
	}
	c := make(chan os.Signal, len(list))
	signal.Notify(c, list...)
	return c, func() { signal.Stop(c) }
}

// forward sends each signal of c to p until done is closed.
func forward(p *os.Process, c <-chan os.Signal, done <-chan struct{}) {
	for {
		select {
		case s := <-c:
			_ = p.Signal(s)
		case <-done:
			return
		}
	}
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
