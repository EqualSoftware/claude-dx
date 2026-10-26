---
id: TASK-024
title: Full tool names in the dx:scout rules
status: Draft
priority: P1
area: [plugin]
depends_on: []
---

# TASK-024: Full tool names in the dx:scout rules

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-009, TASK-008, TASK-020 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: The rules of `scout.md` tell that the dx tools have the prefix `mcp__plugin_dx_mcp__`, or give the full names.
- FR-2: The source of a fact from a dx tool stays the short name, because it is shorter in the report.

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
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given the brief of the TASK-020 AC-2 session, when `dx:scout` runs, then no tool call fails with `No such tool available`.

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

Revert the text change.

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
