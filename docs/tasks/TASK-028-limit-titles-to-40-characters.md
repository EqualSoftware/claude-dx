---
id: TASK-028
title: Limit titles to 40 characters
status: Draft
priority: P2
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-021, TASK-023]
related: [ADR-006, ADR-013, TASK-005, TASK-006]
---

# TASK-028: Limit titles to 40 characters

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

`task_create`, `task_update_fields` and `adr_create` refuse a title of more than 40 characters (ADR-013, step 1). The slug is never longer than the title, so a file name has a slug of 40 characters or less.

## 2. Problem and Motivation

The tools accept titles of any length, and `docfile.Slug` cuts the slug at 60 characters (`internal/docs/docfile/docfile.go:118`). In the v0.1.0 QA session, this made file names such as `TASK-023-refuse-task-dependencies-that-do-not-exist-point-to-the-same.md`. 30 of 40 titles had more than 40 characters.

## 3. Goals

- G-1: Each new title of a task or an ADR has 40 characters or less, and each new file name has a slug of 40 characters or less.

## 4. Out of Scope

- Existing titles and file names: the QA cleanup commit made them short.
- A title change tool for ADRs.

## 5. Context
### 5.1 Current State

`docfile.Slug` (`internal/docs/docfile/docfile.go:118`) cuts the slug at 60 characters (`:123`). `New.check` (`internal/docs/task/create.go:27`) checks only that the title is not empty. `adr.Store.Create` (`internal/docs/adr/store.go:128`) has no title check.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013 (title limit), ADR-006.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional

- FR-1: `task_create`, `task_update_fields` and `adr_create` refuse a title of more than 40 characters, and give the length and the limit, because a long title gives a long file name (ADR-013).
- FR-2: The `title` field descriptions give the limit.
- FR-3: `docfile.Slug` cuts at 40 characters, at a word boundary, as a guard for old titles.
- FR-4: `task_create` and `adr_create` refuse a title that another task or ADR has, in all statuses, with no case difference, and give the ID of that document, because one title names one thing (Q-1). In the v0.1.0 QA session, `adr_create` made ADR-001 and ADR-002 with the title "Use SQLite for cache".

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
One length check in `internal/docs/docfile/` called by `internal/docs/task/create.go`, `internal/docs/task/change.go` and `internal/docs/adr/store.go`. The cut length in `docfile.Slug` (`internal/docs/docfile/docfile.go:123`).

### 7.2 Public API
No change to the inputs.

### 7.3 Data and File Formats
File names: `TASK-NNN-<slug>.md` and `ADR-NNN-<slug>.md`, slug of 40 characters or less.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- Remove stop words at the end of a cut slug: rejected (ADR-013).
- A limit of 60 or 50 characters: rejected by the maintainer.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/docfile.go`, `internal/docs/docfile/docfile_test.go` | A title check function (length and duplicate), and FR-3. | `go test ./internal/docs/docfile/` |
| 2 | `internal/docs/task/create.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/adr/store.go`, `internal/docs/adr/tools.go`, the tests of both packages | FR-1, FR-2, FR-4. One test for each AC. | `go test ./internal/docs/...` |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs with a title of 41 characters, then the tool refuses it and no file is created.
- AC-2: When `adr_create` runs with a title of 41 characters, then the tool refuses it.
- AC-3: When `task_update_fields` runs with a title of 40 characters, then the change is done.
- AC-4: Given ADR-001 with the title "Use SQLite", when `adr_create` runs with the title "use sqlite", then the tool refuses it and gives ADR-001.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Longest slug of a new file | 60 characters | 40 characters | AC-1 and AC-2 tests |
| Duplicate titles that the tools accept | 1 in the QA session | 0 | AC-4 test |

- Tests: unit tests of `docfile`, `task` and `adr`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A good title has more than 40 characters | Low | The refusal gives the length, so the model writes a shorter title. |

## 12. Removal and Compatibility

Delete the length check and its three calls. Existing files do not change.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Does the duplicate title check compare with all tasks and ADRs, also Done, Cancelled and Superseded? Recommended: yes, because one title names one thing. To work on a Cancelled task again, set it to Draft (TASK-034). | Equal | Yes. Compare with all tasks and ADRs, in all statuses. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`/dx:task` and `/dx:adr` skills: the title limit).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions (ADR-013 has the decision).
- [ ] `/dx:eureka` is done.