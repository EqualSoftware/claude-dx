---
id: TASK-NNN
title: Short task title  # no brackets; quote it if it has ": " or " #"
status: Draft  # Draft / In Review / Approved / In Development / Blocked / Done / Cancelled
priority: P2  # P0 / P1 / P2 / P3
area: [<area>]  # the areas of the project, for example [api], [docs, ci]
depends_on: []  # for example [TASK-004]
---

# TASK-NNN: [Task Title]

| Field | Value |
| :--- | :--- |
| Type | `feat` / `fix` / `refactor` / `perf` / `test` / `docs` / `build` / `ci` / `chore` |
| Owner | [Name or role] |
| Reviewers | [Names or roles] |
| Platforms | [Operating systems, architectures or runtimes; or Not applicable] |
| Size | `S` (< 1 day) / `M` (1–3 days) / `L` (> 3 days: split the task) |
| Blocks | [TASK-NNN, or None] |
| Related | [ADR-NNN, docs, issues, external references] |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| YYYY-MM-DD | [Name] | Initial draft specification |

## 1. Summary
[Two or three sentences: what changes and why. A reader must understand the task from this section only.]

## 2. Problem and Motivation
[The current problem. Who has it: the developer, the end user, or the maintainer. Why it is important now. The value for the project.]

## 3. Goals
- G-1: [Measurable result.]

## 4. Out of Scope
- [Item that this task does not do, and why. Link to a follow-up task if one exists.]

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
- FR-1: [The system must ... .]

### 6.2 Non-Functional
- NFR-1: [Performance budget: response time, allocations, load time.]
- NFR-2: [Memory, size, portability, thread safety.]

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
- AC-1: Given [state], when [action], then [verifiable result].
- AC-2: Project checks pass on [platforms].

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
| Q-1 | [Question] | [Name] | [Answer, or Open] |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
