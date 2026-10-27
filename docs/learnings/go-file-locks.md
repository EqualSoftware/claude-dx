---
topic: go-file-locks
description: How OS file locks from Go behave between handles, processes and child processes. Use it when code takes flock or LockFileEx locks.
---

- `flock` on Unix and `LockFileEx` on Windows lock one open file description or one handle, not one process. Two opens of the same file in one process conflict as two processes do, so a test can hold a lock with a second open in the same process. Example: `TestStartKeepsWhileInUse` (`internal/session/session_test.go`).
- A shared lock (`LOCK_SH`, or `LockFileEx` with no `LOCKFILE_EXCLUSIVE_LOCK`) lets many holders read, and an exclusive lock waits for all of them. The tests `TestSharedLock` and `TestExclusiveLockExcludesShared` (`internal/filelock/filelock_test.go`) prove it with helper processes.
- Go opens files with close-on-exec, so a child process does not get a lock of its parent, and a child that runs after its parent ends does not hold the lock. The test `TestCleanAfterChildOfClaude` (`internal/session/session_unix_test.go`) proves it.
