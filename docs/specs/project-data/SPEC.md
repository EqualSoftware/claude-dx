---
slug: project-data
title: Project data directory and key
status: Approved
milestone: v0.1.0
priority: P1
area: [project, docs]
related: [ADR-006, ADR-007, ADR-011, ADR-012]
---

# Project data directory and key

## 1. Summary

Each project has a data directory for the local tasks and the counters, and a project file in the cache. The project key names the project in the cache. This spec sets the place of the data, the key, the project file, the number allocation, and the rule for lock files.

## 2. Problem

A repository can have many git worktrees, and a folder can have no git. All of them need one place for the local data, and the number of each new ADR must be unique in the project (ADR-007). A lock file that is removed while a process waits on it can give two processes the same lock (ADR-006).

## 3. Goals

- G-1: The number of each new ADR is unique in one project, with no lock order rule.
- G-2: Each git worktree of one repository uses one data directory and one key.
- G-3: No process removes a lock file.

## 4. Out of Scope

- The docs lock and the atomic write: see the spec docs-common.
- The cache lock: see the spec session.
- The writer of the project file: see the spec mcp-server.
- The removal of data by the user: see the spec cli-clean.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Number allocation takes no lock of its own. The caller holds the docs lock, and the counters are in project.json of the project data directory. | Done |
| R-2 | The project data directory is <git common dir>/dx/ in git, and <cache>/projects/<key>/ outside git. A pending move of the old data runs with a temporary folder and a rename. | Done |
| R-3 | The project file of the cache, <cache>/projects/<key>/project.json, has the project path and the last use time in RFC 3339 UTC. | Done |
| R-4 | The project key is the first 16 hex characters of the SHA-256 of the git common directory, or of the root outside git. Worktrees of one repository share the key and the data. | Done |
| R-5 | No process removes a lock file. The lock files stay after a removal of data, because a process that waits on a removed file can hold a lock that another process does not see. | Done |

## 6. Design

### Layout

```text
<git common dir>/dx/             data directory of a project in git
  project.json                   counters
  tasks/<spec>/                  local tasks
<cache>/projects/<key>/          data directory outside git, and the project file
  project.json                   project path and last use time (ADR-007)
  docs.lock                      docs lock of a project without git data
```

### Modules

- `internal/numbering/numbering.go`: `Next` allocates the number. It reads and writes `project.json` in the data directory. The caller holds the docs lock.
- `internal/project/data.go`: `DataDir` and `MoveData`. The move copies the data to `<Common>/dx.tmp-*` and renames it to `<Common>/dx/`.
- `internal/project/project.go`: `Open`, `Key` and the git common directory.

### Data flow

1. A caller takes the docs lock (see docs-common).
2. The caller gets the data directory, and `Next` reads and writes the counter in `project.json`.
3. A lock file stays after a removal of data.

### Formats

- `project.json` of the cache: the project path and the last use time in RFC 3339 UTC.

### Decisions

- ADR-006: no process removes a lock file.
- ADR-007: the project data directory, the project file of the cache and number allocation.
- ADR-011: the interactive clean command keeps the lock files.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestNextTakesNoLock`, `TestNextUsesHighestFileAndCounter`, `TestNextParallelProcesses`, `TestNextKeepsOtherKeys` (`internal/numbering/numbering_test.go`) and `TestNextWithNullFile` (`internal/numbering/null_test.go`).
- R-2: `TestDataDir`, `TestDataDirPendingMove` (`internal/project/data_test.go`).
- R-3: `TestProjectFile` (`internal/numbering/numbering_test.go`), `TestProjectFileKeepsCountersOutsideGit` (`internal/mcpserver/server_test.go`), `TestProjectFileHasMainRoot` and `TestWriteProjectFileBadContent` (`internal/mcpserver/projectfile_test.go`).
- R-4: `TestWorktreesShareKey` (`internal/project/project_test.go`).
- R-5: `TestNoLockFiles`, `TestAnswerYes`, `TestRepositoryDocsLockStays` (`internal/clean`) and `TestMoveData`, `TestMoveDataOutsideGit`, `TestMoveDataFails`, `TestMoveDataWarning` for the move.

All tests pass on macOS. Linux and Windows run in CI.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification. Split from locks-and-cache. |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |