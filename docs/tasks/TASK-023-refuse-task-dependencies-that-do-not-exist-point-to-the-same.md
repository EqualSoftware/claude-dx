---
id: TASK-023
title: Refuse task dependencies that do not exist, point to the same task or make a cycle  # no brackets; quote it if it has ": " or " #"
status: Draft  # Draft / In Review / Approved / In Development / Blocked / Done / Cancelled
priority: P1  # P0 / P1 / P2 / P3
area: [docs-tools]  # the areas of the project, for example [api], [docs, ci]
depends_on: [TASK-006]  # for example [TASK-004]
---

# TASK-023: Refuse task dependencies that do not exist, point to the same task or make a cycle

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-006, TASK-006 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

`task_create` and `task_update_fields` check `depends_on`: each ID must exist, must not be the same task, and must not make a cycle.

## 2. Problem and Motivation

In the v0.1.0 QA session (scratch repository, binary from `7063947`), the tools accepted with no warning:
- `task_create(depends_on: [TASK-099])`, and TASK-099 does not exist.
- `task_update_fields(TASK-002, depends_on: [TASK-001])` while TASK-001 depends on TASK-002 (a cycle).
- `task_update_fields(TASK-002, depends_on: [TASK-002])` (the same task).

## 3. Goals

- G-1: The dependencies of all tasks make a graph with no cycles, and each ID exists.

## 4. Out of Scope

- A check of the dependencies that are already in task files.
- A check of the status of a dependency.

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
- FR-4: On a refusal, no file changes.

### 6.2 Non-Functional
- NFR-1: One read of the task frontmatter for each check. No extra parse of the task body.

## 7. Design

### 7.1 Architecture
One check function in `internal/docs/task/`, called before the write in `create.go` (field `DependsOn`, `internal/docs/task/create.go:37`) and `change.go` (`internal/docs/task/change.go:72`).

### 7.2 Public API
No change to the tool inputs.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- A warning for an ID that does not exist: rejected by the maintainer in the QA chat.

## 8. Implementation Plan
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: When `task_create(depends_on: [TASK-099])` runs and TASK-099 does not exist, then the tool refuses it and no file is created.
- AC-2: Given TASK-001 depends on TASK-002, when `task_update_fields(TASK-002, depends_on: [TASK-001])` runs, then the tool refuses it with the cycle path.
- AC-3: When `task_update_fields(TASK-002, depends_on: [TASK-002])` runs, then the tool refuses it.

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

Delete the check function and its two calls. Existing task files are not changed.

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
