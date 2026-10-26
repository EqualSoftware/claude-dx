---
id: TASK-068
title: Project data directory
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp, cli]
depends_on: [TASK-062]
related: [ADR-018, ADR-007, ADR-015]
---

# TASK-068: Project data directory

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Add the project data directory of ADR-015 with `docs.lock` and the Project file (data) with the counters, and the Project file (cache) with the project path and last use date (ADR-018).

## 2. Problem and Motivation

- The docs lock and the counters are in `<cache>/projects/<key>/`. `cache clean` removes the counters, so two worktrees can then get the same ADR number.
- The cache folder of a project has a hash as its name, so `claude-dx clean` cannot show which project it is (ADR-018).

## 3. Goals

- G-1: The docs lock and the counters are in the project data directory: `<git-common-dir>/dx/` in git, else `<cache>/projects/<key>/`.
- G-2: For each project folder in the cache, a tool can read the project path and the last use date.
- G-3: After `git init`, all processes use one docs lock, before and after the move.

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

- R-1: The project data directory is `<git-common-dir>/dx/` in a git repository, else `<cache>/projects/<key>/` (ADR-015). Exception: in git, if `<git-common-dir>/dx/` does not exist and the cache folder of the project root has data, the cache folder stays the project data directory until the move (R-6).
- R-2: `docs.lock` is in the project data directory. The docs tools use it.
- R-3: Project file (data): `project.json` in the project data directory, with `counters`. After TASK-064, `counters` has only `adr`.
- R-4: Project file (cache): `<cache>/projects/<key>/project.json`, with `path` and `last_used`. The MCP server writes it at start.
- R-5: All writes of the two files hold the docs lock of the project and are atomic (`docfile.Write`).
- R-6: At start, while the launcher holds the exclusive cache lock (ADR-015), it moves `tasks/` and the Project file (data) of the project root from the cache to `<git-common-dir>/dx/`. The old `docs.lock` stays, because no process removes a lock file.
- R-7: Remove `counters.json` with no migration, because the allocation uses `max(counter, highest number in the checkout) + 1` (ADR-007), so a lost counter gives no duplicate number in the checkout.
- R-8: A new ADR supersedes ADR-007 with all number changes of v0.1.0: the docs lock protects the counter (TASK-059), only ADRs have numbers (ADR-016), and the counter is in the Project file (data). ADR-007 is pushed, so it does not change (ADR-017).

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

- R-1, R-2: tests in a git repository with two worktrees: both use `<git-common-dir>/dx/docs.lock`. Test outside git: `<cache>/projects/<key>/docs.lock`.
- R-3 to R-5: tests of the start write and of a number allocation.
- R-6: test: a project with tasks in the cache, then `git init`, then a start with no other process moves the data; a start while another process holds the shared cache lock does not move it, and the tools still use the cache folder.
- R-7: test with an old `counters.json`.
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
