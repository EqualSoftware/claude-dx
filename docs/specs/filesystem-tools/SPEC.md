---
slug: filesystem-tools
title: Filesystem tools for dx
status: Approved
milestone: v0.3.0
priority: P2
area: [mcp, filesystem]
related: [ADR-002, ADR-003, ADR-013, ADR-016]
---

# Filesystem tools for dx

## 1. Summary

The filesystem tools replace the shell file operations (mv, cp, rm, mkdir) in the agent, and add fs_tree to show the project tree. Read, Edit, Write, Glob and Grep stay native (ADR-002). This spec is planned for v0.3.0 and has no code yet.

## 2. Problem

The shell file operations have no check of the project boundary. An agent can delete or replace files of the project with rm, mv or cp without a safety check.

## 3. Goals

- G-1: The agent moves, copies, deletes and creates files of the project with the fs_ tools, not with the shell.
- G-2: The agent sees the project tree in one call, without the files that .gitignore names.
- G-3: The tool list is fs_move, fs_copy, fs_delete, fs_mkdir and fs_tree (Q-1). fs_read_many is not built: the native Edit tool needs a native Read of the file first.

## 4. Out of Scope

- The native tools Read, Edit, Write, Glob and Grep (ADR-002).
- The git operations: see the spec git-tools.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The tools fs_move, fs_copy, fs_delete and fs_mkdir replace the shell file operations mv, cp, rm and mkdir. Each tool refuses a path outside the project root. | Done |
| R-2 | fs_tree lists the files from git ls-files --cached --others --exclude-standard (without git: all files except .git). It takes path (default: project root) and depth (default 3), shows an indented tree and stops at 500 lines with a note that gives the number of lines not shown. | Done |
| R-5 | The Bash hook denies mv, rm, cp and mkdir when a path argument is inside the project root (a path with a variable or a glob counts as inside). It uses the parser of the git block and gives the name of the fs_ tool to use. | Done |
| R-6 | fs_delete runs with no user prompt. It needs recursive: true to delete a directory that is not empty. It refuses the project root and the .git directory. | Done |
| R-7 | fs_move and fs_copy take one source, one destination and overwrite (default false). With false they refuse a destination that exists, file or directory. With true they replace it: a file destination is replaced, a directory destination is deleted first. fs_copy copies a directory with all its contents. | Done |
| R-8 | fs_delete and fs_mkdir take a list of paths. fs_mkdir makes the parent directories and accepts a directory that exists. | Done |
| R-9 | The fs_ tools act on a symbolic link, not on its target. They write no log. | Done |
| R-10 | fs_move, fs_copy and fs_mkdir refuse the project root and the .git directory (or a path inside it), also with another letter case, and fs_move and fs_copy refuse a source and destination that are the same file or that contain each other. | Done |

## 6. Design

### Planned modules

- `internal/fs/`: the fs_ tools. Each tool checks its paths with `Project.Resolve`, as the path guard of mcp-server does (ADR-002).
- `internal/hook/`: the parser of the git block also finds `mv`, `rm`, `cp` and `mkdir`, and gives one deny reason for each program (Q-4).
- `internal/plugin/assets/hooks/hooks.json`: no change. The `PreToolUse` hook on Bash already runs `claude-dx hook pre-tool-use`.

### Tools

| Tool | Arguments | Behavior |
| :--- | :--- | :--- |
| fs_move | source, destination, overwrite (default false) | Moves a file or a directory. |
| fs_copy | source, destination, overwrite (default false) | Copies a file, or a directory with all its contents. |
| fs_delete | paths, recursive (default false) | Deletes files and directories. |
| fs_mkdir | paths | Makes directories and their parents. A directory that exists is not an error. |
| fs_tree | path (default: project root), depth (default 3) | Shows an indented tree of the files from git, maximum 500 lines. |

### Hook rule

The hook denies `mv`, `rm`, `cp` or `mkdir` when a path argument is inside the project root. A relative path resolves from the working directory of the command. A path with a variable or a glob counts as inside. Paths outside the project root (for example temporary directories) are allowed. The hook cannot see commands inside scripts or make targets (Q-3).

Deny reasons: `rm` → fs_delete, `mv` → fs_move, `cp` → fs_copy, `mkdir` → fs_mkdir.

### Data flow

1. The agent calls an fs_ tool. The tool checks the paths, then runs the file operation.
2. The Bash hook denies a shell file operation on a project path and gives the name of the fs_ tool.

### Decisions

- ADR-002: the filesystem tools complement the native tools.
- ADR-003: the Bash hook (see git-tools).
- ADR-013: claude-dx does not manage language servers. fs_move does not update imports.

## 7. Acceptance Criteria

Done requirements: none.

