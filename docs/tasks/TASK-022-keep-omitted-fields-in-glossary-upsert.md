---
id: TASK-022
title: Keep omitted fields in glossary_upsert
status: Approved
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [ADR-006, TASK-007]
---

# TASK-022: Keep omitted fields in glossary_upsert

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

## 1. Summary

When `glossary_upsert` changes a term, an omitted optional field keeps its old value, and the term keeps its old spelling.

## 2. Problem and Motivation

In the v0.1.0 QA session (scratch repository, binary from `7063947`):
- `glossary_upsert(Store, do_not_use: [bucket])`, then `glossary_upsert(Store, definition: ...)` with no `do_not_use`, removed `bucket` from the Do not use column. After that, the tool accepted `Bucket` as a term. The synonym check stopped.
- `glossary_upsert(cache)` on the term `Cache` changed the spelling of the term to `cache`.

## 3. Goals

- G-1: A change of one field of a term does not remove the other data of the row.

## 4. Out of Scope

- A change of the spelling of a term. Use `glossary_delete`, then `glossary_upsert`.

## 5. Context
### 5.1 Current State

`Store.Upsert` (`internal/docs/glossary/glossary.go:87`) replaces the full row with the input. The `do_not_use` input (`internal/docs/glossary/tools.go:25`) is a `[]string`: an omitted list decodes as nil, and `[]` decodes as an empty list.

### 5.2 Assumptions

- A-1: The MCP SDK decodes an omitted array as nil and `[]` as an empty slice. Verify: a tool test that calls `glossary_upsert` with no `do_not_use` and with `[]`.

### 5.3 Constraints

- ADR-006 (glossary tools).
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional
- FR-1: On a change, an omitted `do_not_use` keeps the old list. An empty list removes all words (decision of the maintainer in the QA chat).
- FR-2: On a change, the term keeps its old spelling, because the match has no case difference.
- FR-3: The tool description and the `do_not_use` field description give FR-1.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
`Store.Upsert` (`internal/docs/glossary/glossary.go:87`). The tool input must tell "omitted" from "empty list": a nil slice against an empty slice.

### 7.2 Public API
```text
glossary_upsert(term, definition, area, source, do_not_use?)  # do_not_use omitted: keep; []: remove all
```

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- Replace the full row and warn about removed data: rejected by the maintainer in the QA chat.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/glossary/glossary.go`, `internal/docs/glossary/tools.go`, `internal/docs/glossary/glossary_test.go` | FR-1 to FR-3. One test for each AC. | `go test ./internal/docs/glossary/` |

## 9. Acceptance Criteria

- AC-1: Given `Store` with Do not use `bucket`, when `glossary_upsert(Store)` runs with no `do_not_use`, then the row keeps `bucket`.
- AC-2: Given the same row, when `glossary_upsert(Store, do_not_use: [])` runs, then the Do not use cell is empty.
- AC-3: Given the term `Cache`, when `glossary_upsert(cache)` runs, then the row keeps the term `Cache`.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Do not use words lost on a change | All, when `do_not_use` is omitted | 0 | AC-1 test |

- Tests: unit tests in `internal/docs/glossary/glossary_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A caller wants to remove all words and omits the field | Low | FR-3: the field description tells to give `[]`. |

## 12. Removal and Compatibility

Revert the merge in `Store.Upsert`. The file format does not change.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |

No open questions. The QA chat answered the behavior of an omitted field.

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`/dx:glossary` skill: omitted and empty `do_not_use`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.