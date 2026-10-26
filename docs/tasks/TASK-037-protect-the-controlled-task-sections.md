---
id: TASK-037
title: Controlled task sections and DoD tools
status: Done
priority: P1
milestone: v0.1.0
type: fix
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools, plugin]
depends_on: []
related: [ADR-013, TASK-006, TASK-034]
---

# TASK-037: Controlled task sections and DoD tools

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`task_update_section` and `task_create` do not change the Document History, the Open Questions (section 13) or the Definition of Done (section 14). New tools change the Definition of Done: `task_dod_uncheck`, `task_dod_add` and `task_dod_remove`, next to `task_dod_check`. `task_dod_check` writes the evidence on one line.

## 2. Problem and Motivation

`Store.UpdateSection` (`internal/docs/task/change.go:41`) replaces any section. A probe test replaced section 14 with `- [x] Done.` and the Document History with a bad row, with no error. Then `task_set_status(Done)` passes with no evidence. This removes the checks of `task_dod_check`, `task_question_add`, `task_question_answer` and `addHistory` (ADR-013).

`Store.CheckDoD` (`internal/docs/task/change.go:375`) adds the evidence as it is. Evidence with a new line breaks the list item: the second line is outside the `- [x]` item.

## 3. Goals

- G-1: Only the DoD tools change section 14, only the question tools change section 13, and only the status tools and the DoD tools add Document History rows.
- G-2: Each change of a checked Definition of Done item has a reason in the Document History.
- G-3: Each Definition of Done item is one line, and its evidence is one sub-item line.

## 4. Out of Scope

- A change of the status rules or of the flow (ADR-013).
- Existing task files: the rules apply at the next change.

## 5. Context
### 5.1 Current State

`UpdateSection` (`internal/docs/task/change.go:41`) calls `markdown.ReplaceSection` for any ref. `Store.fill` (`internal/docs/task/create.go`) does the same for the `sections` of `task_create`. `CheckDoD` (`internal/docs/task/change.go:375`) appends `strings.TrimSpace(evidence)`. `cell` (`internal/docs/task/task.go:253`) already makes text one line for table cells. `inReview` (`internal/docs/task/change.go:242`) refuses a leaf section with the template text, also section 14.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-013 (the flow gates are in the task tools).
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `task_update_section` refuses the Document History, section 13 and section 14, by number or by title, with no case difference. The message names the tools to use: `task_set_status` for the Document History, `task_question_add` and `task_question_answer` for section 13, `task_dod_add`, `task_dod_remove`, `task_dod_check` and `task_dod_uncheck` for section 14.
- FR-2: `task_create` applies FR-1 to its `sections`. A new task has the template Document History row, section 13 and section 14.
- FR-3: `task_dod_uncheck(id, item, reason)` sets a `- [x]` item to `- [ ]`, removes its evidence sub-items and adds a Document History row with the reason. It refuses an item that is not checked. An item checked in the old format (evidence after the text) keeps its full text.
- FR-4: `task_dod_add(id, text)` adds `- [ ] <text>` after the last item of section 14 and its sub-items, and returns its item number.
- FR-5: `task_dod_remove(id, item, reason)` removes one item, checked or not, with its sub-items, and adds a Document History row with the text of the item and the reason.
- FR-6: `reason` is required for `task_dod_uncheck` and `task_dod_remove`. All DoD tools (also `task_dod_check`) refuse a `Done` task, because `Done` is final (ADR-013).
- FR-7: `task_dod_check` writes the evidence on one indented sub-item under the item: `  - Evidence: <evidence>`. It joins the lines of the evidence with one space. `task_dod_add` joins the lines of the text with one space.
- FR-8: The `In Review` check does not refuse section 13 or section 14 with the template text, because the template content is valid.
- FR-9: The template row of section 13 is `| — | None. | — | — |`. `task_question_add` removes it when it adds the first question.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

- One function in `internal/docs/task/change.go` returns the refusal for a protected section ref, or nil. `UpdateSection` and `fill` call it before `markdown.ReplaceSection`. The ref is compared to the heading number and title, as `markdown.Section` does (`internal/markdown/section.go`).
- The DoD tools are `Store` methods next to `CheckDoD`, with the same item count rule (1 is the first `- [ ]` or `- [x]` line of section 14). They use `Store.edit` and `addHistory`.
- `inReview` skips section 14.

### 7.2 Public API

New MCP tools:
- `task_dod_uncheck`: `id`, `item` (1 is the first item), `reason`. Result: `Unchecked item <n> of <id>.`
- `task_dod_add`: `id`, `text`. Result: `Added item <n> to <id>.`
- `task_dod_remove`: `id`, `item`, `reason`. Result: `Removed item <n> of <id>.`

