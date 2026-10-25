//go:build windows

package launcher

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

// start runs claude as a child process with the same console, and returns its exit code.
// Windows has no exec. Ctrl+C goes to all processes of the console, so claude gets it,
// and this process ignores it.
func start(claude string, args []string) (int, error) {
	signal.Ignore(os.Interrupt)
	cmd := exec.Command(claude, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	return exitCode(err)
}

// exitCode returns the exit code of a finished command error.
func exitCode(err error) (int, error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 1, err
}
