---
id: TASK-031
title: Warn about open tasks on the same files
status: Draft
priority: P2
type: feat
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-028]
related: [ADR-013]
---

# TASK-031: Warn about open tasks on the same files

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

When a task goes to `In Review` or `Approved`, `task_set_status` warns if another open task gives the same file in its Implementation Plan and no `depends_on` path connects the two tasks (ADR-013, step 2).

## 2. Problem and Motivation

In the v0.1.0 QA session, TASK-023, TASK-025 and TASK-028 changed `internal/docs/task/` with no order between them. Parallel work on them gives rebase conflicts, and a person does not see the overlap.

## 3. Goals

- G-1: Each overlap of files between open tasks has an order or a merge before approval.

## 4. Out of Scope

- A refusal: an overlap can have a reason.
- Overlap of directories that are not in the plan.

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
- FR-1: The tool reads the paths in backticks in the Files column of section 8 of each open task.
- FR-2: For each other open task that gives one of the same paths and has no `depends_on` path to or from this task, the result adds one line: `Warning: TASK-X also changes <path>. Set the order with depends_on, or merge the tasks.`
- FR-3: The status change is done also with a warning, because the user decides.

### 6.2 Non-Functional
- NFR-1: One read of each open task file for each status change to `In Review` or `Approved`.

## 7. Design

### 7.1 Architecture
A check function in `internal/docs/task/`, called by `task_set_status`.

### 7.2 Public API
No change to the inputs. The result can have warning lines.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
Not applicable.

## 8. Implementation Plan
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given TASK-A and TASK-B with `internal/x.go` in section 8 and no dependency, when TASK-A goes to `Approved`, then the result has a warning about TASK-B.
- AC-2: Given the same tasks and TASK-B depends on TASK-A, when TASK-A goes to `Approved`, then the result has no warning.

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

Delete the check function and its call.

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
