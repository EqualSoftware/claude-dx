---
id: TASK-070
title: Ask choices found during work
status: Done
priority: P1
milestone: v0.1.0
type: feat
size: S
owner: Equal
platforms: [Not applicable]
area: [plugin]
depends_on: []
related: [ADR-009, ADR-013, TASK-061, TASK-059]
---

# TASK-070: Ask choices found during work

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | The user approved the three text changes in the conversation. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

The dx rules and the grill and task skills tell the model to ask the user about each choice that a task does not decide, also during work, and to stop on a conflict with an Accepted ADR.

## 2. Problem and Motivation

During TASK-061 the model made two choices with no question: the `doctor` runs with no cache lock when the cache is not writable, and the `cache clean` success text changed. The grill check did not find them, because FR-4 and FR-6 did not define the success text and the lock failure of `doctor`. TASK-059 conflicted with ADR-017, and no rule told the model to stop. The rules (`internal/plugin/assets/rules.md`) and the grill skill (`internal/plugin/assets/skills/grill/SKILL.md`, sections 2 to 5) cover only the refinement.

## 3. Goals

- G-1: During work, the model asks the user before each choice that the task or ADR does not decide.
- G-2: The grill check finds missing error paths, fallbacks, user-facing texts, affected commands and ADR conflicts before approval.
- G-3: The model stops when a task conflicts with an Accepted ADR or another task.

## 4. Out of Scope

- Code changes. Only the plugin text files change.
- The skill changes of the spec model (TASK-067).

## 5. Context
### 5.1 Current State

- `internal/plugin/assets/rules.md`: 6 rules, none for work after approval.
- `internal/plugin/assets/skills/grill/SKILL.md` section 4: generic examples of hidden decisions.
- `internal/plugin/assets/skills/task/SKILL.md:20`: step 5 says only that a new finding is a new task.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-009: skills and shared rules are plugin files.
- Text in ASD-STE100.

## 6. Requirements
### 6.1 Functional

- FR-1: `rules.md` has two new rules: (a) During work, if the task or ADR does not decide a choice (user-facing text, error or fallback behavior, a name, a format, a change to another command), stop. Ask the user, with a recommendation, before you write it. Do not report a choice after you made it. (b) If a task conflicts with an Accepted ADR or with another task, stop and ask the user.
- FR-2: Grill section 4 has a list to check in each requirement: each error path and each fallback, each user-facing text (messages, help, docs), each existing command or tool that the change affects, each conflict with an Accepted ADR (`adr_list`, `adr_get`).
- FR-3: Task skill step 5: "Develop: `In Development`. A choice that the task does not decide is a question to the user before you write it (rules). A new finding during work is a new task, not a change of the scope."

### 6.2 Non-Functional

- NFR-1: Each new rule is one line or a short list, because the rules load in each session.

## 7. Design
### 7.1 Architecture

Text changes in three plugin files. No code change.

### 7.2 Public API

Not applicable.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

The model asks more questions during work. It does not report a choice after it made it.

### 7.5 Alternatives

- Put the changes in TASK-067: rejected, they would apply only after the migration.
- A memory of one user: rejected, it applies only to that user and project.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/plugin/assets/rules.md`, `internal/plugin/assets/skills/grill/SKILL.md`, `internal/plugin/assets/skills/task/SKILL.md` | FR-1 to FR-3. | `go test ./internal/plugin/`, read the three files |

## 9. Acceptance Criteria

- AC-1: Given `rules.md`, then it has the two rules of FR-1.
- AC-2: Given grill section 4, then it has the list of FR-2.
- AC-3: Given task skill step 5, then it has the text of FR-3.
- AC-4: Given TASK-067, then it depends on this task, so the skill rewrite keeps the rules.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Rules for choices during work | 0 | 2 | AC-1 |

- Tests: `go test ./internal/plugin/`.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| More questions slow the work | Low | The rule applies only to choices that the task does not decide. |

## 12. Removal and Compatibility

Revert the three text changes.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: AC-1 internal/plugin/assets/rules.md:8-9; AC-2 internal/plugin/assets/skills/grill/SKILL.md:50-54; AC-3 internal/plugin/assets/skills/task/SKILL.md:20; AC-4 TASK-067 depends_on has TASK-070.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues.
- [x] Docs are updated (area docs, `CLAUDE.md`).
  - Evidence: The changed files are the docs of the plugin area. CLAUDE.md needs no change.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: No diagram: the task changes only text (section 7.1).
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: ADR-009 already covers skills and shared rules.