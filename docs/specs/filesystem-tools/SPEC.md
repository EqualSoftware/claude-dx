---
slug: filesystem-tools
title: Filesystem tools for dx
status: Draft
milestone: v0.3.0
priority: P2
area: [mcp, filesystem]
related: [ADR-002]
---

# Filesystem tools for dx

## 1. Summary

The filesystem tools replace the shell file operations (mv, cp, rm, mkdir) in the agent, and add operations that save tool calls. Read, Edit, Write, Glob and Grep stay native (ADR-002). This spec is planned for v0.3.0 and has no code yet.

## 2. Problem

The shell file operations have no check of the project boundary and no record. Some operations need many tool calls (for example to read several files or line ranges). A move of a file also needs its imports to change, and that needs the language servers.

## 3. Goals

- G-1: The agent moves, copies, deletes and creates files with the fs_ tools, not with the shell.
- G-2: The agent reads several files or line ranges in one call.
- G-3: A move of a file updates its imports.
- G-4: The full tool list is decided when v0.3.0 starts (ADR-002). The list is an open question of this spec.

## 4. Out of Scope

- The native tools Read, Edit, Write, Glob and Grep (ADR-002).
- The git operations: see the spec git-tools.
- The LSP servers themselves: see the spec lsp-servers.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The tools fs_move, fs_copy, fs_delete and fs_mkdir replace the shell file operations mv, cp, rm and mkdir. | Planned |
| R-2 | The tool fs_tree lists the project tree and obeys .gitignore. | Planned |
| R-3 | The tool fs_read_many reads many files or line ranges in one call. | Planned |
| R-4 | The tool fs_move updates the imports of the moved file with the LSP servers (lsp-servers). | Planned |
| R-5 | The Bash hook blocks mv, rm, cp and mkdir in the Bash tool, and gives the name of the fs_ tool to use. | Planned |

## 6. Design

### Planned modules

- `internal/fs/`: the fs_ tools. Each tool checks that its paths are inside the project root, as the path guard of mcp-server does.
- `internal/plugin/assets/hooks/hooks.json`: the Bash hook of git-tools also blocks `mv`, `rm`, `cp` and `mkdir`.

### Data flow (planned)

1. The agent calls an fs_ tool. The tool checks the paths, then runs the file operation.
2. For fs_move, the tool asks the LSP servers for the imports that change, and it updates them (lsp-servers).
3. The Bash hook denies the shell file operation and gives the name of the fs_ tool.

### Decisions

- ADR-002: the filesystem tools complement the native tools.
- ADR-003: the Bash hook (see git-tools).
- ADR-004: the LSP servers (see lsp-servers).

## 7. Acceptance Criteria

Done requirements: none.

Planned requirements (tests to write):
- R-1: tests on a temporary directory for fs_move, fs_copy, fs_delete and fs_mkdir, and a test that each path outside the project root is refused.
- R-2: a test that fs_tree does not list a file that .gitignore names.
- R-3: a test that fs_read_many returns the files and the line ranges of one call.
- R-4: a test of fs_move on a Go file and a TypeScript file that updates the import in another file (needs lsp-servers).
- R-5: a hook test that the Bash commands `mv`, `rm`, `cp` and `mkdir` are denied with the name of the fs_ tool.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Changed fields: related. |