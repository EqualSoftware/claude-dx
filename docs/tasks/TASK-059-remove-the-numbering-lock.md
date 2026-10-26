---
id: TASK-059
title: Remove the numbering lock
status: Approved
priority: P2
milestone: v0.1.0
type: refactor
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-057, TASK-061, TASK-055]
related: [ADR-007, ADR-015, TASK-047]
---

# TASK-059: Remove the numbering lock

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |
| 2026-10-26 | Equal | Status: In Review. The scope changed in section 6.1. |
| 2026-10-26 | Equal | Status: Approved. |

## 1. Summary

`numbering.Allocator.Next` does not take its own lock and mutex, because each caller holds the docs lock of the same project.

## 2. Problem and Motivation

The three callers of `Alloc.Next` run in `docfile.Locked`: `internal/docs/task/create.go:50` (Next at :72), `internal/docs/adr/store.go:135` (Next at :151) and `internal/docs/task/change.go:560` (Next at :582). The docs lock and the numbering lock are in the same directory `<cache>/projects/<key>/`, so the docs lock already stops two allocations at the same time. The numbering lock (`internal/numbering/numbering.go:34`) and its mutex (`internal/numbering/numbering.go:18`) add a file, a system call and a lock order rule (ADR-007) with no function.

## 3. Goals

- G-1: One lock, the docs lock, protects the numbers and the documents.
- G-2: No lock order rule is necessary.

## 4. Out of Scope

- The counter `counters.json`: it stays, because a task of one worktree is not in the checkout of another worktree (ADR-007).
- The docs lock: no change.

## 5. Context
### 5.1 Current State

- `docfile.Locked` (`internal/docs/docfile/write.go:21`) holds a process mutex and the OS lock `docs.lock`.
- `Next` takes its mutex and `<cache>/projects/<key>/lock`, then reads the counter, calls `create` and writes the counter.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-007: numbers and locks. It needs a change.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `Next` does not take a mutex or a file lock. Its comment says that the caller must hold the docs lock, because the docs lock is the lock of the project.
- FR-2: The tests of `numbering` that run `Next` at the same time hold a lock in the test.
- FR-3: The glossary removes `Numbering lock` and the lock order of `Docs lock`. ADR-007 does not change, because it is pushed (ADR-017) and ADR-015 already says that the docs lock protects the counter. TASK-068 supersedes ADR-007 with one new ADR for all number changes of v0.1.0.

### 6.2 Non-Functional

- NFR-1: One file lock less for each create.

## 7. Design
### 7.1 Architecture

Remove `mu`, the `filelock.Lock` call and the `filelock` import from `internal/numbering/numbering.go`.

### 7.2 Public API

No change of the tools. `numbering.Allocator.Next` has the same signature.

### 7.3 Data and File Formats

`<cache>/projects/<key>/lock` is not made. An old file stays until `cache clean`.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Keep the two locks: rejected, the numbering lock has no function while each caller holds the docs lock.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/numbering/numbering.go`, `internal/numbering/numbering_test.go`, `internal/docs/docfile/write.go` (comment of `Locked`) | FR-1, FR-2, with tests. | `go test ./internal/numbering/ ./internal/docs/...` |
| 2 | `docs/GLOSSARY.md` (with `glossary_delete` and `glossary_upsert`) | FR-3. | `glossary_lookup Docs lock` |

## 9. Acceptance Criteria

- AC-1: Given `numbering.Next`, then it does not open `<cache>/projects/<key>/lock`.
- AC-2: Given two processes that create a task at the same time in one project, then the tasks get different numbers (the existing concurrency test, now with the docs lock).
- AC-3: Given the glossary, then it names only the docs lock for numbers. ADR-007 does not change.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Locks for one create | 2 | 1 | AC-1 test |
| Duplicate numbers | 0 | 0 | AC-2 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A new caller of `Next` does not hold the docs lock | Medium | The comment of `Next` gives the rule, and the callers are in the docs tools only. |

## 12. Removal and Compatibility

Revert the change, ADR-007 and the glossary rows.

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
