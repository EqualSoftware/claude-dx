---
id: TASK-013
title: The /dx:commit and /dx:eureka skills
status: Approved
priority: P1
milestone: v0.2.0
type: feat
size: S
owner: Equal
platforms: [Not applicable]
area: [plugin]
depends_on: [TASK-010, TASK-011, TASK-026]
related: [ADR-008, ADR-009, ADR-013, TASK-005, TASK-007]
---

# TASK-013: The /dx:commit and /dx:eureka skills

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |
| 2026-10-25 | Equal | Scope change from TASK-026: work batch 2 and AC-3 add the /dx:eureka Definition of Done item again (ADR-013, step 3). The maintainer must approve again. |
| 2026-10-25 | Equal | The maintainer approved the scope change from TASK-026 (work batch 2, AC-3). |

## 1. Summary
Add `/dx:commit`, which splits changes into Conventional Commits with the git tools, and `/dx:eureka`, which finds verified lessons and records them as an ADR, a learning or a glossary term.

## 2. Problem and Motivation
The git tools make one commit. The model needs a procedure to split work into good commits. Lessons from work are lost if no procedure records them.

## 3. Goals
- G-1: `/dx:commit` makes commits only with the git tools.
- G-2: `/dx:eureka` records only to ADRs, learnings or the glossary, and reports `No new lessons` when no lesson passes.

## 4. Out of Scope
- Lessons to `CLAUDE.md`, code comments or skills: The design interview (2026-10-25) limits Eureka to ADR, learnings and glossary.

## 5. Context
### 5.1 Current State
Reference: [equal `eukera` skill](https://github.com/is-equal/claude-equal-workflow/blob/main/plugins/equal/skills/eukera/SKILL.md) and [equal `commit` skill](https://github.com/is-equal/claude-equal-workflow/tree/main/plugins/equal/skills/commit).

### 5.2 Assumptions
- None.

### 5.3 Constraints
- ADR-009.

## 6. Requirements
### 6.1 Functional
- FR-1: `/dx:commit` procedure: `git_status`; `git_diff` (stat first, then files as necessary); group the changes into logical commits; for each group, `git_commit` with `paths`; add `co_authors` when the project or session uses them; do not push unless the user asks.
- FR-2: `/dx:eureka` procedure: get evidence with `git_diff`, `git_log` and the conversation; keep only lessons that are verified, reusable, not obvious and not documented (search with `learning_list`, `adr_list`, `glossary_lookup`, and `dx:scout` for more than 3 files); record with `/dx:adr`, `learning_create` / `learning_update`, or `/dx:glossary`; one lesson is one or two sentences.

### 6.2 Non-Functional
- Not applicable.

## 7. Design
### 7.1 Architecture
Files: `internal/plugin/assets/skills/commit/SKILL.md`, `internal/plugin/assets/skills/eureka/SKILL.md`.

### 7.2 Public API
`/dx:commit`, `/dx:eureka`.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
`/dx:commit` reports each commit: hash and header. `/dx:eureka` reports each lesson and its place, or `No new lessons`.

### 7.5 Alternatives
- Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | the two files of 7.1 | FR-1, FR-2. | Session: commit mixed changes; run eureka after a task |
| 2 | `internal/docs/task/template.md`, `internal/docs/task/task_test.go` | Add the Definition of Done item `/dx:eureka is done` again (TASK-026 removed it until this skill exists). | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given changes in two unrelated areas, when `/dx:commit` runs, then it makes two commits with valid headers.
- AC-2: Given work with no new lesson, when `/dx:eureka` runs, then it reports `No new lessons` and changes no file.
- AC-3: When `task_create` runs, then the Definition of Done of the new task has the `/dx:eureka` item.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Bash git calls in a `/dx:commit` run | Not measured | 0 | Transcript |

- Tests: manual sessions.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Eureka records weak lessons | Medium | The four filters are mandatory in the skill |

## 12. Removal and Compatibility
Delete the two skill files.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
