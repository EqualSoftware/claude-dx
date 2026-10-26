---
id: TASK-032
title: Review again after a scope change
status: Done
priority: P2
milestone: v0.1.0
type: feat
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools, plugin]
depends_on: [TASK-031]
related: [ADR-013, TASK-026, TASK-013]
---

# TASK-032: Review again after a scope change

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |
| 2026-10-25 | Equal | Status: In Development. |
| 2026-10-25 | Equal | Status: Done. |

## 1. Summary

A change of Requirements, Design or Acceptance Criteria of an `Approved` task sets the task to `In Review`. The `/dx:task` skill gives the full flow of ADR-013 (ADR-013, steps 3, 5 and 6).

## 2. Problem and Motivation

In the v0.1.0 QA session, TASK-026 needed a change of the scope of TASK-013, which is `Approved`. The tools let a section change keep the status, so a task can stay `Approved` with a scope that the user did not approve.

## 3. Goals

- G-1: An `Approved` task has only a scope that the user approved.

## 4. Out of Scope

- Tasks in `In Development`: changes during work are normal. A new finding is a new task (skill text).

## 5. Context
### 5.1 Current State

`Store.UpdateSection` (`internal/docs/task/change.go:36`) replaces the section and does not read the status.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013 (step 3). TASK-034 allows `Approved` -> `In Review`.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional

- FR-1: `task_update_section` on an `Approved` task, for section 3, 4, 6, 7 or 9 or one of their subsections, sets the status to `In Review` and adds a Document History row, because these sections define the scope and the user must approve the new scope (Q-1).
- FR-2: The result tells the status change.
- FR-3: `task_update_fields` does not change the status, because the fields change the order or the estimate, not the scope (Q-2).
- FR-4: The `/dx:task` skill gives the flow of ADR-013: create, refine, approve, select with `status: ready`, develop, close. It gives the merge rule and the rule that a new finding is a new task.
- FR-5: ADR-013 step 3 gives the sections of FR-1.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
`task_update_section` in `internal/docs/task/`, and `internal/plugin/assets/skills/task/SKILL.md`.

### 7.2 Public API
No change to the inputs.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-3. One test for each AC. | `go test ./internal/docs/task/` |
| 2 | `internal/plugin/assets/skills/task/SKILL.md`, ADR-013 | FR-4, FR-5. | Read the diff |

## 9. Acceptance Criteria

- AC-1: Given an `Approved` task, when `task_update_section` changes section 6.1, then the task is `In Review` and the result tells it.
- AC-2: Given an `Approved` task, when `task_update_section` changes section 13, then the status does not change.
- AC-3: Given an `Approved` task, when `task_update_fields` changes the priority, then the status does not change.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Scope changes that keep `Approved` | All | 0 | AC-1 test |

- Tests: unit tests in `internal/docs/task/task_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A typo fix in section 6 asks for a new approval | Low | The user approves again with one call. |

## 12. Removal and Compatibility

Delete the status change in `task_update_section` and the skill text.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which sections set an Approved task back to In Review? Recommended: 3 (Goals), 4 (Out of Scope), 6 (Requirements), 7 (Design) and 9 (Acceptance Criteria), because all of them define the scope. | Equal | Yes. Sections 3, 4, 6, 7 and 9. |
| Q-2 | Does task_update_fields on an Approved task also set it to In Review? Recommended: no, because the fields (priority, depends_on, size) change the order or the estimate, not the scope. | Equal | No. task_update_fields does not change the status. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence. `TestScopeChangeNeedsReview` (`internal/docs/task/task_test.go`): AC-1 section 6.1 sets In Review and adds the history row, and the tool result tells it; AC-2 section 13 keeps Approved; AC-3 `task_update_fields` keeps Approved.
- [x] Project checks (build, tests, lint) pass. Local on macOS at `05c3179`: `go build`, `go vet`, `go test ./...` pass; golangci-lint v2.15.0: 0 issues. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (`/dx:task` skill, ADR-013). Commit `05c3179`: the `/dx:task` skill has the Flow section (ADR-013 steps, the merge rule, a new finding is a new task); ADR-013 step 3 gives sections 3, 4, 6, 7 and 9.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: one status change in one tool. The skill Flow section gives the process.
- [x] ADRs are recorded for architectural decisions (ADR-013 has the decision). ADR-013 step 3 records the decision and the sections.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new reusable lesson.