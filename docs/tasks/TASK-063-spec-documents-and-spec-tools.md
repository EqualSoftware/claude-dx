---
id: TASK-063
title: Spec documents and spec tools
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp, docs]
depends_on: [TASK-058]
related: [ADR-016, ADR-013, ADR-006]
---

# TASK-063: Spec documents and spec tools

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Add the spec documents of ADR-016 in `docs/specs/<slug>/SPEC.md` and the `spec_*` MCP tools.

## 2. Problem and Motivation

Task files in git make too many files and do not give the current behavior of a feature (ADR-016).

## 3. Goals

- G-1: A spec gives the current definition of one feature, with requirements `R-N` and their status.
- G-2: The gates of ADR-013 apply to specs.
- G-3: `spec_get` gives the commits of each requirement from the `Refs: <slug>#R-N` trailers.

## 4. Out of Scope

- Local tasks (TASK-064).
- `claude-dx resume` and the hook (TASK-065).
- Migration of docs/tasks (TASK-067).

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

- R-1: `spec_create` makes `docs/specs/<slug>/SPEC.md` from an embedded template, status `Draft`. It refuses the slug `general` and a slug that exists.
- R-2: `spec_list`, `spec_get` (all, or one section), `spec_update_section`, `spec_set_status`, `spec_req_add`, `spec_req_set`, `spec_question_add`, `spec_question_answer`, `spec_delete` (needs a reason).
- R-3: `spec_set_status` allows `Draft` -> `In Review` -> `Approved`, `In Review` -> `Draft`, `Approved` -> `In Review`. `Approved` needs no `Open` question.
- R-4: A change of Goals, Out of Scope, Requirements, Design or Acceptance Criteria of an `Approved` spec sets it to `In Review`.
- R-5: Only their own tools change Requirements status, Open Questions and Document History.
- R-6: Each write makes `docs/specs/README.md` again and holds the docs lock (ADR-015).
- R-7: `spec_get` gives, for each requirement, the commits with the trailer `Refs: <slug>#R-N`.

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

- R-1 to R-6: MCP end-to-end tests of each tool and each refusal.
- R-7: a test repository with two commits that have trailers.
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
| Q-1 | spec_get (R-7) and claude-dx resume (TASK-066) need git log and git status. The git runner is TASK-009 in v0.2.0. Do we move TASK-009 to v0.1.0, or run git directly as internal/project/project.go:57 does? | Maintainer | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
