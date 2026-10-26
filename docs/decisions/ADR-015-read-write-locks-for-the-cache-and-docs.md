# ADR-015: Read-write locks for the cache and docs

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context

Replaces ADR-014.

ADR-014 detects a running session with one lock file for each PID. On Unix the launcher replaces itself with claude (exec), and the children of claude inherit the lock, so a held lock does not prove that a session runs. TASK-050 added a PID check, but a reused PID keeps an ended session as running, and PID plus start time is not unique (LWN 943387). The global lock `sessions.lock` is removed by `cache clean`, so on Unix a process that waits gets the lock of a removed file, and two processes hold the cache lock. `cache clean` also removes `projects/<key>/docs.lock`, which an MCP server of an ended session can hold. The docs tools that read take no lock, so a read during a change of more than one file (`adr_supersede`, `task_renumber`, `task_create` and the README) can see half of the change. Requirements: no session opens while `cache clean` runs, `cache clean` never runs while a session runs, and many sessions of one project run at the same time.

## Decision
Two OS read-write locks protect all shared state. No process removes a lock file, because a process that waits on a removed file gets a lock that other processes do not see.

- Cache lock: `<cache>/cache.lock`. Each claude-dx process that uses the cache holds a shared lock: the launcher for the full session, the MCP server while it runs, `update` and `doctor` during the command. `cache clean` takes the exclusive lock with no wait. If another process holds the lock, it fails with "a claude-dx session uses the cache". A start waits while `cache clean` holds the exclusive lock.
- Plugin cleanup: at start, the launcher tries the exclusive cache lock with no wait. If it gets it, no other process uses the cache, so it removes the old plugin directories and the `*.tmp-*` directories. Then it takes the shared lock, and only after that it extracts or checks its plugin directory.
- The launcher runs claude as a child process on all platforms and waits for it. Go opens files with close-on-exec, so no child of claude gets the cache lock, and the lock lives as long as the session. The launcher ignores SIGINT, because the terminal also sends it to claude, and sends SIGTERM and SIGHUP to claude. It returns the exit code of claude.
- Docs lock: `<cache>/projects/<key>/docs.lock`, with a mutex of the process. The docs tools that read hold a shared lock. The docs tools that write hold the exclusive lock for the full read, change and write. The docs lock also protects the number counter (TASK-059).
- `filelock` gives a shared and an exclusive mode: `flock` `LOCK_SH` and `LOCK_EX` on Unix, `LockFileEx` with and without `LOCKFILE_EXCLUSIVE_LOCK` on Windows, on one byte at offset 1<<62 so other handles can read the file.
- The session files in `<cache>/sessions/`, the PID check and the exec on Unix go away.

## Alternatives
- Keep the session files of ADR-014 and add the process start time to the PID check: Rejected. The start time is not unique enough, and the session files stay a second source of truth next to the lock.
- Remove the cache lock file at the end of cache clean, and after each lock compare the locked file with the file at the path: Rejected. It needs a retry loop on all lock calls for one empty file. util-linux flock and OpenStack Swift need this loop for the same reason.
- A start fails while cache clean runs: Rejected. The lock does not tell which process holds it, so two starts at the same time also fail.
- An exclusive docs lock for the read tools too: Rejected. It gives no more safety than a shared lock, and parallel reads of subagents and sessions wait in a queue. SQLite, Cargo and Nix use shared locks for readers.
- A lock with exclusive create (O_EXCL), as git index.lock: Rejected. A crash leaves the file, and all starts stop until a person removes it.
- Keep exec on Unix and remove close-on-exec only for claude: Rejected. claude starts its own children, and claude-dx does not control their descriptors.

## Consequences
- Positive: The OS lock is the only source of truth: no PID check, no session files and no stale lock after a crash. cache clean never runs while a claude-dx process uses the cache, and no session opens during cache clean. Many sessions run at the same time. A read of the docs never sees half of a change. Less platform code: inherit_unix.go, inherit_windows.go and exec_unix.go go away.
- Negative: One idle launcher process for each session. cache clean does not give the PID of the session that uses the cache. Old plugin directories stay until a session starts with no other claude-dx process. cache.lock and docs.lock stay after cache clean as empty files. cache clean also removes the number counters, so numbers come again only from the files of the checkout (ADR-007). A claude-dx v0.1.0-rc.1 that runs at the same time uses the old lock files.
- Follow-up: Supersede ADR-014. Change the docs lock text of ADR-007 (TASK-059). Tasks: run claude as a child on Unix, the read-write cache lock, and the shared docs lock for read tools. Cancel TASK-051.

## Evidence
Code review of 2026-10-26: internal/session/session.go:28 (sessions directory made before the lock) and internal/session/session.go:69 (lock file removed after unlock). Callers of Alloc.Next run in docfile.Locked: internal/docs/task/create.go:50, internal/docs/adr/store.go:135, internal/docs/task/change.go:560. References: Cargo cache_lock (https://doc.rust-lang.org/beta/nightly-rustc/cargo/util/cache_lock/index.html), Lix gc.cc gc.lock, OpenStack Swift lock_file race fix, LWN 943387 on PID and start time.
