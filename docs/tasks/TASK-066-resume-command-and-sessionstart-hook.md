---
id: TASK-066
title: Resume command and SessionStart hook
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli, plugin]
depends_on: [TASK-064]
related: [ADR-016]
---

# TASK-066: Resume command and SessionStart hook

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Add the command `claude-dx resume` and a `SessionStart` hook of the plugin that runs it (ADR-016).

## 2. Problem and Motivation

HANDOFF.md is written by hand and the model must remember to update it (ADR-016).

## 3. Goals

- G-1: At each session start, `/clear` and compact, the model gets the work state with no tool call.
- G-2: The output is short: 40 lines or less.

## 4. Out of Scope

- A journal of events (ADR-016 rejects it).
- Spec and task tools (TASK-063, TASK-064).

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

- R-1: `claude-dx resume` prints: branch, commits ahead of upstream, changed files; tasks `In Progress` by modification time, with the first unchecked Plan step and the Notes; next `Todo` tasks with all dependencies `Done`; `Blocked` tasks; unpushed commits with their `Refs:` trailers.
- R-2: It uses no network and no MCP server.
- R-3: The plugin has a `SessionStart` hook with the matchers `startup`, `resume`, `clear` and `compact` that runs it.
- R-4: With no tasks and a clean tree, it prints one line.

## 7. Design
### 7.1 Architecture
[Affected modules and targets. Data flow. Ownership and lifetime of memory and resources. Threads. A Mermaid diagram if the task has a call flow, an ownership or lifetime rule, a state change, or data that moves between modules.]

### 7.2 Public API
[New or changed public functions, types, commands or endpoints. Give the exact signatures. State the breaking changes.]

```text
module_verb(argument) -> result
```

### 7.3 Data and File Formats
[New or changed files, formats, schemas or configuration keys. Give examples. State how existing data migrates.]

### 7.4 User-facing Behavior
[Behavior that the end user sees: input, output, feedback, timing, visual specs. Write "Not applicable" for tasks with no user-facing change.]

### 7.5 Alternatives
- [Alternative]: [Reason rejected.] [Record an ADR with `/dx:adr` if the decision is architectural.]

## 8. Implementation Plan
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- R-1, R-4: golden tests of the output in a test repository.
- R-3: plugin test that the hook file is in the extracted plugin.
- Tests pass on macOS, Linux and Windows.

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
[How to remove the feature: files and symbols to delete. Effect on existing users, data and saved state. Deprecation steps if any.]

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
