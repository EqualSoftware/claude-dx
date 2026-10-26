---
id: TASK-042
title: Rename the task file with the title
status: Done
priority: P2
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-040, TASK-048]
related: [TASK-028, TASK-036]
---

# TASK-042: Rename the task file with the title

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`task_update_fields` with a new title renames the task file to `TASK-NNN-<new slug>.md`.

## 2. Problem and Motivation

`UpdateFields` (`internal/docs/task/change.go:136`) changes the `title` field and the `#` header, but not the file name. The file name keeps the slug of the old title, so the name and the title do not agree.

## 3. Goals

- G-1: The file name of each task has the slug of its current title.

## 4. Out of Scope

- ADR titles: `adr_update` cannot change a title.
- A rename of existing files with an old slug: the rename happens at the next title change.

## 5. Context
### 5.1 Current State

`kind.FileName` (`internal/docs/docfile/docfile.go:73`) makes the name. The README index uses the file name for the link (`internal/docs/task/index.go`). Other tasks refer to a task by ID, not by file name.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: When the title changes and the new file name is different, `UpdateFields` writes the task to the new name and removes the old file.
- FR-2: If a file with the new name exists, the tool refuses the change and changes nothing.
- FR-3: The README index is made again with the new name.
- FR-4: The result gives the new file name.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

`Store.edit` (`internal/docs/task/task.go:196`) gets the target path from the change function, so `UpdateFields` sets the new path.

### 7.2 Public API

`task_update_fields` result: `Updated TASK-NNN. File: TASK-NNN-<slug>.md.` when the name changes.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Keep the old name, for links: rejected by the maintainer, because the references use the ID.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/task/task.go`, `internal/docs/task/change.go`, `internal/docs/task/tools.go`, `internal/docs/task/task_test.go` | FR-1 to FR-4, with tests. | `go test ./internal/docs/task/` |

## 9. Acceptance Criteria

- AC-1: Given TASK-001 with the title `Old`, when `task_update_fields` sets the title `New name`, then the file is `TASK-001-new-name.md`, the old file does not exist and the README links to the new file.
- AC-2: Given a title change that keeps the slug, then the file name does not change.
- AC-3: Given a file with the new name, then the tool refuses the change.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Task files with the slug of an old title after a title change | All | 0 | AC-1 test |

- Tests: one unit test for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A link to the old file name in a doc outside `docs/tasks` breaks | Low | The rules tell the model to refer to a task by ID. |

## 12. Removal and Compatibility

Revert the change. The renamed files stay valid.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Rename the file when the title changes? | Equal | Yes (grill Q-8). |
| Q-2 | Refuse the change when a file with the new name exists, and give the new name in the result? | Equal | Yes (grill Q-29). |
| Q-3 | Order with the other tasks that change `internal/docs/task/change.go`? | Equal | After TASK-040 (grill Q-25). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: TestRenameOnTitle passes: AC-1 renames to TASK-001-new-name.md, the old file does not exist and README.md links to the new file; AC-2 "New Name" keeps the name (file ""); AC-3 refuses the change and the file does not change. TestRenameExists: rename refuses an existing path with "b.md exists" and changes no file (FR-2). FR-4: task_update_fields returns "Updated TASK-NNN. File: <name>." (internal/docs/task/tools.go). Commit after this check.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues; govulncheck: No vulnerabilities found. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (no docs change expected).
  - Evidence: Not applicable: no doc describes the task file name after a title change. The UpdateFields comment in internal/docs/task/change.go describes the rename.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: Not applicable: the rename and retitle helpers in internal/docs/task, no design diagram.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: the file name format of ADR-006 does not change; the tool keeps it in agreement with the title.