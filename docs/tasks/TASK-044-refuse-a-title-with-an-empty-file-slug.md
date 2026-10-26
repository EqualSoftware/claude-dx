---
id: TASK-044
title: Refuse a title with an empty file slug
status: In Review
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [TASK-028]
---

# TASK-044: Refuse a title with an empty file slug

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |

## 1. Summary

`docfile.CheckTitle` refuses a title that gives an empty slug, so no file name is `TASK-NNN-.md` or `ADR-NNN-.md`.

## 2. Problem and Motivation

`Slug` (`internal/docs/docfile/docfile.go:155`) keeps only the ASCII letters and digits. A title with no ASCII letter or digit, example `ção`, gives an empty slug and the file name `TASK-NNN-.md`. `CheckTitle` (`:21`) accepts it.

## 3. Goals

- G-1: Each new task or ADR file name has a non-empty slug.

## 4. Out of Scope

- A transliteration of non-ASCII letters in the slug.

## 5. Context
### 5.1 Current State

`CheckTitle` is used by `CheckNewTitle` (tasks and ADRs) and by the task title update (`internal/docs/task/change.go:86`).

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013 limits titles; TASK-028 sets 40 characters.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `CheckTitle` refuses a title when `Slug(title)` is empty, with `give a title with ASCII letters or digits`.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

One check in `CheckTitle` (`internal/docs/docfile/docfile.go:21`).

### 7.2 Public API

New refusal message.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/docfile.go`, `internal/docs/docfile/docfile_test.go` | FR-1 with a test. | `go test ./internal/docs/docfile/` |

## 9. Acceptance Criteria

- AC-1: Given the title `ção`, when `CheckTitle` runs, then it refuses it.
- AC-2: Given the title `Ação 2`, then `CheckTitle` accepts it.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| File names with an empty slug | Possible | 0 | AC-1 test |

- Tests: one unit test for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| None | Low | Not applicable. |

## 12. Removal and Compatibility

Delete the check.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Refuse a title with no ASCII letter or digit? | Equal | Yes (grill Q-11). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (no docs change expected).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.