---
id: TASK-033
title: Task fields in the frontmatter
status: Done
priority: P1
type: refactor
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools, plugin]
depends_on: []
related: [ADR-013, ADR-006, TASK-006, TASK-029]
---

# TASK-033: Task fields in the frontmatter

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding, also for the field order, the Agent default, the filter match rules and the one-time migration script. |
| 2026-10-25 | Equal | Approved by the maintainer. |
| 2026-10-25 | Equal | Status: In Development. |
| 2026-10-25 | Equal | All acceptance criteria pass. Work batches 1 to 4 are in commits 9cddb8e, c4ee448 and 291af8b. |

## 1. Summary

Move all task fields from the Field/Value table under the title into the frontmatter, remove `blocks` and `reviewers`, set `owner` to the git `user.name` of the person who creates the task, and make the frontmatter free of template comments (ADR-013, step 1). TASK-029 is merged into this task.

## 2. Problem and Motivation

A task has two field formats: the frontmatter (`id`, `title`, `status`, `priority`, `area`, `depends_on`) and a Field/Value table (`Type`, `Owner`, `Reviewers`, `Platforms`, `Size`, `Blocks`, `Related`). `internal/docs/task/create.go:27` writes the table, and `internal/docs/task/change.go:20` reads `Size` with a second parser. `task_list` cannot filter table values. After the QA cleanup commit `d9ce536`, `Blocks` was wrong in most tasks, because it is the inverse of `depends_on` and a person must keep it correct. `Owner` and `Reviewers` are always `Maintainer`, so they do not tell who created the task.

Also, `internal/docs/task/template.md:3` to `:7` has a `# ...` comment on each frontmatter field, and `markdown.SetField` keeps the comment of the line (`internal/markdown/frontmatter.go:50`). In the v0.1.0 QA session, TASK-021 to TASK-028 got these comments. The rules of the comments are already in the tools: `format` quotes a YAML value when necessary (`internal/markdown/frontmatter.go:58`), the tools check `status` and `priority`, and the tool field descriptions give the examples of `area` and `depends_on`.

## 3. Goals

- G-1: All task fields are in the frontmatter, with one parser and no comments.
- G-2: Each task records the person who created it.
- G-3: `task_list` can filter by each frontmatter field.

## 4. Out of Scope

- The `milestone` field: TASK-030.
- The Document History table and the other body tables.

## 5. Context
### 5.1 Current State

- Frontmatter writes: `markdown.SetField` (`internal/markdown/frontmatter.go:43`), called from `internal/docs/task/create.go:70` to `:74`, `internal/docs/task/change.go:63` to `:89` and `internal/docs/learning/learning.go:87`, `:110`.
- Table writes: `internal/docs/task/create.go:27` and `:93`. Table read for Size: `internal/docs/task/change.go:20`, `:125`.
- Author of Document History rows: `Store.Author` (`internal/docs/task/task.go:44`), from git `user.name` (`internal/mcpserver/server.go`, `gitUser`).
- Only `internal/docs/task/template.md` has frontmatter comments.

### 5.2 Assumptions

- A-1: Each task file was added by one commit that `git log --follow --diff-filter=A` finds. Verify: the migration script stops when it finds no author.

### 5.3 Constraints

- ADR-013 (task fields in the frontmatter), ADR-006 (tools own the documents), ADR-012 (the type sets the commit type), ADR-005 (compact results).
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional
- FR-1: The frontmatter has these fields, in this order: `id`, `title`, `status`, `priority`, `type`, `size`, `owner`, `platforms`, `area`, `depends_on`, `related`. `platforms`, `area`, `depends_on` and `related` are flow lists (Q-2).
- FR-2: The template and new tasks have no Field/Value table.
- FR-3: `blocks` and `reviewers` do not exist, because `depends_on` has the data of `blocks`, and the project has one reviewer.
- FR-4: `task_create` sets `owner` to the git `user.name` of the project, as the Document History author (`internal/docs/task/task.go:44`). With no `user.name`, the value is `Agent`, as for Document History.
- FR-5: `task_create` requires `title`, `priority`, `area` and `type`. `size`, `platforms` and `related` are optional (Q-3).
- FR-6: The tools refuse a `type` that is not `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `build`, `ci` or `chore`, and a `size` that is not `S` or `M`, because the type sets the commit type (ADR-012) and an `L` task must be split (Q-4).
- FR-7: `task_update_fields` changes `type`, `size`, `owner`, `platforms` and `related`. The input has no `blocks` and no `reviewers`.
- FR-8: `task_set_status(In Review)` refuses a task with no `size` or no `platforms`, and reads them from the frontmatter (Q-3).
- FR-9: `task_list` has one optional filter for each frontmatter field (Q-1). A text field matches with no case difference. A list field matches when the list has the value. Example: `task_list(depends_on: TASK-030)` gives the tasks that wait for TASK-030, so `blocks` is not necessary. One match function reads the frontmatter map, so a new field needs no new filter code.
- FR-10: The template frontmatter has no comments, because the tools write the frontmatter (ADR-006). From TASK-029.
- FR-11: `markdown.SetField` does not keep a comment at the end of the line. Remove `comment` (`internal/markdown/frontmatter.go:50`), because no file has frontmatter comments after FR-10. From TASK-029 (Q-2 of TASK-029).
- FR-12: All existing tasks change to the new format in one commit. `owner` is the author of the commit that added the file (`git log --follow --diff-filter=A --format=%an`) (Q-5). A one-time script does the change. The script is not added to the repository, because it runs one time.
- FR-13: The `/dx:task` skill and ADR-006 give the new fields.

### 6.2 Non-Functional
- NFR-1: One frontmatter parse for each task read. No table parse for fields.

## 7. Design

### 7.1 Architecture
Frontmatter writes with `markdown.SetField` in `internal/docs/task/create.go` and `internal/docs/task/change.go`. Remove the table code at `internal/docs/task/create.go:27` and `:93`, the regex at `internal/docs/task/change.go:20`, and `comment` in `internal/markdown/frontmatter.go`. The `task_list` filter reads the parsed frontmatter as a map.

### 7.2 Public API
```text
task_create(title, priority, area, type, depends_on?, size?, platforms?, related?, sections?)
task_update_fields(id, title?, priority?, type?, size?, owner?, platforms?, area?, depends_on?, related?)
task_list(status?, <any frontmatter field>?)
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
platforms: [macOS, Linux, Windows]
area: [docs-tools, plugin]
depends_on: []
related: [ADR-013, ADR-006, TASK-006, TASK-029]
---
```

### 7.4 User-facing Behavior
GitHub shows the frontmatter as a table at the top of the file.

### 7.5 Alternatives
- Keep the Field/Value table: rejected (ADR-013).
- Calculate `blocks`: rejected by the maintainer.
- Remove the comments in `SetField` only: rejected. FR-10 removes the cause, and FR-11 removes the code that has no use after FR-10.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/markdown/frontmatter.go`, `internal/markdown/markdown_test.go` | FR-11. | `go test ./internal/markdown/` |
| 2 | `internal/docs/task/template.md`, `internal/docs/task/create.go`, `internal/docs/task/change.go`, `internal/docs/task/task.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-10. Tests. | `go test ./internal/docs/task/` |
| 3 | `docs/tasks/*.md` | FR-12 with a one-time script. | `task_list(status: all)` lists all tasks; no file has `\| Field \| Value \|` or a frontmatter comment |
| 4 | `internal/plugin/assets/skills/task/SKILL.md`, ADR-006 | FR-13. | Read the diff |

## 9. Acceptance Criteria

- AC-1: When `task_create` runs with git `user.name` `Equal`, then the frontmatter has `owner: Equal`, the file has no Field/Value table, and no frontmatter line has ` # `.
- AC-2: When `task_create` runs with `type: feature` or `size: L`, then the tool refuses it and no file is created.
- AC-3: Given a task with no `size`, when `task_set_status(In Review)` runs, then the tool refuses it.
- AC-4: Given TASK-030 depends on TASK-033, when `task_list(depends_on: TASK-033)` runs, then the result has TASK-030.
- AC-5: Given TASK-A with `platforms: [Windows]`, when `task_list(platforms: windows)` runs, then the result has TASK-A.
- AC-6: After work batch 3, no task file has a Field/Value table or a frontmatter comment, and each task has an `owner` from git history.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Parsers for task fields | 2 (frontmatter and table) | 1 | Code review |
| Task files with a Field/Value table | 35 | 0 | `git grep -l '\| Field \| Value \|' docs/tasks` |

- Tests: unit tests of work batches 1 and 2; `task_list` after work batch 3.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The migration script loses a table value | Medium | The script stops on a value that it cannot map. Compare `task_list(status: all)` before and after. |
| A link or a tool reads the old table | Low | `git grep 'Field \| Value'` after work batch 3. |

## 12. Removal and Compatibility

Revert the commits. The migration commit changes all task files, so revert it with the code.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which new task_list filters does the frontmatter make useful? Options: type, size, owner. Recommended: owner and type, in TASK-033, because each is one equality check on a frontmatter field. Size is only S or M. | Equal | All frontmatter fields are filters (decision of the maintainer). |
| Q-2 | Is platforms a flow list, example [macOS, Linux, Windows], not a text? Recommended: yes, so the filter task_list(platforms: Windows) matches as for area. | Equal | Yes. platforms is a flow list. |
| Q-3 | Which fields are required, and when? Recommended: task_create requires title, priority, area (as now) and type; size and platforms are optional at create and required for In Review, because a Draft can start with no estimate. | Equal | Yes. task_create requires title, priority, area and type. size and platforms are required for In Review. |
| Q-4 | Does the tool check type against the Conventional Commits types (feat, fix, refactor, perf, test, docs, build, ci, chore), and size against S and M? Recommended: yes, as for priority, because the type sets the commit type (ADR-012). | Equal | Yes. The tool checks type against the Conventional Commits types and size against S and M. |
| Q-5 | For the migration of existing tasks, is owner the author of the commit that added the file (git log --follow --diff-filter=A)? Recommended: yes, because the Document History author can be Claude, which is not a person. | Equal | Yes. owner is the author of the commit that added the file. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence. AC-1: `TestCreate` (`internal/docs/task/task_test.go:49`) checks the full frontmatter with `owner` from git user.name, no ` # ` and no Field/Value table. AC-2: `TestCreateRefusesBadInput` (`:82`) refuses `type: feature` and `size: L`. AC-3: `TestSetStatusRules` (`:153`) gives "give a size" and "give the platforms". AC-4, AC-5: `TestListFilterAnyField` (`:130`) for `depends_on` and `platforms: windows`. AC-6: after commit `c4ee448`, `git grep -l '| Field | Value |' docs/tasks` gives 0 files, no frontmatter comment, 35 of 35 tasks have `owner: Equal` from `git log --follow --diff-filter=A`, and `task_list(status: all)` gives the same rows before and after.
- [x] Project checks (build, tests, lint) pass. Local on macOS at `9cddb8e`: `go build`, `go vet`, `go test ./...` pass; golangci-lint v2.15.0: 0 issues. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (`/dx:task` skill, ADR-006). Commit `291af8b`: `internal/plugin/assets/skills/task/SKILL.md` gives the frontmatter fields, the filters and the In Review checks; ADR-006 Decision gives the fields and the task tools.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no call flow or state change. Section 7.3 gives the file format.
- [x] ADRs are recorded for architectural decisions (ADR-013 has the decision). ADR-013 records the decision. ADR-006 refers to ADR-013.
- [x] `/dx:eureka` is done. The tool does not exist yet. Lessons recorded by hand: `docs/learnings/go-yaml-flow-lists.md`, `docs/learnings/claude-code-plugin-reload.md`.