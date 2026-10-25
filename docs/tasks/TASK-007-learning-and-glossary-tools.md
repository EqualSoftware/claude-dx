---
id: TASK-007
title: Learning and glossary tools and skills
status: Approved
priority: P1
area: [docs-tools, plugin]
depends_on: [TASK-002, TASK-004]
---

# TASK-007: Learning and glossary tools and skills

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | `M` (1–3 days) |
| Blocks | TASK-013 |
| Related | ADR-005, ADR-006, ADR-009 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |

## 1. Summary
Add the learning tools and the glossary tools of ADR-006, and the skills `/dx:learning` and `/dx:glossary`.

## 2. Problem and Motivation
Learnings need a frontmatter and a duplicate check. The glossary is a Markdown table in alphabetical order with a synonym column. The model makes errors in this mechanical work.

## 3. Goals
- G-1: `learning_create` refuses a topic that exists.
- G-2: `glossary_lookup` finds a term from its synonym in `Do not use`.
- G-3: `glossary_upsert` keeps the table in alphabetical order.

## 4. Out of Scope
- `/dx:eureka`: TASK-013.

## 5. Context
### 5.1 Current State
Reference: [equal `eukera` skill](https://github.com/is-equal/claude-equal-workflow/blob/main/plugins/equal/skills/eukera/SKILL.md) (learning format), [equal `glossary` skill](https://github.com/is-equal/claude-equal-workflow/blob/main/plugins/equal/skills/glossary/SKILL.md) and [equal glossary template](https://github.com/is-equal/claude-equal-workflow/blob/main/plugins/equal/templates/GLOSSARY.md).

### 5.2 Assumptions
- None.

### 5.3 Constraints
- ADR-006.

## 6. Requirements
### 6.1 Functional
- FR-1: `learning_list()` returns `<topic>: <description>` for each file in `docs/learnings/`.
- FR-2: `learning_get(topic)` returns the file.
- FR-3: `learning_create(topic, description, body)` makes `docs/learnings/` if necessary and writes `<topic>.md` with the frontmatter `topic` and `description`. `topic` must be kebab-case. It refuses a topic that exists and tells the model to use `learning_update`.
- FR-4: `learning_update(topic, description?, body?)` and `learning_delete(topic)`.
- FR-5: `glossary_lookup(term)` searches `Term` and `Do not use`, with no case difference. It returns the row, or `No match.` If the term is in `Do not use`, it returns `Use "<Term>", not "<term>".`
- FR-6: `glossary_upsert(term, definition, area, do_not_use, source)` adds or changes a row and keeps alphabetical order. If the file does not exist, it makes it with the header of the equal template (with `/dx:glossary`) and no rows. It refuses a term that is in `Do not use` of another row, and names that row.
- FR-7: `glossary_delete(term)` removes the row.
- FR-8: `/dx:learning` tells when a lesson is valid (verified, reusable, not obvious, not documented), and to run `learning_list` before `learning_create`. One lesson is one or two sentences.
- FR-9: `/dx:glossary` is the equal glossary skill, changed to use the tools and `dx:scout`.

### 6.2 Non-Functional
- NFR-1: Glossary operations on 500 rows take less than 20 ms.

## 7. Design
### 7.1 Architecture
`internal/docs/learning/` and `internal/docs/glossary/` register the tools. The glossary package parses and writes the Markdown table and escapes `|` in cells.

### 7.2 Public API
```text
learning_list() -> lines
learning_get(topic) -> markdown
learning_create(topic, description, body) -> message
learning_update(topic, description?, body?) -> message
learning_delete(topic) -> message
glossary_lookup(term) -> row or message
glossary_upsert(term, definition, area, do_not_use: [string], source) -> message
glossary_delete(term) -> message
```

### 7.3 Data and File Formats
```yaml
---
topic: go-embed-paths
description: go:embed paths use forward slashes on all platforms. Use when you embed files.
---
```
Glossary table: `| Term | Definition | Area | Do not use | Source |`. Empty `Do not use` is `—`.

### 7.4 User-facing Behavior
`/dx:glossary` reports the terms that it added, changed or found in conflict, or `No glossary changes`.

### 7.5 Alternatives
- See ADR-006.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/learning/` | FR-1 to FR-4. Tests. | `go test ./internal/docs/learning/` |
| 2 | `internal/docs/glossary/`, `internal/plugin/assets/templates/glossary.md` | FR-5 to FR-7. Tests for order, escape, conflict. | `go test ./internal/docs/glossary/` |
| 3 | `internal/plugin/assets/skills/{learning,glossary}/SKILL.md` | FR-8, FR-9. | Session: add a term and a lesson |

## 9. Acceptance Criteria
- AC-1: Given the term `Task` with `ticket` in `Do not use`, when `glossary_lookup("ticket")` runs, then the result is `Use "Task", not "ticket".`
- AC-2: Given `docs/learnings/x.md`, when `learning_create("x", ...)` runs, then it refuses.
- AC-3: Given no `docs/learnings/`, when `learning_delete` runs, then the result is a message and no directory is made.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tool calls to add one term | 2–3 (read, edit) | 1–2 (`glossary_lookup`, `glossary_upsert`) | Transcript |

- Tests: unit tests; one manual session.
- Platforms not verified: none.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A hand edit breaks the table | Low | The parser reports the line |

## 12. Removal and Compatibility
Delete the two packages, the template and the two skills. Files in projects stay valid Markdown.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
