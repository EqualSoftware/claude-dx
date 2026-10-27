---
topic: go-signals-child-process
description: How Go signal handling of a parent process changes the signals of a child process, and when os/signal drops a signal. Use it when a Go process runs a child and catches or forwards signals.
---

- On Unix, a signal that the parent ignores with `signal.Ignore` (SIG_IGN) stays ignored in the child after exec. A signal that the parent catches with `signal.Notify` goes back to the default action in the child. To make a parent survive Ctrl+C and keep Ctrl+C for the child, catch the signal and drop it. `cmd/go` does the same for `go run`. Proof: cli-launch#R-3 and cli-launch#R-4, the fake claude of `internal/launcher/start_test.go` exits with code 8 when the launcher uses `signal.Ignore`.
- `os/signal` never blocks: when the channel is full, it drops the signal. One channel with space for one signal for many signals can drop SIGTERM after SIGINT. Give each signal that must arrive its own channel with space for one signal; then a drop occurs only when the same signal is not read yet, as the OS does with a pending signal.
