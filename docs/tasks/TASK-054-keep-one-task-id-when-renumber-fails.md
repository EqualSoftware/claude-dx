---
id: TASK-054
title: Keep one task ID when renumber fails
status: In Review
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-053]
related: [TASK-049, ADR-007]
---

# TASK-054: Keep one task ID when renumber fails

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |

## 1. Summary

A failed `task_renumber` leaves the task file with its old ID in the name and in the content, and `task_renumber` refuses a file with bad frontmatter.

## 2. Problem and Motivation

Found by the code review of 2026-10-26.
- `renumber` (`internal/docs/task/change.go:586`) writes the new ID into the old file, then calls `docfile.Move`. If `Move` fails (example: on Windows a reader holds the file after the retry), the file `TASK-005-x.md` has `id: TASK-012`. `Alloc.Next` does not write the counter (`internal/numbering/numbering.go:49`), so the next create also gets TASK-012. Two tasks then have one ID.
- `renumber` ignores the error of `markdown.Decode` (`internal/docs/task/change.go:579`). With bad frontmatter the title is empty, and `setHeader` writes `# TASK-NNN: ` with no title.

## 3. Goals

- G-1: After a failed `task_renumber`, the name and the content of the file have the same ID.
- G-2: `task_renumber` does not remove the title of a task.

## 4. Out of Scope

- The rename on a title change (TASK-049): its file has one ID in all cases.

## 5. Context
### 5.1 Current State

- `docfile.Write` replaces a file with one rename. `docfile.Move` refuses a path that exists (TASK-049).

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-007: numbering lock and docs lock.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `renumber` refuses a file with bad frontmatter, with the Decode error, because the header needs the title.
- FR-2: If `docfile.Move` fails, `renumber` writes the old content back to the old path with `docfile.Write` and returns the error of `Move`, because the file must keep one ID.

### 6.2 Non-Functional

Not applicable.

## 7. Design
### 7.1 Architecture

In the create function of `Alloc.Next`: write the new content, then `Move`. On a `Move` error, `docfile.Write(old, data)` and return the error.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Move first, then write the new content: rejected, a failed write then leaves the new number in the name and the old ID in the content.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/task_test.go` | FR-1, FR-2, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a target path that exists, so `Move` fails, when `task_renumber` runs, then it fails and the old file has its old content.
- AC-2: Given a file with bad frontmatter, when `task_renumber` runs, then it refuses and the file does not change.
- AC-3: Given a duplicate number, when `task_renumber` runs, then the task has one file with the new number (TASK-049 AC-4 stays true).

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Files with a different ID in the name and the content after a failed renumber | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The write back also fails | Low | The error of `task_renumber` has both errors, so the user can fix the file by hand. |

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
