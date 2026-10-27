# ADR-006: OS locks and the claude child process

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
Many claude-dx processes can run at the same time: launchers, MCP servers of many sessions, `update`, `doctor` and `clean`. They share the cache and the docs of a project. Requirements: no session opens while `clean` removes global data, `clean` never removes global data while a session runs, many sessions of one project run at the same time, and a read of the docs never sees half of a change of more than one file.

A PID check does not prove that a session runs, because a PID can be reused, and PID plus start time is not unique (LWN 943387). A lock file that a process removes is not safe: on Unix, a process that waits gets the lock of a removed file, and two processes hold the lock. On Unix, if the launcher replaces itself with `claude` (exec), the children of `claude` inherit the lock.

## Decision
Two OS read-write locks protect all shared state. No process removes a lock file, because a process that waits on a removed file gets a lock that other processes do not see.

- Cache lock: `<cache>/cache.lock`. Each claude-dx process that uses the cache holds a shared lock: the launcher for the full session, the MCP server while it runs, `update` and `doctor` during the command. `clean` takes the exclusive lock with no wait for the global items, and it disables them while another process holds the lock. A start waits while `clean` holds the exclusive lock. The cache lock stays in the cache, because all projects share the cache.
- Plugin cleanup: at start, the launcher tries the exclusive cache lock with no wait. If it gets it, no other process uses the cache, so it removes the old plugin directories and the `*.tmp-*` directories. Then it takes the shared lock, and only after that it extracts or checks its plugin directory.
- Docs lock: `docs.lock` in the project data directory (ADR-007), with a mutex of the process. The docs tools that read hold a shared lock. The docs tools that write hold the exclusive lock for the full read, change and write. The docs lock protects the specs, ADRs, learnings, the glossary, the local tasks and the number counters, so a docs tool needs only the docs lock. An empty lock path is an error.
- Lock order: the cache lock first, then the docs lock. The project items of `clean` take the shared cache lock, then the docs lock of that project. `MoveData` runs under the exclusive cache lock, so a delete of project data never runs during a move.
- Each document write goes to a temporary file in the same directory, which is then renamed to the file. The file keeps its mode. A reader sees the old or the new content, never a partial file. On Windows, Go opens files with no `FILE_SHARE_DELETE`, so the rename tries again for 1 second while a reader has the file open.
- `filelock` gives a shared and an exclusive mode: `flock` `LOCK_SH` and `LOCK_EX` on Unix, `LockFileEx` with and without `LOCKFILE_EXCLUSIVE_LOCK` on Windows, on one byte at offset 1<<62 so other handles can read the file.
- The launcher runs `claude` as a child process on all platforms, with the same stdio and environment, and waits for it. Go opens files with close-on-exec, so no child of `claude` gets the cache lock, and the lock lives as long as the session. The launcher returns the exit code of `claude`. If a signal ended `claude`, the launcher returns 128 plus the signal number.
- Signals: the terminal sends Ctrl+C (and SIGQUIT on Unix) to `claude` too, so the launcher catches these signals and does nothing. On Unix, the launcher sends the SIGTERM and SIGHUP that it gets to `claude`. The launcher catches signals and does not ignore them, because an ignored signal stays ignored in a child after exec.

## Alternatives
- One lock file for each session, with a PID check: Rejected. A reused PID keeps an ended session as running, and the lock files stay a second source of truth.
- Remove the cache lock file at the end of `clean`, and after each lock compare the locked file with the file at the path: Rejected. It needs a retry loop on all lock calls for one empty file.
- A start fails while `clean` runs: Rejected. The lock does not tell which process holds it, so two starts at the same time also fail.
- An exclusive docs lock for the read tools too: Rejected. It gives no more safety than a shared lock, and parallel reads of subagents and sessions wait in a queue. SQLite, Cargo and Nix use shared locks for readers.
- A lock with exclusive create (O_EXCL), as git `index.lock`: Rejected. A crash leaves the file, and all starts stop until a person removes it.
- Keep exec on Unix and remove close-on-exec only for `claude`: Rejected. `claude` starts its own children, and claude-dx does not control their descriptors.

## Consequences
- Positive: the OS lock is the only source of truth: no PID check, no session files and no stale lock after a crash. `clean` never removes global data while a claude-dx process uses the cache, and no session opens during that removal. Many sessions run at the same time. A read of the docs never sees half of a change. The launcher code is the same on all platforms.
- Negative: one idle launcher process for each session. Old plugin directories stay until a session starts with no other claude-dx process. `cache.lock` and `docs.lock` stay as empty files.
- Follow-up: none.

## Evidence
- internal/session/session.go:13-15 (`cache.lock`), :19-24 (`Use`, shared lock), :29-48 (`Start`, exclusive try, then shared), :52-59 (remove the unused plugin directories).
- internal/filelock/filelock.go:15-36 (modes), internal/filelock/lock_unix.go:11-23 (`flock`), internal/filelock/lock_windows.go:12-27 (`LockFileEx`, offset 1<<62).
- internal/docs/docfile/write.go:24-36 (`Locked`, `RLocked`), :69-88 (`Write`, temporary file and rename), :105-110 (rename retry on Windows).
- internal/launcher/start.go:19-33 (child process, exit code), :52-64 (128 plus signal), internal/launcher/signals_unix.go:12-15 (ignored and forwarded signals).
- internal/launcher/launcher.go:46-56 (the lock lives until `claude` ends).
- internal/mcpserver/server.go:65-69 and cmd/claude-dx/main.go:93-97 (shared lock of the server and `update`).
- internal/clean/scan.go:92-103 (global item disabled while a session runs), :108-117 (exclusive try before the delete), :203-214 (`shared`), :230-232 and :311-312 (shared cache lock, then the docs lock of the project).
- internal/launcher/launcher.go:41-45 and :68-77 (`MoveData` runs in `alone`, under the exclusive cache lock).
- LWN 943387 on PID and start time. Cargo `cache_lock` and Lix `gc.lock` use the same shared and exclusive model.
