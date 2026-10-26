---
id: TASK-030
title: Task milestones and the ready list
status: Draft
priority: P1
area: [docs-tools, plugin]
depends_on: [TASK-033]
---

# TASK-030: Task milestones and the ready list

| Field | Value |
| :--- | :--- |
| Type | feat |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | M |
| Blocks | None |
| Related | ADR-013, ADR-006, TASK-006, TASK-018 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: `milestone` is a frontmatter field with a SemVer version, example `v0.1.0`. `task_create` requires it. `task_update_fields` can change it.
- FR-2: `task_list(milestone)` gives only the tasks of that milestone.
- FR-3: `task_list(status: ready)` gives the open tasks that are not `Blocked` and that have all dependencies `Done`, in SemVer order of milestone, then priority, then ID, because this is the order of the work.
- FR-4: `task_list` rows show the milestone.
- FR-5: `INDEX.md` has one table for each milestone, in SemVer order.
- FR-6: All existing tasks get a milestone: TASK-001 to TASK-008 and TASK-017 to TASK-032 `v0.1.0`; TASK-009 to TASK-013 `v0.2.0`; TASK-014 `v0.3.0`; TASK-015 `v0.4.0`; TASK-016 `v0.5.0`.
- FR-7: The README roadmap gives the milestones, not task number ranges.
- FR-8: The `/dx:task` skill tells to use `status: ready` to select the next work.

### 6.2 Non-Functional
- NFR-1: `status: ready` reads only the frontmatter of the task files, one time each.

## 7. Design

### 7.1 Architecture
Frontmatter and list in `internal/docs/task/`, `INDEX.md` in `internal/docs/task/index.go:50`.

### 7.2 Public API
```text
task_create(..., milestone)
task_update_fields(id, milestone?)
task_list(status?: open|all|ready|<status>, priority?, area?, milestone?)
```

### 7.3 Data and File Formats
New frontmatter line `milestone: v0.1.0`.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- An area such as `v0.1`: rejected (ADR-013).

## 8. Implementation Plan
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs with no milestone, then the tool refuses it.
- AC-2: Given tasks in v0.1.0 and v0.2.0, when `task_list(milestone: v0.1.0)` runs, then the result has only v0.1.0 tasks.
- AC-3: Given TASK-A with a dependency that is not `Done`, and TASK-B with all dependencies `Done`, when `task_list(status: ready)` runs, then the result has TASK-B and not TASK-A.
- AC-4: `INDEX.md` has one table for each milestone.

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

Delete the field, the filters and the INDEX groups. Files with a `milestone` line stay valid.

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
