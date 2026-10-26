---
id: TASK-052
title: Keep questions that start with a bracket
status: In Review
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [ADR-013]
---

# TASK-052: Keep questions that start with a bracket

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |

## 1. Summary

`task_question_add` removes only the template rows of section 13, not a real question.

## 2. Problem and Motivation

Found by the code review of 2026-10-26. `withoutPlaceholders` (`internal/docs/task/change.go:379`) removes each question row with a text that starts with `[` or is `None.`. A real question, example `[CLI] keep the -p flag?`, goes away with no message on the next `task_question_add`, and the new question gets its number again. If the removed question is `Open`, the gate for `Approved` (ADR-013) passes.

## 3. Goals

- G-1: `task_question_add` keeps each real question.

## 4. Out of Scope

- Other section 13 tools: `task_question_answer` does not remove rows.

## 5. Context
### 5.1 Current State

- The template row is `noQuestions` (`internal/docs/task/change.go:27`). Old templates had the row `| Q-1 | [Question] | ... |`.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-013: task flow.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `withoutPlaceholders` removes `noQuestions` and a row with the question text `[Question]` exactly. It keeps all other rows, because only these two rows come from a template.

### 6.2 Non-Functional

Not applicable.

## 7. Design
### 7.1 Architecture

Change the condition in `withoutPlaceholders` from a prefix check to an equal check.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Keep the prefix check and refuse a question that starts with `[`: rejected, the rule is not necessary.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/task_test.go` | FR-1, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a question `[CLI] keep the -p flag?`, when `task_question_add` adds a second question, then both questions stay and the second is Q-2.
- AC-2: Given a question `None.`, when `task_question_add` adds a second question, then both questions stay.
- AC-3: Given the template rows `noQuestions` or `[Question]`, when `task_question_add` runs, then the template row goes away.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Real questions removed | Possible | 0 | AC-1, AC-2 tests |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A task file with another old template row keeps that row | Low | The row stays as a question that the user can answer. |

## 12. Removal and Compatibility

Revert the change.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
