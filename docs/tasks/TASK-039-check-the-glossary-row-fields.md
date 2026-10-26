---
id: TASK-039
title: Check the glossary row fields
status: Done
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [TASK-007, TASK-022]
---

# TASK-039: Check the glossary row fields

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`glossary_upsert` refuses a row with an empty term, definition, area or source, and a Do not use word that has a comma.

## 2. Problem and Motivation

`Store.Upsert` (`internal/docs/glossary/glossary.go:89`) does not check the fields. A probe test added a row with an empty term and an empty definition. A Do not use word with a comma (`"a, b"`) is written as one word, but `words` (`:207`) reads it back as two words.

## 3. Goals

- G-1: Each glossary row has a term, a definition, an area and a source.
- G-2: The Do not use words that the tool writes are the words that it reads back.

## 4. Out of Scope

- Existing rows: the check applies at the next upsert of the row.
- `glossary_lookup`, `glossary_list` and `glossary_delete`.

## 5. Context
### 5.1 Current State

`upsertInput` (`internal/docs/glossary/tools.go:21`) has `do_not_use` with `omitempty`. `area` has no `omitempty`, so the schema already marks it as required. `Upsert` keeps the old Do not use words when `DoNotUse` is nil (TASK-022).

### 5.2 Assumptions

- None.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `Upsert` refuses a row with an empty `term`, `definition`, `area` or `source` after trim. One message gives all empty fields, example `Give the definition and the area.`
- FR-2: `do_not_use` stays optional.
- FR-3: `Upsert` refuses a Do not use word that is empty or has a comma. The message says: give each word as one list item.
- FR-4: The tool schema marks `area` as required, and its description says that all fields but `do_not_use` are required.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

One check function for the row in `internal/docs/glossary/glossary.go`, called at the start of `Upsert`.

### 7.2 Public API

`glossary_upsert`: `area` stays required (no `omitempty`, as before). The tool description says that all fields but `do_not_use` are required.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Escape the comma in the file: rejected by the maintainer, because a synonym with a comma is not useful.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/glossary/glossary.go`, `internal/docs/glossary/tools.go`, `internal/docs/glossary/glossary_test.go` | FR-1 to FR-4, one test for each FR. | `go test ./internal/docs/glossary/` |
| 2 | `internal/plugin/assets/skills/glossary/SKILL.md` | Say that term, definition, area and source are required. | Read the skill |

## 9. Acceptance Criteria

- AC-1: Given an upsert with an empty definition and area, then the tool refuses it and names both fields.
- AC-2: Given an upsert with the Do not use word `a, b`, then the tool refuses it.
- AC-3: Given an upsert with all four fields and no `do_not_use`, then the row is added.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Rows with an empty required field after upsert | Possible | 0 | AC-1 test |

- Tests: one unit test for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| An existing row with no area cannot change until it gets an area | Low | The message names the empty field. |

## 12. Removal and Compatibility

Delete the check function. No effect on the glossary file.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which fields are required? | Equal | Term, definition, area and source. Do not use stays optional (grill Q-17, Q-22). |
| Q-2 | What does the tool do with a Do not use word that has a comma? | Equal | Refuse it (grill Q-5, Q-16). |
| Q-3 | Refuse an empty Do not use word too? | Equal | Yes (grill Q-29). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: TestUpsertRequiredFields (AC-1, AC-3) and TestUpsertRefusesBadWords (AC-2) pass. JSON-RPC on a build of the code in a scratch repository: "Give the definition and the area." (isError), "The do_not_use word "a, b" is empty or has a comma..." (isError), "Added "Task"." Commit 47f7b49.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues; govulncheck: No vulnerabilities found. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (`/dx:glossary` skill).
  - Evidence: internal/plugin/assets/skills/glossary/SKILL.md step 4 says that term, definition, area and source are required, and that each Do not use word is one list item with no comma. The glossary_upsert description says that all fields but do_not_use are required.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: Not applicable: one check function in internal/docs/glossary/glossary.go, no design diagram.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: input checks in one tool, no change of structure or file format.