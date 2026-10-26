---
id: TASK-068
title: Project file in the cache
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp, cli]
depends_on: [TASK-064]
related: [ADR-018, ADR-007, ADR-015]
---

# TASK-068: Project file in the cache

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Replace `counters.json` with `<cache>/projects/<key>/project.json`, which also has the project path and the last use date (ADR-018).

## 2. Problem and Motivation

The cache folder of a project has a hash as its name, so `claude-dx clean` cannot show which project it is (ADR-018).

## 3. Goals

- G-1: For each project folder in the cache, a tool can read the project path and the last use date.

## 4. Out of Scope

- The clean command (TASK-069).

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

- R-1: `project.json` has `path`, `last_used` and `counters` (the content of `counters.json` now).
- R-2: The MCP server writes `path` and `last_used` at start.
- R-3: All writes hold the docs lock and are atomic (`docfile.Write`), because the MCP server start and the number allocation both write the file.
- R-4: Remove `counters.json` with no migration, because the allocation uses `max(counter, highest number in the checkout) + 1` (ADR-007), so a lost counter gives no duplicate number in the checkout.

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

- R-1 to R-3: tests of the start write and of a number allocation in one file.
- R-4: test with an old `counters.json`.
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
