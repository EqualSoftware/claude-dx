---
id: TASK-043
title: Warn about ADR and learning read errors
status: Done
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [TASK-005, TASK-007]
---

# TASK-043: Warn about ADR and learning read errors

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`adr_list` and `learning_list` list the files that they can read, and give one warning line for each file that they cannot read.

## 2. Problem and Motivation

`adr.Store.List` (`internal/docs/adr/store.go:83`) and `learning.Store.List` (`internal/docs/learning/learning.go:34`) stop at the first file that they cannot read, or at a learning with a bad frontmatter. Then the full list fails. `task_list` gives a warning and continues (`internal/docs/task/task.go:121`).

## 3. Goals

- G-1: One bad file does not hide the other ADRs or learnings.

## 4. Out of Scope

- A change of the ADR or learning file format.

## 5. Context
### 5.1 Current State

`task.Store.List` adds `Warning: <file>: <error>.` for a bad file. The ADR and learning README indexes use `List` too (`internal/docs/adr/index.go`, `internal/docs/learning/index.go`).

### 5.2 Assumptions

- None.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `adr.Store.List` and `learning.Store.List` skip a file that they cannot read or parse, and return one line `Warning: <file>: <error>.` for it.
- FR-2: `adr_list` and `learning_list` add the warnings after the rows, as `task_list` does.
- FR-3: The README indexes skip the bad files.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

`List` returns the entries and a warnings string, as `task.Store.List` does.

### 7.2 Public API

`adr_list` and `learning_list` results can end with warning lines.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/adr/store.go`, `internal/docs/adr/tools.go`, `internal/docs/adr/index.go`, `internal/docs/adr/store_test.go` | FR-1 to FR-3 for ADRs. | `go test ./internal/docs/adr/` |
| 2 | `internal/docs/learning/learning.go`, `internal/docs/learning/tools.go`, `internal/docs/learning/index.go`, `internal/docs/learning/learning_test.go` | FR-1 to FR-3 for learnings. | `go test ./internal/docs/learning/` |

## 9. Acceptance Criteria

- AC-1: Given two learnings and one with a bad frontmatter, when `learning_list` runs, then it gives the two learnings and one warning.
- AC-2: Given an ADR file that cannot be read, when `adr_list` runs, then it gives the other ADRs and one warning.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Lists that fail because of one bad file | All | 0 | AC tests |

- Tests: one unit test for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The model does not see a warning | Low | The warning is in the result text, as in `task_list`. |

## 12. Removal and Compatibility

Revert the change.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Warn and continue, as `task_list` does, for ADRs and learnings? | Equal | Yes (grill Q-10). |
| Q-2 | The README indexes skip the bad files? | Equal | Yes (grill Q-29). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: TestListBadFrontmatter (AC-1, FR-3): learning_list gives a-topic and b-topic, then "Warning: bad.md: the frontmatter is not valid..."; README.md does not have bad.md. TestListBadFile (AC-2): adr_list gives ADR-001 and ADR-002 and one "Warning: ADR-003-bad.md:" line; WriteIndex does not fail. Commit fc8364a.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues; govulncheck: No vulnerabilities found. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (no docs change expected).
  - Evidence: Not applicable: no doc describes the list errors. The List comments in internal/docs/adr/store.go and internal/docs/learning/learning.go describe the warnings.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: Not applicable: a change in two List functions, no design diagram.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: the ADR and learning lists use the warning rule of task_list, with no change of structure or file format.