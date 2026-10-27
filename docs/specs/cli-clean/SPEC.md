---
slug: cli-clean
title: Clean command of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [cli]
related: [ADR-006, ADR-008, ADR-011, ADR-012]
---

# Clean command of claude-dx

## 1. Summary

claude-dx clean shows the local data of claude-dx as a tree: old plugin versions, projects and their task groups. The user selects items, the command shows what it will delete with sizes, and it deletes only after a Yes.

## 2. Problem

The user must remove the local data of claude-dx without losing tasks that they want to keep (ADR-008, ADR-011). A flag that deletes with no selection can delete too much, so the command is interactive.

## 3. Goals

- G-1: `claude-dx clean` deletes only the items that the user selects and confirms (ADR-011).

## 4. Out of Scope

- Flags that delete with no interactive selection (ADR-011).
- The rules of the cache lock and the docs lock: see the specs session and docs-common.
- The lock files and the project data directory: see the spec project-data.
- The delete of one task (use task_delete), and the delete of the project file or the counters of a project that exists.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | claude-dx clean shows a multi-select of the old plugin versions, the projects and their task groups, and deletes only the selected items after a confirmation. | Done |
| R-2 | After the selection, clean shows the selected items with their size and asks for a confirmation. The default answer is No. Without a Yes, nothing is deleted. | Done |
| R-3 | While a session holds the shared cache lock, clean shows the global items as disabled with the reason, and does not delete them (see session). | Done |
| R-4 | clean takes the exclusive docs lock of a project and waits for it for the project items. A project that has a session can be cleaned. | Done |
| R-5 | clean never shows or deletes lock files. A project folder that has only lock files is not shown (see project-data). | Done |
| R-6 | claude-dx clean --list prints the tree with sizes and status and deletes nothing. | Done |
| R-7 | Without an interactive terminal, claude-dx clean refuses, prints the text that names --list, and exits with code 1. | Done |
| R-8 | The command cache clean is removed. Only clean remains. | Done |

## 6. Design

### Modules

- `internal/clean/`: scans the cache and the project data, builds the tree, runs the selection and the confirmation, and deletes the items.
- `internal/cli/cli.go` and `cmd/claude-dx/main.go`: the command `clean` and its flag `--list`.

### Data flow

1. `clean` scans the cache and the projects. Global items need the exclusive cache lock with no wait (see session). If a session holds the shared lock, the items are disabled with the reason.
2. The user selects items. The command shows the selected items with their size and asks for a confirmation. The default answer is No.
3. Each project item takes the exclusive docs lock of the project and waits for it (see docs-common).
4. The command never shows or deletes lock files (see project-data).

### Decisions

- ADR-006: read-write locks.
- ADR-008: specs and tasks model.
- ADR-011: the interactive clean command.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestScanTree`, `TestPendingMoveIsNotUnknown`, `TestUnknownFolderListOnly` (`internal/clean`).
- R-2: `TestNoSelection`, `TestAnswerNo`, `TestAnswerYes`, `TestNoDataToDelete`, `TestNoDataToDeleteOnlyDisabled`, `TestAllSkippedIsNotDeleted`, `TestFormatSize` (`internal/clean`).
- R-3: `TestGlobalItemDisabled`, `TestGlobalDeleteLockedAfterScan` (`internal/clean`) and `TestGlobalItemAfterChildOfClaude` (`internal/clean/clean_unix_test.go`).
- R-4: `TestTaskDeleteWaitsForDocsLock`, `TestProjectDeleteWaitsForCacheLock`, `TestTaskChangedAfterSelection` (`internal/clean`).
- R-5: `TestNoLockFiles`, `TestRepositoryDocsLockStays` (`internal/clean`), and the symlink tests `TestSymlinksSkipped`, `TestSpecFolderLinkAfterScan`, `TestTasksFolderLinkAfterScan` (unix only).
- R-6: `TestListDeletesNothing` (`internal/clean`), `TestCleanListMCPServer` (`cmd/claude-dx/main_test.go`).
- R-7: `TestCleanNeedsTerminal` (`cmd/claude-dx/main_test.go`).
- R-8: `TestHelpHasNoCacheClean` (`cmd/claude-dx/main_test.go`) and the `cache clean` case of `TestParse` (`internal/cli/cli_test.go`).

The tests give the selection and the answer with functions, not a terminal.

Manual check (open): the selection in Windows Terminal, PowerShell and Git Bash, and on macOS and Linux.

All tests pass on macOS. Linux and Windows run in CI.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification. Split from cli. |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 4. |
| 2026-10-27 | Equal | Approved by the user after the removal of the LSP item (ADR-013). |