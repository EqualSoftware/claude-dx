---
id: TASK-031
title: Warn about open tasks on the same files
status: Approved
priority: P2
type: feat
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-028]
related: [ADR-013]
---

# TASK-031: Warn about open tasks on the same files

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

## 1. Summary

When a task goes to `In Review` or `Approved`, `task_set_status` warns if another open task gives the same file in its Implementation Plan and no `depends_on` path connects the two tasks (ADR-013, step 2).

## 2. Problem and Motivation

In the v0.1.0 QA session, TASK-023, TASK-025 and TASK-028 changed `internal/docs/task/` with no order between them. Parallel work on them gives rebase conflicts, and a person does not see the overlap.

## 3. Goals

- G-1: Each overlap of files between open tasks has an order or a merge before approval.

## 4. Out of Scope

- A refusal: an overlap can have a reason.
- Overlap of directories that are not in the plan.

## 5. Context
### 5.1 Current State

`Store.SetStatus` (`internal/docs/task/change.go:114`) returns only an error, so it cannot give a warning with a status change that passes. `markdown.Section` reads section 8.

### 5.2 Assumptions

- A-1: Each plan gives its files in backticks in the second column of the section 8 table, as the template asks. Verify: `git grep` on `docs/tasks` for section 8 rows.

### 5.3 Constraints

- ADR-013 (step 2), ADR-005.
- Tests pass on macOS, Linux and Windows. Paths in plans use `/`.

## 6. Requirements

### 6.1 Functional

- FR-1: The tool reads the paths in backticks in the Files column of section 8 of each open task, also `Draft` tasks, because a Draft plan already tells which files it changes (Q-2).
- FR-2: Two paths overlap when they are equal or when one is a directory (it ends with `/`) that contains the other, because both change the same code (Q-1).
- FR-3: For each other open task with an overlap and with no `depends_on` path to or from this task, the result adds one line: `Warning: TASK-X also changes <path>. Set the order with depends_on, or merge the tasks.`
- FR-4: The status change is done also with a warning, because the user decides.

### 6.2 Non-Functional
- NFR-1: One read of each open task file for each status change to `In Review` or `Approved`.

## 7. Design

### 7.1 Architecture
A check function in `internal/docs/task/`, called by `task_set_status`.

### 7.2 Public API
No change to the inputs. The result can have warning lines.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/overlap.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-4. `SetStatus` returns the warning lines. One test for each AC. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given TASK-A and TASK-B with `internal/x.go` in section 8 and no dependency, when TASK-A goes to `Approved`, then the result has a warning about TASK-B.
- AC-2: Given the same tasks and TASK-B depends on TASK-A, when TASK-A goes to `Approved`, then the result has no warning.
- AC-3: Given TASK-A with `internal/` and a `Draft` TASK-B with `internal/x.go`, when TASK-A goes to `In Review`, then the result has a warning about TASK-B.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Overlaps with no order that the tools tell | 0 | All | AC-1 and AC-3 tests |

- Tests: unit tests in `internal/docs/task/task_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A wide path, example `docs/tasks/*.md`, gives many warnings | Low | The warning is not a refusal. A glob is not a directory, so it matches only the equal text. |

## 12. Removal and Compatibility

Delete the check function and its call.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Does a directory path in a plan overlap the files in it? Example: internal/docs/task/ and internal/docs/task/change.go. Recommended: yes, because both change the same code. | Equal | Yes. A directory path overlaps the files in it. |
| Q-2 | Do Draft tasks count in the overlap check? Recommended: yes, all open tasks, because a Draft plan already tells which files it changes. | Equal | Yes. All open tasks count, also Draft. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`/dx:task` skill: the overlap warning).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions (ADR-013 has the decision).
- [ ] `/dx:eureka` is done.