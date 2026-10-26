---
id: TASK-026
title: Remove the eureka item from the template
status: Approved
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
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

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

`internal/docs/task/template.md:113` is the `/dx:eureka` item. `TestCheckDoD` (`internal/docs/task/task_test.go:222`) checks 6 items. ADR-006 Decision gives `/dx:eureka` in the Definition of Done of the template.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-006 (task template), TASK-013 (adds the item again).

## 6. Requirements

### 6.1 Functional

- FR-1: Remove the `/dx:eureka` item from `internal/docs/task/template.md`.
- FR-2: TASK-013 (work batch 2, AC-3) adds the item again when the skill exists.
- FR-3: ADR-006 does not give `/dx:eureka` in the Definition of Done of the template, because the template has no such item until TASK-013.

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

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/template.md`, `internal/docs/task/task_test.go`, ADR-006 | FR-1, FR-3. `TestCheckDoD` checks 5 items. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs, then the Definition of Done of the new task has no `/dx:eureka` item.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Definition of Done items that a v0.1 user cannot do | 1 | 0 | AC-1 test |

- Tests: unit tests in `internal/docs/task/task_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| TASK-013 forgets to add the item again | Medium | TASK-013 work batch 2 and AC-3. |

## 12. Removal and Compatibility

Add the line again. Existing tasks keep their 6 items.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |

No open questions. The QA chat answered the change of TASK-013.

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (ADR-006).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.