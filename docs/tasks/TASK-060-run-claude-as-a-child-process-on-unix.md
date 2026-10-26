---
id: TASK-060
title: Run claude as a child process on Unix
status: Approved
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
- FR-2: The launcher ignores SIGINT and SIGQUIT, because the terminal also sends them to claude, and claude handles them.
- FR-3: On Unix, the launcher sends SIGTERM and SIGHUP that it gets to claude, because a `kill` of the launcher PID must end the session.
- FR-4: The launcher returns the exit code of claude. If a signal ended claude, it returns 128 plus the signal number, as a shell does.

### 6.2 Non-Functional

- NFR-1: The launcher uses no CPU while it waits.

## 7. Design
### 7.1 Architecture

One `start` for all platforms with `exec.Command`, `signal.Ignore` and `signal.Notify`. A small platform file keeps only what differs, if Go needs it.

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
| 1 | `internal/launcher/exec_unix.go`, `internal/launcher/exec_windows.go` (or one `internal/launcher/start.go`), `internal/launcher/launcher_test.go` | FR-1 to FR-4, with tests. | `go test ./internal/launcher/` |
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
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
