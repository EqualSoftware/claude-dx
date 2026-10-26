---
id: TASK-029
title: Task frontmatter with no template comments
status: Draft
priority: P1
area: [docs-tools]
depends_on: []
---

# TASK-029: Task frontmatter with no template comments

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-013, TASK-006 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

`task_create` copies the comments of the template frontmatter into each new task. Remove the comments from the template (ADR-013, step 1).

## 2. Problem and Motivation

`internal/docs/task/template.md:3` to `:7` has a `# ...` comment on each frontmatter field, and `markdown.SetField` keeps the comment of the line (`internal/markdown/frontmatter.go:50`). In the v0.1.0 QA session, TASK-021 to TASK-028 got these comments, example `title: ...  # no brackets; quote it if it has ": " or " #"`. The comments make the rows long, and a person or a script that reads the title gets the wrong text.

## 3. Goals

- G-1: A new task file has a frontmatter with no comments.

## 4. Out of Scope

- A change of `SetField`. Learning files also use it.
- The cleanup of existing files: the QA cleanup commit did it.

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
- FR-1: The template frontmatter has no comments, because the tools, not persons, write the frontmatter (ADR-006).
- FR-2: The rules of the removed comments that the tools do not check go to the tool field descriptions.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
A change of `internal/docs/task/template.md` only.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
The frontmatter of new tasks has no comments.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- Remove the comments in `SetField`: rejected, because it changes all frontmatter writes for one template problem.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/template.md`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1, FR-2. A test that reads the frontmatter of a new task. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs, then no line of the new frontmatter has ` # `.

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

Add the comments again. Existing files do not change.

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
