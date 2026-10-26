---
id: TASK-049
title: Rename task files with one rename
status: In Review
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [TASK-042, TASK-047, ADR-007]
---

# TASK-049: Rename task files with one rename

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |

## 1. Summary

A title change and `task_renumber` move a task file with one `os.Rename`, not with a new file and a remove. A failure never leaves two files with one number.

## 2. Problem and Motivation

`rename` (`internal/docs/task/task.go`, TASK-042) creates the new file, then removes the old file. `renumber` (`internal/docs/task/change.go`) does the same. On Windows, `os.Remove` fails while another process reads the file, because Go opens files with no `FILE_SHARE_DELETE` (learning `go-windows-file-sharing`). The reader can be `task_list` of another session, an antivirus or a search indexer. Then two files have one number: each later edit of the task fails with `TASK-NNN has 2 files`, and the README index is not made again.

## 3. Goals

- G-1: A failed rename or renumber leaves exactly one file for the task.
- G-2: A rename over a file that a reader has open passes on Windows.

## 4. Out of Scope

- The rename of ADR or learning files: no tool renames them.
- The delete of a learning: a failed remove leaves one file, so no bad state.

## 5. Context
### 5.1 Current State

- `docfile.Write` (`internal/docs/docfile/write.go`) writes a temporary file and renames it with `rename`, which tries again for 1 second on Windows.
- The docs lock (TASK-047) is held for the full edit, so no other tool makes the target file during the rename.

### 5.2 Assumptions

- A-1: On Windows, `os.Rename` of a file that a reader has open passes when the reader closes it during the retry. Verify: a test that holds the source open for 100 ms.

### 5.3 Constraints

- ADR-007: the numbering lock and the docs lock.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `docfile.Move(old, path)` refuses a path that exists, then renames old to path with the Windows retry of `docfile.Write`. It refuses an existing path because `os.Rename` replaces a file on Unix.
- FR-2: A title change writes the new content to the old path with `docfile.Write`, then calls `docfile.Move` to the new name. If `Move` fails, the task has one file with the new content and the old name, and the next title change tries the rename again.
- FR-3: `task_renumber` writes the new content to the old path, then moves the file to the new number in the create function of `Alloc.Next`.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

One helper `docfile.Move` next to `docfile.Write`. `editLocked` and `renumber` call it.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Retry `os.Remove` of the old file: rejected, a failure after the retry still leaves two files.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/write.go`, `internal/docs/docfile/write_test.go` | FR-1, with tests. | `go test ./internal/docs/docfile/` |
| 2 | `internal/docs/task/task.go`, `internal/docs/task/change.go`, `internal/docs/task/task_test.go` | FR-2, FR-3, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a path that exists, when `docfile.Move` runs, then it fails and both files do not change.
- AC-2: Given a source file that a reader has open for 100 ms, when `docfile.Move` runs, then it passes (on Windows after the retry).
- AC-3: Given a title change, then the task has one file with the new name and the old file does not exist (TASK-042 AC-1 stays true).
- AC-4: Given `task_renumber`, then the task has one file with the new number.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Task numbers with two files after a failed rename | Possible on Windows | 0 | AC-1, AC-3 tests |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The rename fails after the retry | Low | One file stays, with the new content. The next title change tries again. |

## 12. Removal and Compatibility

Revert the change. No effect on task files.

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
