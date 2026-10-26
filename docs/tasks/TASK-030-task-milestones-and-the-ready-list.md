---
id: TASK-030
title: Task milestones and the ready list
status: Approved
priority: P1
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools, plugin]
depends_on: [TASK-036]
related: [ADR-013, ADR-006, TASK-006, TASK-018, TASK-033]
---

# TASK-030: Task milestones and the ready list

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

## 1. Summary

Add the required frontmatter field `milestone`, the `task_list` filters `milestone` and `status: ready`, and the milestone groups in `INDEX.md` (ADR-013, steps 1 and 4).

## 2. Problem and Motivation

No field tells the release of a task. The README roadmap gives number ranges (v0.1 = TASK-001 to TASK-008), but TASK-017 to TASK-032 are also v0.1.0 work. `task_list` gives all open tasks, so a person cannot see the work that can start now.

## 3. Goals

- G-1: One `task_list` call gives the open work of one milestone.
- G-2: One `task_list` call gives the tasks that can start now.

## 4. Out of Scope

- The dependency check for a later milestone: TASK-023.
- A milestone file or a release date.

## 5. Context
### 5.1 Current State

`Front` (`internal/docs/task/task.go:55`) has no milestone. `WriteIndex` (`internal/docs/task/index.go:50`) makes one open and one closed table (`split`, `:70`). The README roadmap (`README.md:26` to `:30`) gives task number ranges for each version.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013 (milestone, ready list), ADR-012 (SemVer), ADR-006.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional

- FR-1: `milestone` is a frontmatter field after `priority` (ADR-013). `task_create` requires it. `task_update_fields` can change it. The tools refuse a value that is not `vMAJOR.MINOR.PATCH`, example `v0.1.0`, because the ready order compares versions (Q-3).
- FR-2: `task_list(milestone)` gives only the tasks of that milestone. The filter of each frontmatter field (TASK-033) does this with no new filter code.
- FR-3: `task_list(status: ready)` gives the `Approved` tasks that have all dependencies `Done`, in SemVer order of milestone, then priority, then ID, because only an `Approved` task can start (TASK-034, Q-1).
- FR-4: `task_list` rows show the milestone.
- FR-5: The task index `docs/tasks/README.md` (TASK-036) has one table of open tasks for each milestone, in SemVer order, and one table of closed tasks at the end, because closed tasks are history (Q-2).
- FR-6: All existing tasks get a milestone (Q-4): `v0.1.0` for TASK-001 to TASK-008 and TASK-017 to TASK-036; `v0.2.0` for TASK-009 to TASK-013; `v0.3.0` for TASK-014; `v0.4.0` for TASK-015; `v0.5.0` for TASK-016. The title of TASK-014 loses "(v0.3)".
- FR-7: The README roadmap gives the milestones, not task number ranges.
- FR-8: The `/dx:task` skill tells to use `status: ready` to select the next work.

### 6.2 Non-Functional
- NFR-1: `status: ready` reads only the frontmatter of the task files, one time each.

## 7. Design

### 7.1 Architecture
`Front` (`internal/docs/task/task.go:55`) gets `Milestone`. The template gets `milestone: ""`. `status: ready` is a case in `matchField` (`internal/docs/task/task.go:107`) that also needs the status of the dependencies, so `List` reads all tasks first. `WriteIndex` (`internal/docs/task/index.go:50`) groups the open tasks by milestone; `split` (`:70`) stays for the closed table.

### 7.2 Public API
```text
task_create(title, priority, milestone, type, area, ...)
task_update_fields(id, milestone?, ...)
task_list(status?: open|all|ready|<status>, milestone?, <other frontmatter field>?)
```

### 7.3 Data and File Formats

New frontmatter line `milestone: v0.1.0`, after `priority`. `docs/tasks/README.md` has a `## v0.1.0` heading and table for each milestone, then `## Closed`.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- An area such as `v0.1`: rejected (ADR-013).
- One open and one closed table for each milestone: rejected (Q-2).

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/template.md`, `internal/docs/task/task.go`, `internal/docs/task/create.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-4. Tests for AC-1 to AC-4. | `go test ./internal/docs/task/` |
| 2 | `internal/docs/task/index.go`, `internal/docs/task/task_test.go` | FR-5. | `go test ./internal/docs/task/` |
| 3 | `docs/tasks/*.md`, `README.md`, `internal/plugin/assets/skills/task/SKILL.md` | FR-6 with `task_update_fields`, FR-7, FR-8. | `task_list(status: all)` shows a milestone in each row; `TestIndexMatchesRepository` passes |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs with no milestone, or with `milestone: 0.1`, then the tool refuses it.
- AC-2: Given tasks in v0.1.0 and v0.2.0, when `task_list(milestone: v0.1.0)` runs, then the result has only v0.1.0 tasks.
- AC-3: Given an `Approved` TASK-A with a dependency that is not `Done`, an `Approved` TASK-B with all dependencies `Done`, and a `Draft` TASK-C with no dependencies, when `task_list(status: ready)` runs, then the result has only TASK-B.
- AC-4: Given tasks in v0.2.0 and v0.1.0, when `task_list(status: ready)` runs, then the v0.1.0 tasks come first.
- AC-5: `docs/tasks/README.md` has one open table for each milestone and one closed table.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tool calls to find the next work | `task_list` and reads of each dependency | 1 `task_list(status: ready)` | AC-3 test |

- Tests: unit tests in `internal/docs/task/task_test.go`; `task_list` after work batch 3.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A task in work batch 3 gets the wrong milestone | Low | FR-6 gives the full list. Compare `task_list(milestone: ...)` with it. |

## 12. Removal and Compatibility

Delete the field, the filters and the INDEX groups. Files with a `milestone` line stay valid.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Does status: ready give only Approved tasks with all dependencies Done? Recommended: yes, because only an Approved task can start (TASK-034). Drafts and In Review tasks are not ready. | Equal | Yes. status: ready gives only Approved tasks with all dependencies Done. |
| Q-2 | INDEX.md layout with milestones. Recommended: one table of open tasks for each milestone, in SemVer order, and one Closed table at the end as now, because closed tasks are history. | Equal | Yes. One open table for each milestone in SemVer order, and one Closed table at the end. |
| Q-3 | Do the tools refuse a milestone that is not vMAJOR.MINOR.PATCH (example v0.1.0)? Recommended: yes, because the ready order compares versions. | Equal | Yes. Refuse a milestone that is not vMAJOR.MINOR.PATCH. |
| Q-4 | Milestones of existing tasks. Recommended: v0.1.0 for TASK-001 to TASK-008 and TASK-017 to TASK-035 (except TASK-027, see its Q-1); v0.2.0 for TASK-009 to TASK-013; v0.3.0 TASK-014; v0.4.0 TASK-015; v0.5.0 TASK-016. Also remove "(v0.3)" from the title of TASK-014. | Equal | Yes, and TASK-027 is also v0.1.0 (TASK-027 Q-1). Remove "(v0.3)" from the title of TASK-014. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README roadmap, `/dx:task` skill, ADR-006).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions (ADR-013 has the decision).
- [ ] `/dx:eureka` is done.