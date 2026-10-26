---
id: TASK-062
title: Lock the docs for read tools
status: In Review
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
- FR-2: The read tools `task_list`, `task_get`, `adr_list`, `adr_get`, `learning_list`, `learning_get`, `glossary_lookup` and `glossary_list` run in `docfile.RLocked`.

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
| 1 | `internal/docs/docfile/write.go`, `internal/docs/docfile/write_test.go` | FR-1, with tests. | `go test ./internal/docs/docfile/` |
| 2 | `internal/docs/task/`, `internal/docs/adr/`, `internal/docs/learning/`, `internal/docs/glossary/` (the read functions and their tests) | FR-2, with tests. | `go test ./internal/docs/...` |
| 3 | ADR-007 (with `adr_update`), `docs/GLOSSARY.md` (with `glossary_upsert` for `Docs lock`) | The shared lock for reads. | `glossary_lookup Docs lock` |

## 9. Acceptance Criteria

- AC-1: Given a process that holds the exclusive docs lock, when `task_list` runs, then it waits until the lock is released.
- AC-2: Given a process that holds a shared docs lock, when `task_get` runs, then it does not wait.
- AC-3: Given a process that holds a shared docs lock, when `task_update_section` runs, then it waits.

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
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
