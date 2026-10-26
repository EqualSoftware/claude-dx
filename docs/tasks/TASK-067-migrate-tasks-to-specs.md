---
id: TASK-067
title: Migrate tasks to specs
status: Draft
priority: P1
milestone: v0.1.0
type: docs
size: M
owner: Equal
platforms: [Not applicable]
area: [docs, plugin]
depends_on: [TASK-064, TASK-065, TASK-066, TASK-070]
related: [ADR-016, ADR-017, ADR-013]
---

# TASK-067: Migrate tasks to specs

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Move the content of docs/tasks to specs, delete docs/tasks and HANDOFF.md, and change the skills, rules and ADR references (ADR-016, ADR-017).

## 2. Problem and Motivation

After TASK-063 to TASK-066 the project still has 60 or more task files in git and text that gives TASK-NNN IDs.

## 3. Goals

- G-1: No `docs/tasks/` folder and no HANDOFF.md.
- G-2: Each feature of the project has a spec with its requirements and their status.
- G-3: No ADR, learning, glossary row, skill or rule gives a TASK-NNN ID.

## 4. Out of Scope

- New features. The open tasks of v0.2.0 and later become `Draft` specs, and the user plans the milestones again after this task.

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

- R-1: Group the `Done` tasks into specs by area. Their requirements are `Done`.
- R-2: The open tasks of v0.2.0 and later become `Draft` specs with their milestone.
- R-3: Change TASK-NNN references in ADRs, learnings and the glossary to `<slug>#R-N` (the one exception of ADR-017).
- R-4: Change the task, adr and grill skills, `rules.md`, the scout agent and `CLAUDE.md` to the spec and local task model.
- R-5: Delete `docs/tasks/` and HANDOFF.md, and remove `/HANDOFF.md` from `.gitignore`.
- R-6: Set ADR-016 and ADR-017 to `Accepted` with the user approval, then supersede ADR-013 with ADR-016.

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

- G-1: `ls docs/tasks HANDOFF.md` finds nothing.
- G-3: `grep -rn TASK-[0-9] docs internal/plugin CLAUDE.md` finds nothing.
- The user approves the spec list.

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
| Q-1 | Which specs do the 62 tasks become? Proposal: launcher-and-cli, cache-and-update, mcp-server, document-tools, git-tools, lsp-servers, release. | Maintainer | Yes: launcher-and-cli, cache-and-update, mcp-server, document-tools, git-tools, lsp-servers, release. |
| Q-2 | The glossary rows Task and the task terms describe the old model. Do we change them in this task, with the new rows Spec, Requirement and Resume? | Maintainer | Yes. Change the task terms and add Spec, Requirement and Resume in this task. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
