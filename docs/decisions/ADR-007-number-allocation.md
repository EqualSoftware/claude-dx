# ADR-007: Allocate numbers with a counter and lock

- Status: Superseded by ADR-019
- Date: 2026-10-25
- Owners: Maintainer

## Context
Many agents can create ADRs or tasks at the same time: subagents of one session, two sessions, worktrees, and branches. Two documents with the same number cause a conflict at rebase.

## Decision

- The project key is a hash of the absolute path of the git common directory (`git rev-parse --git-common-dir`). All worktrees of one repository have the same key. Outside git, the key is a hash of the project root.
- `<cache>/projects/<key>/counters.json` keeps the last allocated number for each kind (`adr`, `task`).
- To allocate a number, the server:
  1. Locks `<cache>/projects/<key>/lock` with an OS file lock (`flock` on macOS and Linux, `LockFileEx` on Windows).
  2. Calculates `max(counter, highest number in the current checkout) + 1`.
  3. Creates the file with exclusive create (`O_EXCL`).
  4. Writes the counter and releases the lock.
- One server also uses a mutex for its own concurrent calls.
- `adr_list` and `task_list` show a warning when two files have the same number. `task_renumber` fixes it.
- Docs lock (TASK-047): each tool that changes a task, ADR, learning, the glossary or a README index holds an exclusive OS lock on `<cache>/projects/<key>/docs.lock` and a mutex of the process for the full read, change and write. Two sessions then do not lose an edit. Tools that only read do not take the lock.
- A create takes the docs lock first, then the numbering lock, so the lock order is always the same and no deadlock is possible.
- Each document write goes to a temporary file in the same directory, which is then renamed to the file (`docfile.Write`). A reader sees the old or the new content, never a partial file. On Windows, Go opens files with no `FILE_SHARE_DELETE`, so the rename tries again for 1 second while a reader has the file open.

## Alternatives
- Search all worktrees and branches for the highest number: rejected. The counter in the cache gives the same result on one machine with less work.
- Date IDs (`TASK-20261025-1432`): rejected. They are hard to read.

## Consequences
- Positive: no duplicate numbers on one machine.
- Negative: numbers can have gaps (a task in a deleted branch). A teammate with an unpushed branch can still use the same number. The warning and `task_renumber` handle this.
- Follow-up: none.

## Evidence
- Design interview (2026-10-25): Q-44.
