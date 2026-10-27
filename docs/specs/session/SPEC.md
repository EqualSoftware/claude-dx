---
slug: session
title: Cache lock and sessions
status: Approved
milestone: v0.1.0
priority: P1
area: [cache, session]
related: [ADR-006, ADR-011, ADR-012]
---

# Cache lock and sessions

## 1. Summary

Many claude-dx processes use the cache at the same time. A read-write lock file, cache.lock, tells which processes use it. A process holds the shared lock while it uses the cache. A process that removes cache data takes the exclusive lock.

## 2. Problem

Many sessions run at the same time, in one repository and in worktrees. Without a lock, one process can remove the plugin that another session uses. A PID check or a session file can be wrong after a crash, and a child process can outlive its parent (ADR-006).

## 3. Goals

- G-1: No process removes a plugin that a running session uses.
- G-2: Many sessions run at the same time.
- G-3: The lock ends with the process, with no PID check and no session file.

## 4. Out of Scope

- The plugin extraction: see the spec plugin.
- The docs lock: see the spec docs-common.
- The lock files of the project and the project data: see the spec project-data.
- The commands that use the cache lock: see the specs cli-launch, cli-update, cli-doctor, cli-clean and mcp-server.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Each process that uses the cache holds the shared cache lock <cache>/cache.lock until its session ends. Many sessions run at the same time. | Done |
| R-2 | At start, a process that gets the exclusive cache lock with no wait removes each plugin directory that no session uses, and keeps its own directory. | Done |
| R-3 | A session that starts while the exclusive cache lock is held waits for its release, then starts. | Done |
| R-4 | A removal of global cache data takes the exclusive cache lock with no wait. It is refused while a process holds the shared lock, and it keeps cache.lock and projects/<key>/docs.lock. | Done |
| R-5 | The cache lock has no PID check and no session file. A child process that stays after the main process ends does not hold the cache lock. | Done |
| R-6 | A command process holds the shared cache lock during the command. A long-running process holds it while it runs. | Done |

## 6. Design

### Cache layout

```text
<cache>/                         os.UserCacheDir()/claude-dx (internal/cache/cache.go)
  cache.lock                     read-write lock of all processes
  plugin/<version>-<key>/        release plugin (see plugin)
  plugin/dev-<key>/              dev plugin (see plugin)
  plugin/*.tmp-*/                temporary extraction, removed at start
  projects/<key>/docs.lock       docs lock of a project without git data (see docs-common)
```

### Modules

- `internal/filelock/`: `Lock`, `RLock`, `TryLock`, `TryRLock`. `flock` on Unix, `LockFileEx` on Windows.
- `internal/session/session.go`: `Use` takes the shared cache lock. `Start` takes the exclusive lock with no wait, removes the unused plugin directories, calls `alone`, then takes the shared lock and calls `ensure`.

### Data flow

1. A long-running process calls `session.Start`. It takes the exclusive lock with no wait, removes the unused plugins, and then takes the shared lock for its session.
2. A command process calls `session.Use` and holds the shared lock during the command.
3. A removal of global cache data takes the exclusive lock with no wait. A held shared lock refuses it.

### Decisions

- ADR-006: read-write locks for the cache, and the claude child process.
- ADR-011: the interactive clean command removes global data under the exclusive lock.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestStartTwoSessions`, `TestStartKeepsWhileInUse` (`internal/session/session_test.go`).
- R-2: `TestStartRemovesUnused`, `TestStartCallsAlone` (`internal/session/session_test.go`).
- R-3: `TestStartWaitsForClean` (`internal/session/session_test.go`).
- R-4: `TestGlobalItemDisabled`, `TestGlobalDeleteLockedAfterScan`, `TestNoLockFiles`, `TestScanMissingCache` (`internal/clean`).
- R-5: `TestGlobalItemAfterChildOfClaude` (`internal/clean/clean_unix_test.go`).
- R-6: `TestCleanListMCPServer`, `TestCleanListRefusesSession` (`cmd/claude-dx/main_test.go`).

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
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |