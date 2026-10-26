---
id: TASK-053
title: Check section bodies of task tools
status: Approved
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-052]
related: [ADR-013]
---

# TASK-053: Check section bodies of task tools

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |

## 1. Summary

`task_update_section` and `task_create` refuse a section body that adds headings of other sections, and `task_create` gives the same result for the same sections each time.

## 2. Problem and Motivation

Found by the code review of 2026-10-26.
- `UpdateSection` (`internal/docs/task/change.go:61`) and `fill` (`internal/docs/task/create.go:100`) put the body in with no check. A body with `## 14. Definition of Done` and checked items adds a second section 14 before the real section. `bounds` (`internal/markdown/section.go:47`) uses the first match, so the Done gate reads the false list, and `task_set_status` Done passes with real items not checked. The same works for section 13 and the Open-question gate. So the protected sections of `checkSection` do not stay protected.
- `fill` uses the `Sections` map in random order. With sections `7` and `7.1`, if `7` comes first, the `7.1` heading goes away and the create fails. If `7.1` comes first, section `7` replaces it. The same input fails or loses text.

## 3. Goals

- G-1: A section body cannot add a heading of the same or a higher level than the target section.
- G-2: `task_create` gives one result for one input.

## 4. Out of Scope

- The ADR, learning and glossary tools: they do not replace sections by number.

## 5. Context
### 5.1 Current State

- `headings` (`internal/markdown/section.go:66`) finds `##` and `###` headings outside code blocks.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-013: task flow and its gates.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `task_update_section` and `task_create` refuse a body with a heading, outside a code block, of the same or a higher level than the target section, because that heading ends the section and starts a new one. A `###` heading in a `##` section is allowed.
- FR-2: `task_create` refuses `Sections` with a section and one of its child sections, example `7` and `7.1`, because the parent body replaces the child.

### 6.2 Non-Functional

Not applicable.

## 7. Design
### 7.1 Architecture

A function in `internal/markdown/section.go` returns the level of the first heading in a body. `ReplaceSection` refuses a body with a heading of level less than or equal to the target level. `fill` checks for a parent and child pair before the loop.

### 7.2 Public API

No change of the tool schemas. The tools give new refusals.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Escape the headings in the body: rejected, a refusal tells the agent to fix its input.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/markdown/section.go`, `internal/markdown/section_test.go` | FR-1, with tests. | `go test ./internal/markdown/` |
| 2 | `internal/docs/task/create.go`, `internal/docs/task/task_test.go` | FR-2, and the tests of FR-1 for the task tools. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given a body with `## 14. Definition of Done`, when `task_update_section` runs on section 12, then it refuses and the file does not change.
- AC-2: Given a body with `### 7.1 Architecture`, when `task_update_section` runs on section 7, then it passes.
- AC-3: Given a body with `### 6.1 Functional`, when `task_update_section` runs on section 7.1, then it refuses.
- AC-4: Given `Sections` with `7` and `7.1`, when `task_create` runs, then it refuses and makes no file.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Bodies that add a protected section | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| An ADR or learning tool uses `ReplaceSection` with a body that has headings | Medium | The tests of the ADR and learning tools run in work batch 1. |

## 12. Removal and Compatibility

Revert the change.

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
