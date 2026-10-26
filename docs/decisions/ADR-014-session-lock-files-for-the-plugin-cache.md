# ADR-014: Session lock files for the plugin cache

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context
The launcher extracts one plugin directory for each release and for each dev build (ADR-001, TASK-027). No code removed old directories, so the cache grew with each build. `cache clean` could remove files that a running session reads later. On Unix the launcher replaces itself with claude (exec), so no claude-dx code runs when a session ends. Windows cannot remove an open file, and a Windows file lock also stops reads of the locked bytes from other handles.

## Decision
Each session holds an exclusive OS lock on `<cache>/sessions/<pid>.lock`. The file content is the name of the plugin directory of the session. On Unix the descriptor has no close-on-exec, so claude holds the lock after exec; the OS releases it when the process ends, also after a crash. On Windows the launcher holds the file until claude ends, then removes it. The global lock `<cache>/sessions.lock` serializes the session start (`plugin.Ensure`, session file, cleanup) and `cache clean`. A session file with a free lock is of a session that ended. At each start, the launcher removes each plugin directory that no running session uses. `cache clean` removes nothing while a session holds its lock; else it removes `sessions.lock` last, after it releases the lock. On Windows, `filelock` locks one byte at offset 1<<62, after the end of the file, so other handles can read the file. The code is in `internal/session`.

## Alternatives
- One shared lock in each plugin directory: Rejected. Session files can be removed when a session ends with no cleanup, and need only an exclusive lock.
- A random session ID as the file name: Rejected. The PID is sufficient, because the lock and not the name tells if the session runs.
- Remove old plugin directories only with cache clean: Rejected. Each dev build adds a directory.
- cache clean releases the global lock before it removes the cache: Rejected. A session that starts after the check can lose its files.

## Consequences
- Positive: The plugin cache keeps only the directories of running sessions and of the new session. cache clean never removes the files of a running session. A crash leaves no lock that blocks work.
- Negative: Child processes of claude on Unix inherit the session descriptor, so a child that stays after claude ends keeps the lock. A running session of a claude-dx version before this ADR has no session file, so a new start removes its plugin directory. Session files of ended Unix sessions stay until the next start or cache clean.
- Follow-up: None.

## Evidence
TASK-041. internal/session/session.go, internal/filelock/lock_windows.go. TestLockStaysAfterExec (internal/session/session_unix_test.go) passes, and fails when inherit does not clear close-on-exec. TestCacheCleanRefusesSession (cmd/claude-dx/main_test.go).
