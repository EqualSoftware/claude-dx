---
id: TASK-034
title: Allowed task status changes
status: Draft
priority: P2
type: feat
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-032]
related: [ADR-013, TASK-006]
---

# TASK-034: Allowed task status changes

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

`task_set_status` checks the current status as well as the new status, so a task cannot skip a step of the ADR-013 flow.

## 2. Problem and Motivation

The rules of `task_set_status` check only the new status (`internal/docs/task/change.go:98`). A task can go from `Draft` to `Approved` with no `In Review` checks (template text, Size), and from `Draft` to `In Development` with no check. The ADR-013 flow has an order, but the tools do not keep it.

## 3. Goals

- G-1: Each task passes the `In Review` checks before `Approved`, and is `Approved` before `In Development`.

## 4. Out of Scope

- A change of the checks of each status.
- Existing tasks: the check applies at the next status change.

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: The allowed changes are (Q-1):
  - `Draft` -> `In Review`, `Cancelled`
  - `In Review` -> `Approved`, `Draft`, `Cancelled`
  - `Approved` -> `In Development`, `In Review`, `Blocked`, `Cancelled`
  - `In Development` -> `Done`, `Blocked`, `Cancelled`
  - `Blocked` -> `Approved`, `In Development`, `Cancelled`
  - `Cancelled` -> `Draft`
  - `Done`: no change, because a new finding is a new task (ADR-013, step 5).
- FR-2: A refusal gives the current status and the allowed new statuses, example `Cannot set Approved from Draft. Allowed: In Review, Cancelled.`
- FR-3: ADR-013 gives the table.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
One map from the current status to the allowed statuses in `internal/docs/task/change.go`, checked before the `rules` map (`internal/docs/task/change.go:98`).

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
| 1 | `internal/docs/task/change.go`, `internal/docs/task/task_test.go`, ADR-013 | FR-1 to FR-3. One test for each row of FR-1. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a `Draft` task, when `task_set_status(Approved)` runs, then the tool refuses it and gives `In Review, Cancelled`.
- AC-2: Given a `Done` task, when `task_set_status(In Development)` runs, then the tool refuses it.
- AC-3: Given an `In Review` task with no problems, when `task_set_status(Approved)` runs, then the change is done.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| [Response time (ms)] | [value] | [value] | [tool, setup, steps] |

- Tests: [unit tests, manual steps].
- Platforms not verified: [list, or None].

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| [Risk] | Low / Medium / High | [Action] |

## 12. Removal and Compatibility

Delete the map and its check.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Confirm the table of FR-1. Recommended: as written. Points to check: Done is final (no reopen); Cancelled goes back only to Draft; Blocked goes back to Approved or In Development. | Equal | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