New refusal messages for FR-1 and FR-6. No change to the inputs of the other tools.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Accept text changes and new items in section 14 with `task_update_section`: rejected by the maintainer, because dedicated tools are safer (grill Q-24, Q-30).
- Refuse `task_dod_remove` for a checked item: rejected by the maintainer (grill Q-32).
- Two tasks (protection, DoD tools): rejected by the maintainer, one M task (grill Q-34).

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/change.go`, `internal/docs/task/create.go`, `internal/docs/task/task_test.go` | FR-1, FR-2, FR-8, with tests. | `go test ./internal/docs/task/` |
| 2 | `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go`, `internal/mcpserver/server_test.go` | FR-3 to FR-7, with tests. Add the new tools to the end-to-end tool list. | `go test ./internal/docs/task/ ./internal/mcpserver/` |
| 3 | `internal/plugin/assets/skills/task/SKILL.md`, ADR-013 | Name the tools that change sections 13, 14 and the Document History. Add the DoD tools to the Close step. | Read the skill and ADR-013 |

## 9. Acceptance Criteria

- AC-1: Given a task, when `task_update_section` replaces `Document History`, `13`, `Open Questions`, `14` or `Definition of Done`, then the tool refuses it and names the tools to use.
- AC-2: Given `task_create` with section 14 or 13 in `sections`, then the tool refuses it.
- AC-3: Given a checked item, when `task_dod_uncheck` runs with a reason, then the item is `- [ ]` with no evidence sub-item and the Document History has a row with the reason.
- AC-4: Given a task, when `task_dod_add` runs, then the new item is the last item and the result gives its number.
- AC-5: Given a checked item, when `task_dod_remove` runs with a reason, then the item and its evidence do not exist and the Document History has a row with its text and the reason.
- AC-6: Given a `Done` task, when a DoD tool runs, then the tool refuses it.
- AC-7: Given evidence with two lines, when `task_dod_check` runs, then the item has one sub-item `  - Evidence: <line 1> <line 2>`.
- AC-8: Given a task with the template sections 13 and 14 and all other sections filled, when `task_set_status(In Review)` runs, then the change is done.
- AC-9: Given a new task, when `task_question_add` runs, then section 13 has the new question and no `None.` row.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Ways to check a DoD item with no evidence | `task_update_section` on section 14 | 0 | AC-1 test |
| Changes of a checked item with no reason in the Document History | All | 0 | AC-3 and AC-5 tests |

- Tests: one unit test for each AC; the end-to-end test lists the new tools.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The model needs to fix a bad row in section 13 or the Document History | Low | The person can edit the file. The refusal names the correct tool. |
| Three more tools add to the tool list of each session | Low | Short descriptions (ADR-005: one tool for each operation). |

## 12. Removal and Compatibility

Delete the protection function, the three DoD tools and their skill lines. No effect on task files.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | How do the tools protect sections 13, 14 and the Document History? | Equal | Refuse them in `task_update_section` (grill Q-2, Q-16). |
| Q-2 | Does `task_create` apply the same rule to its `sections`? | Equal | Yes (grill Q-23). |
| Q-3 | How does a person change section 14? | Equal | New tools `task_dod_uncheck`, `task_dod_add` and `task_dod_remove` (grill Q-24, Q-30). |
| Q-4 | Rules of the DoD tools. Recommended: reason for uncheck and remove, a Document History row, no change of a `Done` task, remove only unchecked items. | Equal | Yes, but `task_dod_remove` also removes a checked item (grill Q-32). |
| Q-5 | Is the template section 14 valid for `In Review`? | Equal | Yes (grill Q-33). |
| Q-6 | One task or two? | Equal | One task, size M (grill Q-34). |
| Q-7 | How does a task with no questions pass In Review, when task_update_section cannot change section 13? Recommended: the template row is "\| — \| None. \| — \| — \|", In Review does not check section 13, and task_question_add removes that row. | Equal | Yes, as recommended (grill Q-35). |
| Q-8 | Which format does task_dod_check use for the evidence, so that task_dod_uncheck can remove it? Recommended: "- [x] <text> Evidence: <evidence>" on one line. | Equal | A sub-item: the evidence goes on an indented line "  - Evidence: <evidence>" under the item (grill Q-36). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence. Commit d121c6e. internal/docs/task/task_test.go: TestProtectedSections (AC-1, AC-2), TestDoDCheckUncheck (AC-3, AC-7), TestDoDAddRemove (AC-4, AC-5), TestDoDDoneTask (AC-6), TestInReviewWhenComplete (AC-8, template sections 13 and 14), TestQuestions (AC-9, no None. row). internal/mcpserver/server_test.go lists the three new tools. go test ./... passes.
- [x] Project checks (build, tests, lint) pass. Local on macOS at d121c6e: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (`/dx:task` skill, ADR-013). Commit d121c6e: internal/plugin/assets/skills/task/SKILL.md (Flow step 6, Create step 3, Refine: the tools of each protected section, Change the status: Done). ADR-013 Decision has the protected sections.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. No diagram is necessary: the rules are a list in ADR-013 Decision and in the /dx:task skill.
- [x] ADRs are recorded for architectural decisions (ADR-013 gets the DoD tools). ADR-013 Decision records the protected sections and the Definition of Done tools.