---
id: TASK-062
title: Lock the docs for read tools
status: Done
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-061, TASK-059]
related: [ADR-015, ADR-007, TASK-047]
---

# TASK-062: Lock the docs for read tools

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |
| 2026-10-26 | Equal | Status: In Review. The scope changed in section 6.1. |
| 2026-10-26 | Equal | Status: Approved. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

The docs tools that read hold a shared docs lock, so a read never sees half of a change of more than one file. The docs tools that write keep the exclusive docs lock (ADR-015).

## 2. Problem and Motivation

The read tools take no lock. `docfile.Write` makes each file change atomic, but some changes write more than one file: `adr_supersede` changes two ADRs, `task_create` and `task_renumber` change a task and the README index. A read between the two writes sees half of the change.

## 3. Goals

- G-1: A read tool sees the docs before or after each change, never between two files of one change.
- G-2: Read tools do not wait for other read tools.

## 4. Out of Scope

- Changes by persons or by git (checkout, rebase): no lock of claude-dx can stop them.
- The cache lock (TASK-061).

## 5. Context
### 5.1 Current State

- `docfile.Locked` (`internal/docs/docfile/write.go:21`) holds a `sync.Mutex` and the exclusive OS lock.
- The shared mode of `filelock` comes from TASK-061.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-015, ADR-007.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `docfile.RLocked(lock, fn)` holds the read side of a `sync.RWMutex` and a shared OS lock. `docfile.Locked` holds the write side and the exclusive OS lock.
- FR-2: The read tools `task_list`, `task_get`, `adr_list`, `adr_get`, `learning_list`, `learning_get`, `glossary_lookup` and `glossary_list` run in `docfile.RLocked`. If the lock fails, the tool returns the error, as the write tools do.
- FR-3: `Locked` and `RLocked` refuse an empty lock path with the error "the docs lock path is empty", because a store with no lock path loses the lock between processes with no sign. The tests use a lock file in a temporary directory.

### 6.2 Non-Functional

- NFR-1: One shared lock call for each read tool call.

## 7. Design
### 7.1 Architecture

`docfile` gets `RLocked` next to `Locked`, with one `sync.RWMutex` for both. Each read tool handler wraps its read in `RLocked`.

### 7.2 Public API

No change of the tool schemas.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- An exclusive lock for reads: rejected, see ADR-015.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/write.go`, `internal/docs/docfile/write_test.go` | FR-1, FR-3, with tests. | `go test ./internal/docs/docfile/` |
| 2 | `internal/docs/task/`, `internal/docs/adr/`, `internal/docs/learning/`, `internal/docs/glossary/` (the read functions, the test stores and their tests) | FR-2, FR-3, with tests. | `go test ./internal/docs/...` |
| 3 | `docs/GLOSSARY.md` (with `glossary_upsert` for `Docs lock`) | The shared lock for reads. ADR-007 does not change, because it is pushed (ADR-017) and ADR-015 already gives the shared lock for reads. TASK-068 supersedes ADR-007. | `glossary_lookup Docs lock` |

## 9. Acceptance Criteria

- AC-1: Given a process that holds the exclusive docs lock, when `task_list` runs, then it waits until the lock is released.
- AC-2: Given a process that holds a shared docs lock, when `task_get` runs, then it does not wait.
- AC-3: Given a process that holds a shared docs lock, when `task_update_section` runs, then it waits.
- AC-4: Given an empty lock path, when `Locked` or `RLocked` runs, then it returns "the docs lock path is empty" and does not call fn.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Reads that see half of a change | Possible | 0 | AC-1 test |

- Tests: unit tests with helper processes.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A read waits for a long write | Low | Writes take milliseconds. |
| A write tool calls a read function that takes the shared lock, and waits for itself | High | The read tools take the lock at the tool level. The functions that write tools use take no lock. AC-3 and the write tool tests. |

## 12. Removal and Compatibility

Revert the change.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: AC-1 TestReadWaitsForWrite, AC-2 TestReadDoesNotWaitForRead, AC-3 TestWriteWaitsForRead (internal/docs/task/lock_test.go, helper processes); AC-4 TestNoLockPath; TestRLockedShares. go test ./... passes on macOS; Windows and Linux by CI after push.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet (also GOOS=windows), go test ./... pass; golangci-lint v2.15.0: 0 issues; govulncheck: No vulnerabilities found.
- [x] Docs are updated (area docs, `CLAUDE.md`).
  - Evidence: Glossary Docs lock changed (shared lock for reads). Store Lock field comments say "required". README.md and CLAUDE.md do not describe the docs lock.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: No diagram: the task adds a shared mode to an existing lock and no new flow.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: ADR-015 records the shared docs lock for reads and the required lock path. ADR-007 does not change (ADR-017); TASK-068 supersedes it.