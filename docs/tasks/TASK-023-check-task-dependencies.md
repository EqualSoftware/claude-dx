---
id: TASK-023
title: Check task dependencies
status: Done
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-030]
related: [ADR-005, ADR-006, ADR-013, TASK-006, TASK-025]
---

# TASK-023: Check task dependencies

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |
| 2026-10-25 | Equal | Status: In Development. |
| 2026-10-25 | Equal | Status: Done. |

## 1. Summary

`task_create` and `task_update_fields` check `depends_on` (ADR-013, step 1). `task_set_status` gives a compact list of the unchecked Definition of Done items (merged from TASK-025, because both change `internal/docs/task/change.go`).

## 2. Problem and Motivation

In the v0.1.0 QA session (scratch repository, binary from `7063947`), the tools accepted with no warning:
- `task_create(depends_on: [TASK-099])`, and TASK-099 does not exist.
- `task_update_fields(TASK-002, depends_on: [TASK-001])` while TASK-001 depends on TASK-002 (a cycle).
- `task_update_fields(TASK-002, depends_on: [TASK-002])` (the same task).
- TASK-021 to TASK-028 were created with `depends_on` on `Done` tasks. This sets no order.

Also, `internal/docs/task/change.go:144` adds one full phrase for each unchecked Definition of Done item: `Definition of Done item 1 is not checked; Definition of Done item 2 is not checked; ...`. ADR-005 asks for compact results.

## 3. Goals

- G-1: The dependencies of open tasks make a graph with no cycles, each ID exists, and each dependency can still set an order.
- G-2: The refusal for unchecked Definition of Done items is one phrase.

## 4. Out of Scope

- A check of the dependencies that are already in task files.
- Other refusal messages.

## 5. Context
### 5.1 Current State

`New.check` (`internal/docs/task/create.go:27`) and `Update.check` (`internal/docs/task/change.go:58`) check the field values, but not `depends_on`. `uncheckedDoD` (`internal/docs/task/change.go:180`) adds one phrase for each item (`:184`).

### 5.2 Assumptions

- A-1: Each task has a `milestone` field (TASK-030). Verify: `task_list(status: all)` after TASK-030.

### 5.3 Constraints

- ADR-013 (dependency rules), ADR-005 (compact results), ADR-006.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional

- FR-1: The tools refuse a dependency ID that has no task file (decision of the maintainer in the QA chat).
- FR-2: The tools refuse the ID of the same task.
- FR-3: The tools refuse a list that makes a cycle, and give the cycle path, example `TASK-001 -> TASK-002 -> TASK-001`.
- FR-4: The tools refuse a dependency that is `Done` or `Cancelled`, and tell to put it in `related`, because it sets no order (ADR-013, Q-1).
- FR-5: The tools refuse a dependency with a later milestone than the task, because the task cannot be done in its milestone (ADR-013).
- FR-6: The result has a warning when a dependency has a lower priority than the task, because the task waits for it.
- FR-7: On a refusal, no file changes.
- FR-8: The refusal for unchecked Definition of Done items is `Definition of Done items 2, 3, 4 are not checked`, or `Definition of Done item 2 is not checked` for one item.

### 6.2 Non-Functional
- NFR-1: One read of the task frontmatter for each check. No extra parse of the task body.

## 7. Design

### 7.1 Architecture
One check function in a new file `internal/docs/task/deps.go`. `Store.Create` (`internal/docs/task/create.go:44`) and `Store.UpdateFields` (`internal/docs/task/change.go:96`) call it before the write, when `depends_on` is given. The message change is in `uncheckedDoD` (`internal/docs/task/change.go:180`).

### 7.2 Public API
No change to the tool inputs. The result can have warning lines.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- A warning for an ID that does not exist: rejected by the maintainer in the QA chat.
- A warning for a `Done` dependency: rejected (ADR-013).

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/deps.go`, `internal/docs/task/create.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-7. One test for each of AC-1 to AC-6. | `go test ./internal/docs/task/` |
| 2 | `internal/docs/task/change.go`, `internal/docs/task/task_test.go` | FR-8 and AC-7. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: When `task_create(depends_on: [TASK-099])` runs and TASK-099 does not exist, then the tool refuses it and no file is created.
- AC-2: Given TASK-001 depends on TASK-002, when `task_update_fields(TASK-002, depends_on: [TASK-001])` runs, then the tool refuses it with the cycle path.
- AC-3: When `task_update_fields(TASK-002, depends_on: [TASK-002])` runs, then the tool refuses it.
- AC-4: Given TASK-001 is `Done` and TASK-003 is `Cancelled`, when `task_create` runs with one of them in `depends_on`, then the tool refuses it and tells to use `related`.
- AC-5: Given TASK-002 in `v0.2.0`, when a `v0.1.0` task gets `depends_on: [TASK-002]`, then the tool refuses it.
- AC-6: Given TASK-002 with `P3`, when a `P1` task gets `depends_on: [TASK-002]`, then the change is done and the result has a warning.
- AC-7: Given a task with items 2 to 6 not checked, when `task_set_status(Done)` runs, then the refusal has one phrase with the numbers 2, 3, 4, 5, 6.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Invalid dependencies that the tools accept | 3 kinds in the QA session | 0 | AC-1 to AC-5 tests |
| Length of the Done refusal for 5 items | 5 phrases | 1 phrase | AC-7 test |

- Tests: unit tests in `internal/docs/task/task_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The cycle check reads all task files at each change | Low | It reads only the frontmatter, one time each (NFR-1). |

## 12. Removal and Compatibility

Delete the check function and its two calls, and revert the message change. Existing task files are not changed.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Do the tools also refuse a dependency on a Cancelled task? Recommended: yes, with the same message as for Done (use related), because a Cancelled task sets no order either. | Equal | Yes. Refuse a dependency on a Cancelled task, with the same message as for Done. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence. `TestDependsOnRules` (`internal/docs/task/task_test.go`): AC-1 TASK-099 is refused and docs/tasks is not made; AC-2 the cycle is refused with `TASK-002 -> TASK-001 -> TASK-002` and the file does not change; AC-3 the same task is refused; AC-4 Done and Cancelled are refused with "Put it in related"; AC-5 the later milestone v0.2.0 is refused; AC-6 a P3 dependency gives a warning. `TestUncheckedDoDMessage`: AC-7 gives "Definition of Done items 2, 3, 4, 5, 6 are not checked".
- [x] Project checks (build, tests, lint) pass. Local on macOS at `168f32f`: `go build`, `go vet`, `go test ./...` pass; golangci-lint v2.15.0: 0 issues. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (`/dx:task` skill: the dependency rules). Commit `168f32f`: `internal/plugin/assets/skills/task/SKILL.md` "Create a task" step 4 gives the dependency rules.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: checks before a write, no new call flow.
- [x] ADRs are recorded for architectural decisions (ADR-013 has the decision). ADR-013 step 1 records the decision.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new reusable lesson.