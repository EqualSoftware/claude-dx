---
id: TASK-051
title: Keep the cache lock during cache clean
status: Draft
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli]
depends_on: []
related: [ADR-014, TASK-041, TASK-050]
---

# TASK-051: Keep the cache lock during cache clean

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

A session start that waits for `cache clean` starts correctly, and the cache lock stops two session starts at the same time also after a `cache clean`.

## 2. Problem and Motivation

Found by the code review of 2026-10-26.
- `Start` (`internal/session/session.go:28`) makes `<cache>/sessions` before it takes the cache lock. `Clean` removes that directory. A start that waits for the lock during `cache clean` then fails in `hold` with `no such file or directory`, and `claude-dx` does not start.
- `Clean` removes `<cache>/sessions.lock` after it releases the lock (`internal/session/session.go:69`). On Unix, a start that waits holds the lock of the removed file. A second start makes a new `sessions.lock` and locks it. Then the two starts run `ensure` and `removeUnused` at the same time, and one start can remove the `*.tmp-*` extraction of the other.

## 3. Goals

- G-1: A session start that waits for `cache clean` starts.
- G-2: Only one session start or `cache clean` runs at a time, also after a `cache clean`.

## 4. Out of Scope

- The session lock of each session (TASK-050).
- The numbering lock and the docs lock (ADR-007).

## 5. Context
### 5.1 Current State

- `Start` takes the cache lock with `filelock.Lock` after `os.MkdirAll` of `sessions`.
- `Clean` removes all files but the lock file with `cache.Clean`, releases the lock, then removes the lock file and the cache directory, because Windows cannot remove an open file.

### 5.2 Assumptions

- A-1: On Unix, a process that waits in `flock` on a file that another process removes gets the lock of the removed file. Verify: a test.

### 5.3 Constraints

- ADR-014: session lock files.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `Start` takes the cache lock before it makes `<cache>/sessions`, because `Clean` can remove that directory while `Start` waits.
- FR-2: `Clean` does not remove `<cache>/sessions.lock` and the cache directory. The cache then has one empty file after `cache clean`. Because the lock file stays, all processes lock the same file.

### 6.2 Non-Functional

- NFR-1: No new system call in `Start`.

## 7. Design
### 7.1 Architecture

`Start` calls `filelock.Lock` first, then `os.MkdirAll`. `Clean` stops after `cache.Clean` and the unlock. The loop that removes the lock file and the cache directory goes away.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

`claude-dx cache clean` leaves `<cache>/sessions.lock`. Its message does not change.

### 7.5 Alternatives

- After `Lock`, compare the locked file with the file at the path (`os.SameFile`) and lock again if they are different: rejected, more code for one empty file.
- Put the lock file outside the cache: rejected, the cache is the only directory of claude-dx.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/session/session.go`, `internal/session/session_test.go`, `internal/cli/` (help text of `cache clean`, if it names the lock file) | FR-1, FR-2, with tests. | `go test ./internal/session/ ./internal/cli/` |

## 9. Acceptance Criteria

- AC-1: Given `Clean` that removed the cache, when `Start` runs, then it passes and the session file exists.
- AC-2: Given `Clean` that ran, then `<cache>/sessions.lock` exists and the cache has no other file.
- AC-3: Given a running session, when `Clean` runs, then it refuses (TASK-041 AC-2 stays true).

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Failed starts after `cache clean` | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A user thinks that `cache clean` did not remove all files | Low | The file is empty and the next start uses it. |

## 12. Removal and Compatibility

Revert the change.

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
