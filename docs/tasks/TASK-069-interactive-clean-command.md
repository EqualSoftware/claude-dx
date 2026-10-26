---
id: TASK-069
title: Interactive clean command
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli, cache]
depends_on: [TASK-061, TASK-068]
related: [ADR-018, ADR-016, ADR-015]
---

# TASK-069: Interactive clean command

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Replace `cache clean` with the interactive command `claude-dx clean` of ADR-018.

## 2. Problem and Motivation

`cache clean` removes all of the cache (`internal/cache/cache.go:24`), and that can delete local tasks of the user (ADR-016, ADR-018).

## 3. Goals

- G-1: The user selects each item to delete and confirms. Nothing else is deleted.
- G-2: The user sees all known projects, their data and their local tasks.
- G-3: A project can be cleaned while another project has a session.

## 4. Out of Scope

- Flags that delete with no interactive selection (ADR-018).
- The project file (TASK-068).

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

- R-1: `claude-dx clean` shows a `charmbracelet/huh` multi-select: Global (old plugin versions, each LSP server) and Projects (for each `project.json`: cache data and local tasks by spec and status). A project path that does not exist is marked missing.
- R-2: After the selection, the command shows the items and their size and asks for a confirmation.
- R-3: Global items need the exclusive cache lock, with no wait. If a session holds the shared lock, they are disabled with the reason.
- R-4: Project items hold the docs lock or the task lock of that project.
- R-5: Lock files are never shown and never deleted (ADR-015).
- R-6: `claude-dx clean --list` prints the tree with sizes and status and deletes nothing.
- R-7: With no interactive terminal, `claude-dx clean` refuses and gives `--list`.
- R-8: Remove `cache clean`.

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

- R-1 to R-6: tests of the item tree, the locks and the delete, with the selection given by a test function, not a terminal.
- R-7: test with no terminal.
- Manual test of the selection in Windows Terminal, PowerShell and Git Bash, and on macOS and Linux.
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
