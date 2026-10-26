---
id: TASK-032
title: Review again after a scope change
status: Draft
priority: P2
area: [docs-tools, plugin]
depends_on: [TASK-031]
---

# TASK-032: Review again after a scope change

| Field | Value |
| :--- | :--- |
| Type | feat |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-013, TASK-026, TASK-013 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: `task_update_section` on an `Approved` task, for a section under 6, 7 or 9, sets the status to `In Review` and adds a Document History row, because the user must approve the new scope.
- FR-2: The result tells the status change.
- FR-3: The `/dx:task` skill gives the flow of ADR-013: create, refine, approve, select with `status: ready`, develop, close. It gives the merge rule and the rule that a new finding is a new task.

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
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given an `Approved` task, when `task_update_section` changes section 6.1, then the task is `In Review`.
- AC-2: Given an `Approved` task, when `task_update_section` changes section 13, then the status does not change.

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

Delete the status change in `task_update_section` and the skill text.

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
