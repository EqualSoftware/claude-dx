---
id: TASK-033
title: Task fields in the frontmatter
status: Draft
priority: P1
area: [docs-tools, plugin]
depends_on: [TASK-029]
---

# TASK-033: Task fields in the frontmatter

| Field | Value |
| :--- | :--- |
| Type | refactor |
| Owner | Equal |
| Reviewers | Equal |
| Platforms | macOS, Linux, Windows |
| Size | M |
| Blocks | None |
| Related | ADR-013, ADR-006, TASK-006 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

Move all task fields from the Field/Value table under the title into the frontmatter, remove `blocks` and `reviewers`, and set `owner` to the git `user.name` of the person who creates the task (ADR-013, step 1).

## 2. Problem and Motivation

A task has two field formats: the frontmatter (`id`, `title`, `status`, `priority`, `area`, `depends_on`) and a Field/Value table (`Type`, `Owner`, `Reviewers`, `Platforms`, `Size`, `Blocks`, `Related`). `internal/docs/task/create.go:27` writes the table, and `internal/docs/task/change.go:20` reads `Size` with a second parser. `task_list` cannot filter table values. After the QA cleanup commit `d9ce536`, `Blocks` was wrong in most tasks, because it is the inverse of `depends_on` and a person must keep it correct. `Owner` and `Reviewers` are always `Maintainer`, so they do not tell who created the task.

## 3. Goals

- G-1: All task fields are in the frontmatter, with one parser.
- G-2: Each task records the person who created it.

## 4. Out of Scope

- The `milestone` field: TASK-030.
- The Document History table and the other body tables.

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
- FR-1: The frontmatter has these fields, in this order: `id`, `title`, `status`, `priority`, `type`, `size`, `owner`, `platforms`, `area`, `depends_on`, `related`. `related` is a flow list, as `area`.
- FR-2: The template and new tasks have no Field/Value table.
- FR-3: `blocks` and `reviewers` do not exist, because `depends_on` has the data of `blocks`, and the project has one reviewer.
- FR-4: `task_create` sets `owner` to the git `user.name` of the project, as the Document History author (`internal/docs/task/task.go:44`). With no `user.name`, the value is `Agent`, as for Document History.
- FR-5: `task_update_fields` changes `type`, `size`, `owner`, `platforms` and `related` in the frontmatter. The input `fields` object has no `blocks` and no `reviewers`.
- FR-6: The `Size` check of `task_set_status` reads the frontmatter.
- FR-7: All existing tasks change to the new format in one commit. `owner` is the author of the commit that added the file (`git log --follow --diff-filter=A --format=%an`). A one-time script does the change. The script is not added to the repository, because it runs one time.
- FR-8: The `/dx:task` skill and ADR-006 give the new fields.
- FR-9: `task_list` has one optional filter for each frontmatter field (Q-1). A text field matches with no case difference. A list field (`area`, `depends_on`, `related`) matches when the list has the value. Example: `task_list(depends_on: TASK-029)` gives the tasks that wait for TASK-029, so `blocks` is not necessary. One match function reads the frontmatter map, so a new field needs no new filter code.

### 6.2 Non-Functional
- NFR-1: One frontmatter parse for each task read. No table parse for fields.

## 7. Design

### 7.1 Architecture
Frontmatter writes with `markdown.SetField` in `internal/docs/task/create.go` and `internal/docs/task/change.go`. Remove the table code at `internal/docs/task/create.go:27` and `:93`, and the regex at `internal/docs/task/change.go:20`.

### 7.2 Public API
```text
task_create(title, priority, area, depends_on?, fields?{type, size, platforms, related}, sections?)
task_update_fields(id, ..., fields?{type, size, owner, platforms, related})
```

### 7.3 Data and File Formats
```yaml
---
id: TASK-033
title: Task fields in the frontmatter
status: Draft
priority: P1
type: refactor
size: M
owner: Equal
platforms: macOS, Linux, Windows
area: [docs-tools, plugin]
depends_on: [TASK-029]
related: [ADR-013, ADR-006, TASK-006]
---
```

### 7.4 User-facing Behavior
GitHub shows the frontmatter as a table at the top of the file.

### 7.5 Alternatives
- Keep the Field/Value table: rejected (ADR-013).
- Calculate `blocks`: rejected by the maintainer.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/template.md`, `internal/docs/task/create.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-6. Tests. | `go test ./internal/docs/task/` |
| 2 | `docs/tasks/*.md` | FR-7 with a one-time script. | `task_list(status: all)` lists all tasks; no file has `\| Field \| Value \|` |
| 3 | `internal/plugin/assets/skills/task/SKILL.md`, ADR-006 | FR-8. | Read the diff |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs with git `user.name` `Equal`, then the frontmatter has `owner: Equal` and the file has no Field/Value table.
- AC-2: Given a task with `size: L`, when `task_set_status(In Review)` runs, then the tool refuses it.
- AC-3: After work batch 2, no task file has a Field/Value table, and each task has an `owner` from git history.
- AC-4: Given TASK-030 depends on TASK-033, when `task_list(depends_on: TASK-033)` runs, then the result has TASK-030.

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

Revert the commits. The migration commit changes all task files, so revert it with the code.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which new task_list filters does the frontmatter make useful? Options: type, size, owner. Recommended: owner and type, in TASK-033, because each is one equality check on a frontmatter field. Size is only S or M. | Equal | All frontmatter fields are filters (decision of the maintainer). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
