---
id: TASK-056
title: Refuse a used title in a title change
status: Approved
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-054]
related: [TASK-042]
---

# TASK-056: Refuse a used title in a title change

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |

## 1. Summary

`task_update_fields` refuses a new title that another task has, as `task_create` does.

## 2. Problem and Motivation

Found by the code review of 2026-10-26. `Update.check` (`internal/docs/task/change.go:112`) calls only `docfile.CheckTitle`. `task_create` also calls `checkNewTitle` (`internal/docs/task/create.go:61`), because one title names one thing. So a title change can give two tasks the same title.

## 3. Goals

- G-1: No two tasks have the same title after a title change.

## 4. Out of Scope

- ADR titles: `adr_update` does not change a title.

## 5. Context
### 5.1 Current State

- `checkNewTitle` (`internal/docs/task/create.go:112`) refuses a title of another task in all statuses, with no case difference.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: A title change calls `checkNewTitle` and ignores the title of the task that changes, so a change of only the case passes.

### 6.2 Non-Functional

Not applicable.

## 7. Design
### 7.1 Architecture

`UpdateFields` calls `checkNewTitle` with the ID of the task to ignore.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/create.go`, `internal/docs/task/task_test.go` | FR-1, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given the title of TASK-001, when `task_update_fields` sets it on TASK-002, then it refuses and the file does not change.
- AC-2: Given TASK-001, when `task_update_fields` changes only the case of its title, then it passes.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tasks with the same title | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| None | — | — |

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