Planned requirements (tests to write, on a temporary directory):
- R-1: tests for fs_move, fs_copy, fs_delete and fs_mkdir, and a test that each path outside the project root is refused.
- R-2: a test that fs_tree does not list a file that .gitignore names, obeys depth, and stops at 500 lines with the note.
- R-5: hook tests that `mv`, `rm`, `cp` and `mkdir` on a project path are denied with the name of the fs_ tool, also inside `sh -c` and pipes, and that the same commands on a path outside the project are allowed.
- R-6: tests that fs_delete refuses a directory that is not empty without recursive, and refuses the project root and `.git`.
- R-7: tests that fs_move and fs_copy refuse a destination that exists, and replace it with overwrite true.
- R-8: tests that fs_delete and fs_mkdir take many paths, and that fs_mkdir makes parents and accepts a directory that exists.
- R-9: a test that fs_delete on a symbolic link deletes the link and keeps the target.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Tool list (G-4): is the final list fs_move, fs_copy, fs_delete, fs_mkdir and fs_tree, without fs_read_many? Fact: the native Edit tool refuses a file that the native Read tool did not read in the session, so a file that fs_read_many reads cannot be edited without a second Read. | Equal | Yes. The list is fs_move, fs_copy, fs_delete, fs_mkdir and fs_tree. R-3 (fs_read_many) is deleted. ADR-002 must change to match. |
| Q-2 | R-4 is withdrawn by ADR-013 but has the status Planned. Do we delete R-4? | Equal | Yes. Delete R-4. |
| Q-3 | Hook scope (R-5): does the hook deny mv, rm, cp and mkdir only when a path argument is inside the project root (a path with a variable or a glob counts as inside), and allow them for paths outside the project (for example temporary directories)? The hook cannot see commands inside scripts or make targets. | Equal | Yes. The hook denies the command only when a path argument is inside the project root. A path with a variable or a glob counts as inside. Paths outside the project root are allowed. |
| Q-4 | Hook parser (R-5): does the hook use the same parser as the git block (internal/hook), with the same nesting rules (sh -c, eval, xargs, env, pipes, subshells), and give one deny reason for each program? | Equal | Yes. The hook uses the parser of the git block in internal/hook and gives one deny reason for each program, with the name of the fs_ tool. |
| Q-5 | fs_delete safety: does fs_delete run with no user prompt, require recursive: true to delete a directory that is not empty, and refuse the project root and the .git directory? | Equal | Yes. fs_delete runs with no user prompt, needs recursive: true for a directory that is not empty, and refuses the project root and the .git directory. |
| Q-6 | Arguments: do fs_delete and fs_mkdir take a list of paths, and do fs_move and fs_copy take one source and one destination? | Equal | Yes. fs_delete and fs_mkdir take a list of paths. fs_move and fs_copy take one source and one destination. |
| Q-7 | Behavior: do fs_move and fs_copy refuse a destination that exists (no overwrite option), does fs_copy copy a directory with all its contents, does fs_mkdir make the parent directories and accept a directory that exists, and do the tools act on a symbolic link, not on its target? | Equal | fs_move and fs_copy take overwrite (bool, default false). With false, they refuse a destination that exists. With true, they replace the destination. fs_copy copies a directory with all its contents. fs_mkdir makes the parent directories and accepts a directory that exists. The tools act on a symbolic link, not on its target. |
| Q-8 | fs_tree: does it get the files from git ls-files --cached --others --exclude-standard (in a directory without git: all files except .git), take path (default: project root) and depth (default 3), show an indented tree, and stop at 500 lines with a note that tells the number of lines not shown? | Equal | Yes. fs_tree gets the files from git ls-files --cached --others --exclude-standard (without git: all files except .git). It takes path (default: project root) and depth (default 3), shows an indented tree and stops at 500 lines with a note that gives the number of lines not shown. |
| Q-9 | Record: section 2 says that shell file operations have no record. Do the fs_ tools write no log (git shows the changes), and do we remove the word record from section 2? | Equal | Yes. The fs_ tools write no log. Section 2 does not use the word record. |
| Q-10 | Decisions found during the build: refuse a source and destination that contain each other; no copy fallback across devices for fs_move; fs_tree does not show empty directories; fs_tree shows a directory deeper than depth as name/; fs_mkdir also refuses .git; a failed restore keeps the backup. Do we accept them? | Equal | Yes, all accepted. Known limits: a race between the check and the rename can replace a destination that another process makes; .fs-old-* or .fs-copy-* entries stay if the process stops during a replace; on Windows a move across devices gives the raw system error. |

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
| 2026-10-27 | Equal | R-4: text changed. |
| 2026-11-02 | Equal | Deleted R-3. |
| 2026-11-02 | Equal | Deleted R-4. |
| 2026-11-02 | Equal | R-5: text changed. |
| 2026-11-02 | Equal | Added R-6. |
| 2026-11-02 | Equal | R-1: text changed. |
| 2026-11-02 | Equal | R-2: text changed. |
| 2026-11-02 | Equal | Added R-7. |
| 2026-11-02 | Equal | Added R-8. |
| 2026-11-02 | Equal | Added R-9. |
| 2026-11-02 | Equal | R-7: text changed. |
| 2026-11-02 | Equal | Added R-10. |
| 2026-11-02 | Equal | Refined with the user (Q-1 to Q-9). |
| 2026-11-02 | Equal | Changed fields: related. |
| 2026-11-02 | Equal | Approved by the user. ADR-016 records the scope. |
| 2026-11-02 | Equal | R-10: text changed. |
| 2026-11-02 | Equal | Status: In Review. The scope changed in section 5. |
| 2026-11-02 | Equal | R-10 extended to fs_mkdir and same-file checks (Q-10). Approved by the user. |
| 2026-11-02 | Equal | R-1, R-2, R-6, R-7, R-8, R-9, R-10: Done. |
| 2026-11-02 | Equal | R-5: Done. |