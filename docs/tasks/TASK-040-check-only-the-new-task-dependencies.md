---
id: TASK-040
title: Check only the new task dependencies
status: In Review
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-037]
related: [ADR-013, TASK-023]
---

# TASK-040: Check only the new task dependencies

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |

## 1. Summary

`task_update_fields` checks only the dependencies that are not in the old `depends_on` list. A dependency that became `Done` stays in the list with no error.

## 2. Problem and Motivation

`UpdateFields` (`internal/docs/task/change.go:136`) checks the full new `depends_on` list with `checkDeps` (`internal/docs/task/deps.go:13`). `checkDep` (`:44`) refuses a `Done` dependency. When a dependency of a task becomes `Done`, each later change of `depends_on` fails until the model removes that dependency. Example: TASK-013 depends on TASK-026, which is `Done`.

## 3. Goals

- G-1: A change of `depends_on` fails only for a problem of a new dependency, or for a cycle.

## 4. Out of Scope

- The checks of a new dependency (TASK-023).
- A move of `Done` dependencies to `related`.

## 5. Context
### 5.1 Current State

`ready` (`internal/docs/task/milestone.go`) already treats a `Done` dependency as complete.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013, step 1.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `UpdateFields` runs `checkDep` only for the IDs that are not in the old `depends_on` list.
- FR-2: The cycle check runs on the full new list, as now.
- FR-3: `task_create` has no old list, so it checks all IDs, as now.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

`UpdateFields` reads the old `depends_on` before `SetField` and gives it to `checkDeps`. `checkDeps` skips `checkDep` for an old ID.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Move `Done` dependencies to `related` automatically: rejected by the maintainer.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/deps.go`, `internal/docs/task/task_test.go` | FR-1 to FR-3, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a task that depends on a `Done` task, when `task_update_fields` adds a new open dependency and keeps the old one, then the change is done.
- AC-2: Given the same task, when the new dependency is `Done`, then the tool refuses it.
- AC-3: Given a change that makes a cycle, then the tool refuses it.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Refusals caused by an old `Done` dependency | Each change | 0 | AC-1 test |

- Tests: one unit test for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| An old dependency that is not valid stays | Low | It was checked when it was added. |

## 12. Removal and Compatibility

Revert the change. No effect on task files.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Check only the new dependency IDs, and keep old `Done` dependencies? | Equal | Yes (grill Q-6). |
| Q-2 | Order with the other tasks that change `internal/docs/task/change.go`? | Equal | After TASK-037 (grill Q-25). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`/dx:task` skill, step 4 of Create a task).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.