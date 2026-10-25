---
id: TASK-005
title: ADR tools and the /dx:adr skill
status: Done
priority: P1
area: [docs-tools, plugin]
depends_on: [TASK-002, TASK-004]
---

# TASK-005: ADR tools and the /dx:adr skill

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | `S` (< 1 day) |
| Blocks | TASK-013 |
| Related | ADR-005, ADR-006, ADR-007, ADR-009 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |
| 2026-10-25 | Equal | Work started. |
| 2026-10-25 | Claude | Template is in `internal/docs/adr/` (only the server uses it). FR-7 allows the whole dx server (`mcp__plugin_dx_dx`); later git tools that change state go in an `ask` list. Result helpers moved to `internal/result` (import cycle). |
| 2026-10-25 | Equal | Done. All acceptance criteria pass. |

## 1. Summary
Add the ADR tools of ADR-006 and the `/dx:adr` skill. The skill is the equal `adr` skill, changed to use the tools and the dx names.

## 2. Problem and Motivation
To record an ADR, the model now reads the template, lists files, finds the number and writes the file. One tool call must do this.

## 3. Goals
- G-1: `adr_create` makes a complete ADR file in one call.
- G-2: `adr_supersede` changes the old and the new ADR in one call.
- G-3: With no `docs/decisions/`, `adr_list` returns a message and makes no directory.

## 4. Out of Scope
- Delete of ADRs: ADR-006.
- Configurable paths or templates: ADR-006.

## 5. Context
### 5.1 Current State
Reference: [equal `adr` skill and template](https://github.com/is-equal/claude-equal-workflow/tree/main/plugins/equal/skills/adr). TASK-004 gives the base packages.

### 5.2 Assumptions
- None.

### 5.3 Constraints
- ADR-005, ADR-006, ADR-007.

## 6. Requirements
### 6.1 Functional
- FR-1: `adr_list()` returns one line for each ADR: `ADR-NNN <status> <title>`, sorted by number, and a warning line for each duplicate number.
- FR-2: `adr_get(id, section?)` returns the full file, or one section (`Context`, `Decision`, `Alternatives`, `Consequences`, `Evidence`).
- FR-3: `adr_create(title, context, decision, alternatives, consequences, evidence, status?, owners?)` allocates the number (ADR-007), makes `docs/decisions/` if necessary, fills the template, sets the date to today and returns the path. `status` is `Proposed` (default) or `Accepted`. `owners` default is `Maintainer`. The file name is `ADR-NNN-<kebab of title>.md`, with a maximum of 60 characters after the number.
- FR-4: `adr_update(id, status?, section?, body?)` changes the status or replaces one section.
- FR-5: `adr_supersede(old_id, new_id)` sets `Status: Superseded by ADR-<new>` in the old ADR and adds `Replaces ADR-<old>.` at the start of Context of the new ADR.
- FR-7: The launcher adds `--settings` that allows all dx tools (`mcp__plugin_dx_dx`, ADR-005). Git tools that change state will be in an `ask` list, because `ask` wins over `allow`.
- FR-6: The skill `/dx:adr` has the "When to use", "Procedure" and "Report" parts of the equal skill. Procedure: run `/dx:grill` unless the user already confirmed all answers; run `adr_list` to find a related ADR; update it with `adr_update`, or make a new ADR with `adr_create`; set `Accepted` only when the user approved; use `adr_supersede` for a replacement.

### 6.2 Non-Functional
- NFR-1: Each tool result is less than 300 characters, except `adr_get` and `adr_list`.

## 7. Design
### 7.1 Architecture
`internal/docs/adr/` registers the tools on the server and embeds the template (`internal/docs/adr/template.md`).

### 7.2 Public API
```text
adr_list() -> lines
adr_get(id: string, section?: string) -> markdown
adr_create(title, context, decision: string, alternatives: [{option, reason}], consequences: {positive, negative, follow_up}, evidence: string, status?: "Proposed"|"Accepted", owners?: string) -> "Created docs/decisions/ADR-012-<title>.md (Proposed)."
adr_update(id, status?: "Proposed"|"Accepted", section?: string, body?: string) -> message
adr_supersede(old_id, new_id) -> message
```

### 7.3 Data and File Formats
The equal ADR template, with no change.

### 7.4 User-facing Behavior
`/dx:adr` reports the ADR path, its status and one sentence that states the decision.

### 7.5 Alternatives
- See ADR-006.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/adr/`, `internal/plugin/assets/templates/adr.md` | FR-1 to FR-5. Golden file tests. Test with no directory. | `go test ./internal/docs/adr/` |
| 2 | `internal/plugin/assets/skills/adr/SKILL.md` | FR-6. | Session: `/dx:adr` records an ADR with one `adr_create` call |

## 9. Acceptance Criteria
- AC-1: Given no `docs/decisions/`, when `adr_list` runs, then the result is `No ADRs. docs/decisions does not exist.` and no directory is made.
- AC-2: Given ADR-001 to ADR-003, when `adr_create` runs, then it makes ADR-004 with the date of today.
- AC-3: Given ADR-002 and ADR-004, when `adr_supersede(ADR-002, ADR-004)` runs, then ADR-002 has `Superseded by ADR-004` and ADR-004 Context starts with `Replaces ADR-002.`

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tool calls to record one ADR (after grill) | ~6 | 2 (`adr_list`, `adr_create`) | Count tool calls in a session transcript |

- Tests: golden file tests; one manual session.
- Platforms not verified: none.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Long tool arguments for big ADRs | Low | Sections are short by rule (equal skill) |

## 12. Removal and Compatibility
Delete `internal/docs/adr/`, the template and the skill. ADR files in projects stay valid Markdown.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

## 14. Definition of Done
- [x] All acceptance criteria pass, with evidence. AC-1: `TestMissingDir` and `TestToolsFlow` (`adr_list` returns `No ADRs. docs/decisions does not exist.` and makes no directory). AC-2: `TestCreateNextNumber`. AC-3: `TestSupersede`. CI run 40231357876 on macOS, Linux and Windows. Real session (Claude Code 2.1.308, Haiku): `/dx:adr` made `docs/decisions/ADR-001-store-the-local-cache-in-sqlite.md` at the git root from a subdirectory, with no permission prompt. Metric: 2 dx tool calls (`adr_list`, `adr_create`); target 2.
- [x] Project checks (build, tests, lint) pass. CI run 40231357876; lint with GOOS=windows also passes.
- [x] Docs are updated (README tool list). Not applicable yet: the README lists features by version in Roadmap; a tool list comes with the v0.1 release notes.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no diagram.
- [x] ADRs are recorded for architectural decisions. No new decision; ADR-005 and ADR-006 describe the code.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new lesson: the behavior of the SDK for missing required fields is covered by `TestToolSchemaRequiresFields`.
