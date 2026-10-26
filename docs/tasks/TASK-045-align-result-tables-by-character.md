---
id: TASK-045
title: Align result tables by character
status: Approved
priority: P3
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp]
depends_on: []
related: [ADR-005]
---

# TASK-045: Align result tables by character

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |

## 1. Summary

`result.Table` aligns the columns by the number of characters, not bytes.

## 2. Problem and Motivation

`Table` (`internal/result/result.go:29`) uses `len(cell)`, the number of bytes. A cell with a non-ASCII letter (example `é`) has more bytes than characters, so the next columns move.

## 3. Goals

- G-1: The columns of `task_list` and `adr_list` align when a cell has non-ASCII letters.

## 4. Out of Scope

- Wide characters (CJK, emoji): one character is one column.

## 5. Context
### 5.1 Current State

`Table` is used by `task_list` and `adr_list`.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-005: compact results.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `Table` measures each cell with `utf8.RuneCountInString`.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

A change in `Table` only.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- A display width library for wide characters: rejected, it adds a dependency for a rare case (grill Q-12).

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/result/result.go`, `internal/result/result_test.go` | FR-1 with a test. | `go test ./internal/result/` |

## 9. Acceptance Criteria

- AC-1: Given rows with `é` in the first column, when `Table` runs, then the second column starts at the same character position in each row.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Rows that are not aligned with non-ASCII letters | All | 0 | AC-1 test |

- Tests: AC-1 unit test.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| None | Low | Not applicable. |

## 12. Removal and Compatibility

Revert the change.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Count characters with `utf8.RuneCountInString`, with no wide character rule? | Equal | Yes (grill Q-12). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (no docs change expected).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.