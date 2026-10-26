---
id: TASK-025
title: Compact Definition of Done message
status: Cancelled
priority: P3
area: [docs-tools]
depends_on: [TASK-006]
---

# TASK-025: Compact Definition of Done message

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-005, TASK-006 |

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Merged into TASK-023 (FR-8, AC-7), because both change internal/docs/task/change.go (ADR-013 merge rule). |

## 1. Summary

When `task_set_status` refuses `Done`, it gives the unchecked Definition of Done items in one short phrase.

## 2. Problem and Motivation

`internal/docs/task/change.go:144` adds one full phrase for each unchecked item. In the v0.1.0 QA session, the refusal was: `Definition of Done item 1 is not checked; Definition of Done item 2 is not checked; ...` (6 times). ADR-005 asks for compact results.

## 3. Goals

- G-1: The refusal for unchecked items is one phrase.

## 4. Out of Scope

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
- FR-1: The refusal gives `Definition of Done items 2, 3, 4 are not checked`, or `Definition of Done item 2 is not checked` for one item.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
Collect the item numbers at `internal/docs/task/change.go:144`, then make one phrase. Change the test at `internal/docs/task/task_test.go:133`.

### 7.2 Public API
Not applicable.

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

- AC-1: Given a task with items 2 to 6 not checked, when `task_set_status(Done)` runs, then the refusal has one phrase with the numbers 2, 3, 4, 5, 6.

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

Revert the message change.

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
