---
id: TASK-050
title: Free session files of ended sessions
status: In Review
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli, plugin]
depends_on: []
related: [TASK-041, ADR-014]
---

# TASK-050: Free session files of ended sessions

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |

## 1. Summary

A session file that a process of an ended session still holds does not stop `claude-dx` or `cache clean`. The launcher replaces a session file with its own PID, and on Unix a session file with a PID that does not run counts as ended.

## 2. Problem and Motivation

On Unix the session descriptor has no close-on-exec (`internal/session/inherit_unix.go`), so each child of claude also gets it: the MCP server, hooks and Bash commands. A child that runs after claude ends (example: a server started with `nohup ... &`) holds the lock of `<cache>/sessions/<pid>.lock`. Then:
- `cache clean` refuses each time with the PID of a session that ended (`internal/session/session.go:62`), and `Close it and try again` does not help.
- If a new launcher gets the same PID, `hold` (`internal/session/session.go:79`) fails with `the file is locked`, and `claude-dx` does not start.

## 3. Goals

- G-1: `claude-dx` starts when an old session file with its PID is locked.
- G-2: On Unix, `cache clean` and the plugin cleanup treat a session file of a PID that does not run as a session that ended.

## 4. Out of Scope

- Close of the descriptor in the children of claude: claude-dx does not control how claude starts processes.
- Windows: Go starts child processes with a list of the handles to inherit, so no child gets the session handle.

## 5. Context
### 5.1 Current State

- `running` (`internal/session/session.go:97`) counts each session file with a held lock as a running session.
- `hold` writes the file, then takes the lock with `TryLock`.

### 5.2 Assumptions

- A-1: On Unix, a remove of a file that another process locks with `flock` passes, and a new file with that name has a free lock. Verify: a test.
- A-2: `kill(pid, 0)` returns `ESRCH` when no process has the PID. Verify: a test with the PID of a process that ended.

### 5.3 Constraints

- ADR-014: session lock files.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `hold` removes the session file of its own PID before it writes it. A file with the PID of this process is always of an ended session, because two running processes do not have one PID.
- FR-2: On Unix, `running` treats a held session file as ended when `kill(pid, 0)` returns `ESRCH`: it removes the file and does not count it. Another error, example `EPERM`, counts as running, because a wrong removal is worse than a stale file. On Windows, a held file always counts as running.

### 6.2 Non-Functional

- NFR-1: One system call for each held session file, with no new process.

## 7. Design
### 7.1 Architecture

A function `alive(pid int) bool`: `inherit_unix.go` uses `unix.Kill(pid, 0)`, `inherit_windows.go` returns true. `running` calls it after a failed `TryLock`. `hold` calls `os.Remove` first.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

`cache clean` does not refuse because of a session that ended.

### 7.5 Alternatives

- Clear the lock of the children: not possible, claude starts them.
- Use only `kill(pid, 0)` with no lock: rejected, a reused PID then keeps a stale file, and the lock is the source of truth for a running session.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/session/session.go`, `internal/session/inherit_unix.go`, `internal/session/inherit_windows.go`, `internal/session/session_test.go`, `internal/session/session_unix_test.go` | FR-1, FR-2, with tests. | `go test ./internal/session/` |

## 9. Acceptance Criteria

- AC-1 (Unix): Given a session file with the PID of this process that another process locks, when `Start` runs, then it passes. Windows is not in AC-1: Windows cannot remove a file that another process has open, and no child gets the session handle there.
- AC-2 (Unix): Given a session file that another process locks and a PID that does not run, when `Clean` runs, then it removes the cache.
- AC-3: Given a session file that another process locks and the PID of a running process, when `Clean` runs, then it refuses (TASK-041 AC-2 stays true).

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| `cache clean` refusals for an ended session | Each time | 0 | AC-2 test |

- Tests: unit tests with helper processes for the lock.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A new process gets the PID of an ended session, so its file counts as running | Low | The file stays until that process ends. This is the same as now. |

## 12. Removal and Compatibility

Remove `alive` and the `os.Remove` in `hold`.

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
