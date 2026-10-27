---
slug: interactive-clean
title: Interactive clean command
status: Draft
milestone: v0.1.0
priority: P1
area: [cli, cache]
related: [ADR-018, ADR-016, ADR-015]
---

# Interactive clean command

## 1. Summary

Replace `cache clean` with the interactive command `claude-dx clean` of ADR-018.

## 2. Problem

`cache clean` removes all of the cache (`internal/cache/cache.go:24`), and that can delete local tasks of the user (ADR-016, ADR-018).

## 3. Goals

- G-1: The user selects each item to delete and confirms. Nothing else is deleted.
- G-2: The user sees all known projects, their data and their local tasks.
- G-3: A project can be cleaned while another project has a session.

## 4. Out of Scope

- Flags that delete with no interactive selection (ADR-018).
- The project file (TASK-068).

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | `claude-dx clean` shows a `charmbracelet/huh` multi-select: Global (old plugin versions, each LSP server) and Projects (for each `project.json`: cache data and local tasks by spec and status). A project path that does not exist is marked missing. | Planned |
| R-2 | After the selection, the command shows the items and their size and asks for a confirmation. | Planned |
| R-3 | Global items need the exclusive cache lock, with no wait. If a session holds the shared lock, they are disabled with the reason. | Planned |
| R-4 | Project items hold the docs lock or the task lock of that project. | Planned |
| R-5 | Lock files are never shown and never deleted (ADR-015). | Planned |
| R-6 | `claude-dx clean --list` prints the tree with sizes and status and deletes nothing. | Planned |
| R-7 | With no interactive terminal, `claude-dx clean` refuses and gives `--list`. | Planned |
| R-8 | Remove `cache clean`. | Planned |

## 6. Design
[How the feature works: modules, data flow, formats, user-facing behavior. Link to the ADRs; do not copy them. A new architecture approach needs an ADR.]

## 7. Acceptance Criteria

- R-1 to R-6: tests of the item tree, the locks and the delete, with the selection given by a test function, not a terminal.
- R-7: test with no terminal.
- Manual test of the selection in Windows Terminal, PowerShell and Git Bash, and on macOS and Linux.
- Tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Added R-1. |
| 2026-10-26 | Equal | Added R-2. |
| 2026-10-26 | Equal | Added R-3. |
| 2026-10-26 | Equal | Added R-4. |
| 2026-10-26 | Equal | Added R-5. |
| 2026-10-26 | Equal | Added R-6. |
| 2026-10-26 | Equal | Added R-7. |
| 2026-10-26 | Equal | Added R-8. |