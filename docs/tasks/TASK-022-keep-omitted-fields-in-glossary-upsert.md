---
id: TASK-022
title: Keep omitted fields in glossary_upsert
status: Draft
priority: P1
area: [docs-tools]
depends_on: []
---

# TASK-022: Keep omitted fields in glossary_upsert

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-006, TASK-007 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

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
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given `Store` with Do not use `bucket`, when `glossary_upsert(Store)` runs with no `do_not_use`, then the row keeps `bucket`.
- AC-2: Given the same row, when `glossary_upsert(Store, do_not_use: [])` runs, then the Do not use cell is empty.
- AC-3: Given the term `Cache`, when `glossary_upsert(cache)` runs, then the row keeps the term `Cache`.

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

Revert the merge in `Store.Upsert`. The file format does not change.

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
