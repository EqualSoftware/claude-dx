---
slug: mcp-server
title: MCP server of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [mcp, project]
related: [ADR-002, ADR-005, ADR-007, ADR-012]
---

# MCP server of claude-dx

## 1. Summary

The MCP server `claude-dx mcp` gives the dx tools to Claude Code over stdio. It holds the shared packages that the tools use: the project root, the path guard, the project key, the file lock, and the Markdown and result helpers.

## 2. Problem

The dx tools must run in a session with no change to the project. Each tool must refuse a path outside the project, and each change must keep the documents of two parallel sessions correct. The shared helpers must work the same on macOS, Linux and Windows.

## 3. Goals

- G-1: Claude Code connects to `claude-dx mcp` and shows the server as connected.
- G-2: Each tool helper refuses a path outside the project root, also through `..` and symlinks.
- G-3: A test finds a missing tool registration or a wrong tool input schema.
- G-4: The result tables stay readable with non-ASCII text.
- G-5: The server checks for a new release at most one time in 24 hours.

## 4. Out of Scope

- The document tools themselves: see the specs of each document type (docs-task, docs-adr and the others).
- The docs lock: see the spec docs-common. The number allocation and the project file: see the spec project-data.
- The cache lock: see the spec session.
- The git tools and the filesystem tools: see the specs git-tools and filesystem-tools.
- The update command and the notice at start: see the specs cli-update and cli-launch.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | claude-dx mcp serves MCP on stdin and stdout with the server key mcp. Logs go to stderr only. | Done |
| R-2 | The project root is the git top level, or the current directory outside git. | Done |
| R-3 | A path of a tool resolves inside the project root. A path with .. or a symlink that leaves the root is refused with the message Path <rel> is outside the project. | Done |
| R-4 | The server registers a fixed list of tools. A missing or extra tool makes the test fail. | Done |
| R-5 | A test client calls a flow of the task, ADR, glossary and learning tools through an in-memory MCP transport, with no process. | Done |
| R-6 | The file lock gives an exclusive lock and a shared lock, each with a blocking and a no-wait form. It uses flock on Unix and LockFileEx on Windows. | Done |
| R-7 | The Markdown helpers read and write YAML frontmatter, and read or replace one section by its number or heading, with no change to the other sections. | Done |
| R-8 | The result tables align the columns by the number of characters, not bytes. | Done |
| R-9 | The MCP server adds less than 100 ms to the start of the `claude-dx` process, and it has no data race. | Planned |
| R-10 | While it runs, the server holds the shared cache lock. | Done |
| R-11 | At start, the server writes the project file of the cache (see project-data). | Done |
| R-12 | The update check writes cache update.json at most once in 24 hours. Network errors are ignored and the file does not change. | Done |

## 6. Design

### Modules

- `cmd/claude-dx/main.go` and `internal/cli/cli.go`: the `mcp` command starts `mcpserver.New` on stdio.
- `internal/mcpserver/server.go`: registers the tools, holds the shared cache lock (see session), writes the Project file (cache) at start (see project-data), and starts the update check in the background.
- `internal/update/update.go`: `Check` writes `<cache>/update.json` at most once in 24 hours (ADR-005). It ignores network errors.
- `internal/project/project.go`: `Root`, `Key` (first 16 hex characters of the SHA-256 of the git common directory), and `Resolve`. `Resolve` resolves `..` and symlinks, and refuses a path outside the root (ADR-002).
- `internal/filelock/`: the OS file lock, exclusive and shared (`lock_unix.go`, `lock_windows.go`).
- `internal/markdown/`: YAML frontmatter with `go.yaml.in/yaml/v3`, and the section reader and writer.
- `internal/result/result.go`: the compact result text and the table. The table counts characters with `utf8.RuneCountInString`.

### Plugin file

- `internal/plugin/assets/.mcp.json`: the server key is `mcp` and the command is `{{BINARY}} mcp` (ADR-002). The tool names are `mcp__plugin_dx_mcp__<tool>`.

### Data flow

1. Claude Code starts `{{BINARY}} mcp` from the plugin file.
2. The server takes the shared cache lock, writes the Project file (cache), and starts the update check.
3. The server sends the tool list. Each tool call gets the project root and the docs lock, then reads or writes the files.
4. The result is text, with one line for each row.

### Decisions

- ADR-002: the MCP server on stdio, the tool conventions and the document tools.
- ADR-005: the MCP server does the update check.
- ADR-007: the Project file (cache).

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestMCPServerInitialize` (`cmd/claude-dx/main_test.go`).
- R-2: `TestOpenFindsRootFromSubdir` and `TestOpenOutsideGit` (`internal/project/project_test.go`).
- R-3: `TestResolve` and `TestResolveRefusesSymlinkOutside` (`internal/project/project_test.go`).
- R-4: `TestToolNames` (`internal/mcpserver/server_test.go`).
- R-5: `TestToolsFlow` (`internal/mcpserver/server_test.go`), with the in-memory transport.
- R-6: `TestLockExcludes`, `TestTryLock`, `TestSharedLock`, `TestExclusiveLockExcludesShared`, `TestLockWaitsForShared` (`internal/filelock/filelock_test.go`).
- R-7: `TestSplitJoin`, `TestSection`, `TestReplaceSection`, `TestSectionNumber` (`internal/markdown/markdown_test.go`).
- R-8: `TestTableRunes` (`internal/result/result_test.go`).
- R-10: `TestCleanListMCPServer`, `TestCleanListRefusesSession` (`cmd/claude-dx/main_test.go`).
- R-11: `TestProjectFileKeepsCountersOutsideGit` (`internal/mcpserver/server_test.go`), `TestProjectFileHasMainRoot` and `TestWriteProjectFileBadContent` (`internal/mcpserver/projectfile_test.go`).
- R-12: `TestCheck` and `TestCheckIgnoresErrors` (`internal/update/update_test.go`).

Planned requirements (tests to write):
- R-9 start time: the start time is the time from the start of the `claude-dx mcp` process to the initialize result. The process cost is the time of a run of `claude-dx --version` with the same binary, to its exit. The test uses the binary that is built without the race detector. The test measures each one 3 times and uses the fastest of each. The fastest start minus the fastest process cost must be less than 100 ms. The reason: the process start on the Windows runner takes more than 100 ms alone (CI run 40563088951: fastest start 171.6 ms).
- R-9 data race: a test calls many tools at the same time through the in-memory transport. `go test -race` runs it in the acceptance check and in CI.

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
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | Added R-10. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 5. |
| 2026-10-27 | Equal | Added R-11. |
| 2026-10-27 | Equal | Added R-12. |
| 2026-10-27 | Equal | R-10: status Done. |
| 2026-10-27 | Equal | R-11: status Done. |
| 2026-10-27 | Equal | R-12: status Done. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | Approved by the user after the split of cli and locks-and-cache (ADR-012). |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 7. |
| 2026-10-27 | Equal | Approved by the user with the R-9 test criteria. |
| 2026-10-27 | Equal | R-9: Done. |
| 2026-10-27 | Equal | R-9: text changed. |
| 2026-10-27 | Equal | R-9: status Planned. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 5. |
| 2026-10-27 | Equal | Approved by the user: R-9 start time subtracts the process cost. |