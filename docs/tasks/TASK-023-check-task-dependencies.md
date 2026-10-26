---
id: TASK-023
title: Check task dependencies
status: Draft
priority: P1
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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: The tools refuse a dependency ID that has no task file (decision of the maintainer in the QA chat).
- FR-2: The tools refuse the ID of the same task.
- FR-3: The tools refuse a list that makes a cycle, and give the cycle path, example `TASK-001 -> TASK-002 -> TASK-001`.
- FR-4: The tools refuse a dependency that is `Done`, and tell to put it in Related, because it sets no order (ADR-013).
- FR-5: The tools refuse a dependency with a later milestone than the task, because the task cannot be done in its milestone (ADR-013).
- FR-6: The result has a warning when a dependency has a lower priority than the task, because the task waits for it.
- FR-7: On a refusal, no file changes.
- FR-8: The refusal for unchecked Definition of Done items is `Definition of Done items 2, 3, 4 are not checked`, or `Definition of Done item 2 is not checked` for one item.

### 6.2 Non-Functional
- NFR-1: One read of the task frontmatter for each check. No extra parse of the task body.

## 7. Design

### 7.1 Architecture
One check function in `internal/docs/task/`, called before the write in `create.go` (field `DependsOn`, `internal/docs/task/create.go:37`) and `change.go` (`internal/docs/task/change.go:72`). The message change is at `internal/docs/task/change.go:144`.

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
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: When `task_create(depends_on: [TASK-099])` runs and TASK-099 does not exist, then the tool refuses it and no file is created.
- AC-2: Given TASK-001 depends on TASK-002, when `task_update_fields(TASK-002, depends_on: [TASK-001])` runs, then the tool refuses it with the cycle path.
- AC-3: When `task_update_fields(TASK-002, depends_on: [TASK-002])` runs, then the tool refuses it.
- AC-4: Given TASK-001 is `Done`, when `task_create(depends_on: [TASK-001])` runs, then the tool refuses it and tells to use Related.
- AC-5: Given TASK-002 in `v0.2.0`, when a `v0.1.0` task gets `depends_on: [TASK-002]`, then the tool refuses it.
- AC-6: Given TASK-002 with `P3`, when a `P1` task gets `depends_on: [TASK-002]`, then the change is done and the result has a warning.
- AC-7: Given a task with items 2 to 6 not checked, when `task_set_status(Done)` runs, then the refusal has one phrase with the numbers 2, 3, 4, 5, 6.

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

Delete the check function and its two calls, and revert the message change. Existing task files are not changed.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | [Question] | [Name] | [Answer, or Open] |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
