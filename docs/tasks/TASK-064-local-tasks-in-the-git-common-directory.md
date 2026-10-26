---
id: TASK-064
title: Local tasks in the git common directory
status: Draft
priority: P1
milestone: v0.1.0
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp, docs]
depends_on: [TASK-063]
related: [ADR-016, ADR-007, ADR-015]
---

# TASK-064: Local tasks in the git common directory

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

Move the task tools to local tasks in `<git-common-dir>/dx/tasks/<spec>/<task-slug>.md` (ADR-016).

## 2. Problem and Motivation

Task files in git make too many files and number collisions between branches (ADR-016).

## 3. Goals

- G-1: Tasks are local, shared by all worktrees, with no counter.
- G-2: Many processes can change tasks of different specs at the same time.
- G-3: A `Done` task sets its requirements to `Done` in the spec.

## 4. Out of Scope

- Spec tools (TASK-063).
- `claude-dx resume` (TASK-065).
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

- R-1: Path `<git-common-dir>/dx/tasks/<spec>/<task-slug>.md`, or `general/` for a task with no spec. ID `<spec>/<task-slug>`.
- R-2: Frontmatter `id`, `title`, `status`, `priority`, `covers`, `depends_on`. Sections Plan (checklist), Notes, Definition of Done.
- R-3: Status `Todo`, `In Progress`, `Blocked`, `Done`. `Done` needs all Definition of Done items checked.
- R-4: `depends_on` refuses a task of another spec, a missing task, the same task and a cycle.
- R-5: `covers` refuses a requirement that the spec does not have. A `general` task has no `covers`.
- R-6: `task_set_status` to `Done` sets the `covers` requirements to `Done` in the spec and adds a Document History row with no task ID.
- R-7: New tool `task_step_check`. Remove `task_renumber`, `task_question_add`, `task_question_answer`, the task counter and the task index.
- R-8: Task writes hold the lock `<git-common-dir>/dx/tasks.lock` (ADR-015 `filelock`).

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

- R-1 to R-7: MCP end-to-end tests of each tool and each refusal.
- R-8: a test with two worktrees of one repository that see the same tasks.
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
| Q-1 | Outside a git repository there is no git common directory. Do the task tools refuse with a message, or use a folder in the cache (<cache>/projects/<key>/tasks)? | Maintainer | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
