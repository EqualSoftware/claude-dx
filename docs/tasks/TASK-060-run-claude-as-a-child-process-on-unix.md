---
id: TASK-060
title: Run claude as a child process on Unix
status: Done
priority: P1
milestone: v0.1.0
type: refactor
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli]
depends_on: []
related: [ADR-015, ADR-014, ADR-003]
---

# TASK-060: Run claude as a child process on Unix

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

On all platforms, the launcher starts claude as a child process, waits for it and returns its exit code. So a lock of the launcher lives as long as the session, and no child of claude gets it (ADR-015).

## 2. Problem and Motivation

On Unix the launcher replaces itself with claude (`syscall.Exec`, `internal/launcher/exec_unix.go:13`). A lock of the launcher then needs a descriptor with no close-on-exec, and each child of claude also gets it. A child that runs after claude ends keeps the lock (TASK-050). Windows already runs claude as a child (`internal/launcher/exec_windows.go:15`).

## 3. Goals

- G-1: On all platforms, the launcher process lives until claude ends.
- G-2: The user sees no change: same terminal, same keys, same exit code.

## 4. Out of Scope

- The read-write cache lock: a separate task.
- The session files of ADR-014: they stay until the cache lock task removes them.

## 5. Context
### 5.1 Current State

- `start` in `internal/launcher/exec_unix.go` calls `syscall.Exec`.
- `start` in `internal/launcher/exec_windows.go` ignores `os.Interrupt`, runs claude with `exec.Command` and the same stdio, and returns its exit code.

### 5.2 Assumptions

- A-1: The terminal sends SIGINT, SIGQUIT, SIGTSTP and SIGHUP to all processes of the foreground process group, so claude gets them with no forward. Verify: AC-4.

### 5.3 Constraints

- ADR-015: launcher as a parent process.
- ADR-003: arguments go to claude with no change.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: On Unix, `start` runs claude as a child with the same stdin, stdout, stderr and environment, and waits for it.
- FR-2: The launcher catches SIGINT and SIGQUIT and does nothing with them, because the terminal also sends them to claude, and claude handles them. The launcher does not ignore them (`signal.Ignore`), because an ignored signal stays ignored in claude after exec (Q-1).
- FR-3: On Unix, the launcher sends SIGTERM and SIGHUP that it gets to claude, because a `kill` of the launcher PID must end the session.
- FR-4: The launcher returns the exit code of claude. If a signal ended claude, it returns 128 plus the signal number, as a shell does.

### 6.2 Non-Functional

- NFR-1: The launcher uses no CPU while it waits.

## 7. Design
### 7.1 Architecture

One `start` for all platforms in `internal/launcher/start.go` with `exec.Command` and `signal.Notify`. One channel catches the ignored signals and nobody reads it, so the package drops them. Each forwarded signal has its own channel with space for one signal and one goroutine that sends it to claude. So the package drops a forwarded signal only when the same signal is not sent yet, as the OS does with a pending signal. `signals_unix.go` and `signals_windows.go` give the lists of ignored and forwarded signals. Windows forwards no signal.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

No visible change. `ps` shows `claude-dx` as the parent of `claude`.

### 7.5 Alternatives

- Keep exec and give the lock only to claude: rejected, claude starts its own children.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/launcher/start.go`, `internal/launcher/signals_unix.go`, `internal/launcher/signals_windows.go`, `internal/launcher/start_test.go`, `internal/launcher/start_unix_test.go`; remove `internal/launcher/exec_unix.go` and `internal/launcher/exec_windows.go`; ADR-003 | FR-1 to FR-4, with tests. | `go test -race ./internal/launcher/` |
| 2 | Manual check on macOS | AC-4. | Start `claude-dx`, use Ctrl+C, Ctrl+Z and `fg`, close the terminal. |

## 9. Acceptance Criteria

- AC-1: Given a fake claude that exits with code 3, when the launcher runs it, then the launcher returns 3.
- AC-2 (Unix): Given a fake claude, when the launcher gets SIGTERM, then the fake claude gets SIGTERM and the launcher returns its exit code.
- AC-3 (Unix): Given a fake claude, when the launcher gets SIGINT, then the launcher does not end.
- AC-4 (macOS, manual): Given `claude-dx` in a terminal, then Ctrl+C, Ctrl+Z with `fg`, and the close of the terminal act as with `claude`.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Children of claude that hold a launcher lock | All on Unix | 0 | ADR-015, the cache lock task tests |

- Tests: unit tests with a fake claude for AC-1 to AC-3. AC-4 is a manual check.
- Platforms not verified: Linux terminal behavior (AC-4) is verified only by the unit tests.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A signal does not get to claude | Medium | The terminal sends SIGINT, SIGTSTP and SIGHUP to the foreground process group, which has claude. The launcher sends SIGTERM and SIGHUP to claude. AC-2 to AC-4. |
| One more process for each session | Low | The launcher only waits. |

## 12. Removal and Compatibility

Revert the change. The session lock of ADR-014 works with exec and with a child process.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | FR-2 and 7.1 say signal.Ignore for SIGINT and SIGQUIT. On Unix an ignored signal (SIG_IGN) stays ignored in claude after exec, so claude does not get Ctrl+C. A caught signal goes back to the default action in the child. Use signal.Notify and drop the signals (as cmd/go does for go run)? | Maintainer | Yes. Use signal.Notify and drop the signals. Each forwarded signal has its own channel with space for one signal, so a SIGINT does not make the package drop a SIGTERM. The maintainer approved the work with this recommendation on 2026-10-26. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: Commit 757a705. AC-1: TestStartExitCode. AC-2: TestStartForwardsSIGTERM. AC-3: TestStartIgnoresSIGINT (the fake claude also checks that SIGINT is not ignored, FR-2). FR-4: TestStartSignalExitCode gives 137. Each test fails when its fix is removed. AC-4 (macOS, manual, 2026-10-26, claude-dx of 68953b5): Ctrl+C closes claude; Ctrl+Z and fg work; after the close of the terminal window, pgrep shows no claude-dx or claude process of that session.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: Commit 757a705: go build, go vet (darwin, linux, windows), go test ./... and go test -race -count=5 ./internal/launcher/ pass; golangci-lint v2.15.0 gives 0 issues for darwin, linux and windows; govulncheck finds no vulnerabilities.
- [x] Docs are updated (area docs, `CLAUDE.md`).
  - Evidence: ADR-003 Decision gives the child process and the signal rules (757a705). Learning go-signals-child-process (c973791). No area doc or CLAUDE.md describes the exec.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: Not applicable: one function and two signal lists, with no design diagram.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: ADR-015 records the decision (launcher as parent). ADR-003 Decision changed to match it (757a705).