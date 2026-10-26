---
id: TASK-057
title: Escape pipes in README index rows
status: Draft
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [ADR-007]
---

# TASK-057: Escape pipes in README index rows

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

The README index of tasks, ADRs and learnings escapes `|` in titles and descriptions, so each row has the correct columns.

## 2. Problem and Motivation

Found by the code review of 2026-10-26. `docfile.CheckTitle` allows `|`. The index writers put the text in table rows with no escape: `internal/docs/task/index.go:65`, `internal/docs/task/index.go:69`, `internal/docs/adr/index.go:30` and `internal/docs/learning/index.go:33`. A title `Fix a|b parse` adds a column, and GitHub shows a broken table.

## 3. Goals

- G-1: Each README index row has the columns of its header.

## 4. Out of Scope

- The glossary: its `cell` escapes `|` already (`internal/docs/glossary/glossary.go:278`).

## 5. Context
### 5.1 Current State

- The task tools have `cell` (`internal/docs/task/task.go:301`) for the question table.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: The index writers replace `|` with `\|` in titles and descriptions, because a title with `|` is valid.

### 6.2 Non-Functional

Not applicable.

## 7. Design
### 7.1 Architecture

One function `docfile.Cell` that escapes `|` and replaces line breaks with a space. The task `cell` and the index writers use it.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

The README index rows have `\|` where a title has `|`.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Refuse `|` in `CheckTitle`: rejected, the escape is one line and keeps valid titles.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/docfile.go`, `internal/docs/task/index.go`, `internal/docs/task/task.go`, `internal/docs/adr/index.go`, `internal/docs/learning/index.go`, and their tests | FR-1, with tests. | `go test ./internal/docs/...` |

## 9. Acceptance Criteria

- AC-1: Given a task, an ADR and a learning with `|` in the title or the description, when the README index is made, then each row has `\|` and the column count of its header.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Broken index rows | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| None | — | — |

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
