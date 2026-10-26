---
id: TASK-026
title: Remove the eureka item from the template
status: Draft
priority: P1
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-030]
related: [TASK-006, TASK-013]
---

# TASK-026: Remove the eureka item from the template

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

The task template has the Definition of Done item `/dx:eureka is done` (`internal/docs/task/template.md:118`). The skill comes in v0.2 (TASK-013). Remove the item from the template in v0.1.

## 2. Problem and Motivation

Each v0.1 task must check an item for a skill that does not exist. Example: TASK-019 and TASK-020 checked it with "The tool does not exist yet". Users of v0.1 see an item that they cannot do.

## 3. Goals

- G-1: A new task in v0.1 has only Definition of Done items that a user can do.

## 4. Out of Scope

- A change of the tasks that exist.

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
- FR-1: Remove the `/dx:eureka` item from `internal/docs/task/template.md`.
- FR-2: TASK-013 (work batch 2, AC-3) adds the item again when the skill exists.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
One line in `internal/docs/task/template.md`, and the tests that count the template items.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
The Definition of Done of new tasks has 5 items.

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

- AC-1: When `task_create` runs, then the Definition of Done of the new task has no `/dx:eureka` item.

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

Add the line again. Existing tasks keep their 6 items.

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
