---
id: TASK-047
title: Atomic writes and locks for doc edits
status: Approved
priority: P2
milestone: v0.1.0
type: fix
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-040, TASK-042, TASK-043, TASK-044, TASK-048]
related: [ADR-006, ADR-007, TASK-037, TASK-039]
---

# TASK-047: Atomic writes and locks for doc edits

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |

## 1. Summary

Each change of a dx document (task, ADR, learning, glossary, README index) holds a lock of the project, and writes the file to a temporary file that it renames. Two sessions do not lose an edit, and no reader sees a partial file.

## 2. Problem and Motivation

The number allocation has a lock between processes (ADR-007, `internal/numbering/numbering.go:28`), but the document edits do not. Each edit reads the file, changes it and writes it with `os.WriteFile`, which truncates the file first:
- `internal/docs/task/task.go:210`, `internal/docs/task/index.go:70`
- `internal/docs/adr/store.go:312`, `internal/docs/adr/index.go:31`
- `internal/docs/learning/learning.go:118`, `internal/docs/learning/index.go:34`
- `internal/docs/glossary/glossary.go:171`

Two sessions in one repository (example: two worktrees, or two terminals) can change the same file at the same time. The second write removes the change of the first, and a reader can see a partial file.

## 3. Goals

- G-1: No edit of a dx document is lost when two sessions change it at the same time.
- G-2: A reader always sees the old or the new content of a file.

## 4. Out of Scope

- Files that a person edits by hand while a tool writes.
- The plugin files (`internal/plugin/plugin.go`), already written to a temporary directory and renamed.

## 5. Context
### 5.1 Current State

`filelock.Lock` (`internal/filelock/filelock.go`) waits for an exclusive OS lock. `numbering.Allocator` writes `counters.json` to a temporary file and renames it (`internal/numbering/numbering.go:95`). The project key is in `project.Project.Key` (`internal/project/project.go`), and the allocator uses `<cache>/projects/<key>` (`internal/mcpserver/server.go`).

### 5.2 Assumptions

- A-1: On Windows, `os.Rename` replaces a file that another process reads with Go `os.Open`. Verify: a Windows CI test that renames over an open file.

### 5.3 Constraints

- ADR-007: the lock and the counter.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `docfile.Write(path, data)` writes data to a temporary file in the same directory, then renames it to path. A failure removes the temporary file.
- FR-2: All writes of section 2 use `docfile.Write`.
- FR-3: Each tool that changes a document holds an exclusive lock on `<cache>/projects/<key>/docs.lock` for the full read, change and write, and the README index. A mutex orders the calls of one process.
- FR-4: The docs lock is a different file from the numbering lock. A create takes the docs lock first, then the numbering lock, so the order is always the same and no deadlock is possible.
- FR-5: Tools that only read do not take the lock.

### 6.2 Non-Functional

- NFR-1: The lock adds no network call and one small file.

## 7. Design
### 7.1 Architecture

A `Lock` field (the path of `docs.lock`) in the task, ADR, learning and glossary stores. `mcpserver.New` sets it. One helper takes the mutex and the file lock and runs the change.

### 7.2 Public API

No change to the tools.

### 7.3 Data and File Formats

`<cache>/projects/<key>/docs.lock`: empty file, only for the lock.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Atomic writes only, with no lock: rejected by the maintainer, because it does not stop a lost edit (grill Q-9).

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/docfile.go`, `internal/docs/docfile/docfile_test.go` | FR-1 and the lock helper. Test A-1. | `go test ./internal/docs/docfile/` |
| 2 | `internal/docs/task/`, `internal/docs/adr/`, `internal/docs/learning/`, `internal/docs/glossary/`, `internal/mcpserver/server.go` | FR-2 to FR-5. A test with many goroutines that change one task: no change is lost. | `go test -race ./internal/...` |
| 3 | ADR-007 | Add the docs lock to the Decision. | Read ADR-007 |

## 9. Acceptance Criteria

- AC-1: Given 20 goroutines that each add a question to one task, when they end, then the task has 20 questions.
- AC-2: Given a write that fails, then the old file stays and no temporary file stays.
- AC-3 (Windows): Given a file that another reader has open, when `docfile.Write` replaces it, then the write passes.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Lost edits with parallel writers | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC, with `-race` on macOS and Linux.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Windows refuses the rename over a file in use | Medium | AC-3 test. If it fails, retry the rename for a short time. |
| A tool waits for a lock of a session that hangs | Low | The lock is released when the process ends. |

## 12. Removal and Compatibility

Remove the lock helper and the `Lock` fields, and use `os.WriteFile` again.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Atomic writes only, or atomic writes and a lock? | Equal | Both (grill Q-9). |
| Q-2 | Do this task last, after the other fixes of the stores? | Equal | Yes, also after TASK-044 (grill Q-15, Q-26). |
| Q-3 | One `docs.lock` for each project, no lock for read-only tools, size M? | Equal | Yes (grill Q-28). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (ADR-007).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions (ADR-007 gets the docs lock).