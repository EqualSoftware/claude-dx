---
id: TASK-020
title: The glossary_list tool
status: In Development
priority: P1
area: [docs-tools, plugin]
depends_on: [TASK-007, TASK-008]
---

# TASK-020: The glossary_list tool

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | `S` (< 1 day) |
| Blocks | None |
| Related | ADR-005, ADR-006, TASK-007 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial specification from the v0.1.0 QA of the maintainer |
| 2026-10-25 | Equal | Q-1 to Q-3 answered in the chat. Approved for v0.1.0. Work started. The maintainer does the real session test. |

## 1. Summary
Add `glossary_list(area?)`. It returns one row for each term of `docs/GLOSSARY.md`. `dx:scout` and `/dx:glossary` use it in place of a read of the file.

## 2. Problem and Motivation
In a QA session, the request "which terms are in docs/GLOSSARY.md" made `dx:scout` read the full file, because `glossary_lookup` finds only one term. With a list tool, the model does not read the header text and gets the same row format as `glossary_lookup`.

## 3. Goals
- G-1: A question about all terms, or the terms of one area, uses one `glossary_list` call and no file read.

## 4. Out of Scope
- A text search in definitions: `glossary_lookup` and Grep cover it.

## 5. Context
### 5.1 Current State
`internal/docs/glossary/tools.go` registers `glossary_lookup`, `glossary_upsert` and `glossary_delete`. `Row.line` (`internal/docs/glossary/glossary.go`) makes the row text of `glossary_lookup`.

### 5.2 Assumptions
- None.

### 5.3 Constraints
- ADR-005: compact results; `*_list` tools are allowed with no prompt. ADR-006.

## 6. Requirements
### 6.1 Functional
- FR-1: `glossary_list(area?)` returns the Markdown row of each term, in file order, one on each line (Q-1).
- FR-2: A non-empty `area` keeps only the rows of that area, with no case difference (Q-2).
- FR-3: With no rows: `No terms in docs/GLOSSARY.md.`, or `No terms with the area "<area>".` With no file: `No glossary. docs/GLOSSARY.md does not exist.`
- FR-4: `dx:scout` has `glossary_list` in its tools. `/dx:glossary` and `dx:scout` use it for many terms and do not read the file. ADR-006 lists the tool (Q-3).

### 6.2 Non-Functional
- NFR-1: Not applicable. The table parse of TASK-007 NFR-1 applies.

## 7. Design
### 7.1 Architecture
`Store.List(area)` in `internal/docs/glossary/glossary.go`, `listText` in `tools.go`.

### 7.2 Public API
```text
glossary_list(area?) -> rows or message
```

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- Term and area only, or term, area and definition: rejected (Q-1). The full row has the same format as `glossary_lookup`.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/glossary/`, `agents/scout.md`, `skills/glossary/SKILL.md`, ADR-006 | FR-1 to FR-4. Tests. | `go test ./internal/docs/glossary/`; session of the maintainer |

## 9. Acceptance Criteria
- AC-1: Given a glossary with terms, when `glossary_list` runs with no area, then the result has one row for each term.
- AC-2: Given a fact search about all terms, when `dx:scout` runs, then it calls `glossary_list` and does not read `docs/GLOSSARY.md`.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tool calls of `dx:scout` to list all terms | 3 (read of the file) | 1 | Session of the maintainer |

- Tests: `TestToolsFlow`, `TestListArea`; one manual session by the maintainer.
- Platforms not verified: none.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A large glossary gives a large result | Low | The `area` filter |

## 12. Removal and Compatibility
Delete `List`, `listText`, the tool registration and the skill and agent lines.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Result of each term. Recommended: the full Markdown row, as `glossary_lookup`. | Maintainer | Accepted. |
| Q-2 | An optional `area` filter. Recommended: yes, as `task_list`. | Maintainer | Accepted. |
| Q-3 | Use in `dx:scout`, `/dx:glossary` and ADR-006. Recommended: yes. | Maintainer | Accepted. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
