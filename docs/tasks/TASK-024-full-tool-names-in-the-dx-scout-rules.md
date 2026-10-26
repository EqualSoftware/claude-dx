---
id: TASK-024
title: Full tool names in the dx:scout rules
status: Approved
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [plugin]
depends_on: []
related: [ADR-009, TASK-008, TASK-020]
---

# TASK-024: Full tool names in the dx:scout rules

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

## 1. Summary

`dx:scout` calls dx tools by the short names in its rules. Make the rules give names that the model can call.

## 2. Problem and Motivation

In the v0.1.0 QA session (Claude Code 2.1.308, binary from `7063947`), `dx:scout` (model `haiku`) called `glossary_list` two times, and Claude Code returned `No such tool available: glossary_list`. Then it called `mcp__plugin_dx_mcp__glossary_list`. 2 of 7 tool calls failed. `internal/plugin/assets/agents/scout.md:19` and `:21` give only short names. The frontmatter (`scout.md:6`) gives the full names.

## 3. Goals

- G-1: `dx:scout` makes no tool call with a name that does not exist.

## 4. Out of Scope

- The shared rules (`rules.md`) and the skills, because the main model did not fail in the session. A new task covers them if a failure occurs.

## 5. Context
### 5.1 Current State

`internal/plugin/assets/agents/scout.md:21` gives the dx tools with short names only. `:19` uses `glossary_list` as the example source. The frontmatter (`:6`) gives the full names in `tools`.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-009 (skills, scout agent and shared rules), ADR-005 (MCP tool names `mcp__plugin_dx_mcp__*`).

## 6. Requirements

### 6.1 Functional

- FR-1: The rule line that tells which dx tools to use gives the full names, example `mcp__plugin_dx_mcp__task_list`, because a subagent takes the tool name from the rules text (`docs/learnings/subagent-tool-names.md`, Q-1).
- FR-2: The source of a fact from a dx tool stays the short name, example `glossary_list`, because it is shorter in the report.

### 6.2 Non-Functional
- NFR-1: The agent prompt does not become longer by more than one line.

## 7. Design

### 7.1 Architecture
One text change in `internal/plugin/assets/agents/scout.md`.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/plugin/assets/agents/scout.md` | FR-1, FR-2. | `go test ./internal/plugin/`; session of the maintainer with the brief of TASK-020 AC-2 |

## 9. Acceptance Criteria

- AC-1: Given the brief of the TASK-020 AC-2 session, when `dx:scout` runs, then no tool call fails with `No such tool available`.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Failed tool calls of `dx:scout` for the TASK-020 AC-2 brief | 2 of 7 | 0 | Session of the maintainer |

- Tests: one manual session.
- Platforms not verified: Linux and Windows (agent text only).

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The long names make the agent prompt longer | Low | NFR-1: one line only. |

## 12. Removal and Compatibility

Revert the text change.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Full tool names or a prefix rule in the scout rules? Recommended: full names (mcp__plugin_dx_mcp__task_list ...) in the line that tells which tools to use, because the learning subagent-tool-names shows that the model takes the name from the rules text. The source of a fact stays the short name. | Equal | Full tool names in the rule line. The source of a fact stays the short name. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (Not applicable: the agent file is the doc).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.