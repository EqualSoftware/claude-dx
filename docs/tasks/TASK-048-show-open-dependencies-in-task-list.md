---
id: TASK-048
title: Show open dependencies in task_list
status: In Review
priority: P2
milestone: v0.1.0
type: feat
size: S
owner: Equal
platforms: [Not applicable]
area: [docs-tools]
depends_on: []
related: [TASK-030, ADR-013]
---

# TASK-048: Show open dependencies in task_list

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |

## 1. Summary

Each `task_list` row shows the status of each dependency that is not `Done`. The `task_list` description says that the filters combine: a task must match all of them.

## 2. Problem and Motivation

The `deps:` column of `task_list` (`internal/docs/task/tools.go:201`) gives only the IDs of the dependencies. A list with `milestone: v0.1.0` does not show which tasks wait for open work. The agent then ran a second `task_list` with `status: ready` to answer one question. `Store.List` already reads the status of all tasks (`internal/docs/task/task.go:134`), so the data is available with no more I/O.

The tool description (`internal/docs/task/tools.go:122`) does not say how the filters combine. `Filter.match` (`internal/docs/task/task.go:100`) requires all of them.

## 3. Goals

- G-1: One `task_list` call shows which tasks of the list wait for a dependency that is not `Done`.
- G-2: The `task_list` description says that a task must match all filters.

## 4. Out of Scope

- The `README.md` index of `docs/tasks/` (`internal/docs/task/index.go`).
- A new filter, example `status: blocked`.
- The `/dx:task` skill text.
- The order of the rows.

## 5. Context
### 5.1 Current State

- `listText` (`internal/docs/task/tools.go:187`) makes one row for each task. The last column is `deps:` and the IDs, or `deps:None`.
- `Store.List` (`internal/docs/task/task.go:120`) reads all task files and keeps `done[id]` for the `ready` filter.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-013: the task tools check the flow.
- No new I/O: use the data that `Store.List` reads.

## 6. Requirements
### 6.1 Functional

- FR-1: In the `deps:` column, a dependency that is `Done` shows only its ID. A dependency that is not `Done` shows its ID and its status in parentheses, example `deps:TASK-037, TASK-040 (Approved)`. Thus the agent sees the blocked tasks in one call.
- FR-2: A dependency with no task file shows `(missing)`, because the agent must not think that it is `Done`.
- FR-3: The `task_list` description says: a task must match all filters, and a dependency that is not `Done` shows its status.

### 6.2 Non-Functional

- NFR-1: No more file reads than now.

## 7. Design
### 7.1 Architecture

`Store.List` keeps the status of each task ID (it replaces the `done` map). After the loop, it sets a new field `Task.Deps` of each selected task: the `deps:` text of FR-1 and FR-2. `listText` writes `Task.Deps`. The `ready` filter reads the same status map.

### 7.2 Public API

`task_list` output: the `deps:` column of FR-1 and FR-2. The input does not change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Example:

```
TASK-040  Approved  P2  v0.1.0  Check only the new task dependencies  [docs-tools]  deps:TASK-037
TASK-042  Approved  P2  v0.1.0  Rename the task file with the title   [docs-tools]  deps:TASK-040 (Approved)
```

### 7.5 Alternatives

- A new column `blocked`: rejected, it makes each row longer and does not tell which dependency.
- Rely on the `/dx:task` skill to run `status: ready`: rejected, the agent does not always load the skill for a list question.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/task.go`, `internal/docs/task/milestone.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-3, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a task with one `Done` dependency and one `Approved` dependency, when `task_list` runs, then the row shows `deps:<done ID>, <approved ID> (Approved)`.
- AC-2: Given a task with a dependency that has no task file, when `task_list` runs, then the row shows `<ID> (missing)`.
- AC-3: Given a filter that does not select the dependency, example `milestone` of another release, when `task_list` runs, then the row still shows the status of the dependency.
- AC-4: The `task_list` description has the FR-3 text.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| `task_list` calls to list the pending and the blocked tasks of a milestone | 2 | 1 | AC-1 test |

- Tests: unit tests in `internal/docs/task/task_test.go`.
- Platforms not verified: none. The change has no platform code.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A parser of the `task_list` text expects only IDs after `deps:` | Low | No code of the project reads this text. The agent reads it. |

## 12. Removal and Compatibility

Remove `Task.Deps` and write the IDs again in `listText`. No file format changes.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | TASK-042 and TASK-047 also change internal/docs/task/task.go and tools.go. Which order? Recommended: TASK-048 has no dependency and goes first, because it is small. Then add TASK-048 to depends_on of TASK-042 and TASK-047. | Maintainer | Yes. TASK-048 has no dependency and goes first. TASK-042 and TASK-047 get TASK-048 in depends_on. A dependency with no task file shows (missing). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
