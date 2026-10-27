# ADR-007: Project data dir and number allocation

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
Many agents can create ADRs at the same time: subagents of one session, two sessions, worktrees, and branches. Two ADRs with the same number cause a conflict at rebase. The local tasks, the docs lock and the number counters must be per project, and not lost when the user cleans the cache. All worktrees of one repository must share them. The cache folder of a project has a hash as its name, so a person cannot see which project it is.

## Decision
- The project key is the first 16 hex characters of the SHA-256 of the absolute path of the git common directory (`git rev-parse --git-common-dir`). All worktrees of one repository have the same key. Outside git, the key is a hash of the project root.
- The project data directory is `<git-common-dir>/dx/` in a git repository: git does not track it, and all worktrees share it. Outside git it is `<cache>/projects/<key>/`.
- The data directory has `docs.lock` (ADR-006), `tasks/` (ADR-008) and `project.json`.
- `project.json` in the data directory keeps the number counters: `{"counters": {"adr": 11}}`. A write changes only the key `counters` and keeps the other keys.
- A second file `<cache>/projects/<key>/project.json` keeps the project path and the last use date (`path`, `last_used`), so `claude-dx clean` can show each project (ADR-011). The MCP server writes it at start, as best effort. Outside git, both files are the same file, so each writer keeps the keys of the other.
- All writes of `project.json` hold the docs lock of the project and are atomic (ADR-006).
- When the project becomes a git repository, the launcher moves `tasks/` and `project.json` from the cache folder to `<git-common-dir>/dx/` at start, only while it holds the exclusive cache lock, so no other claude-dx process uses the old location. The old `docs.lock` stays. Until the move, each process uses the old location if `<git-common-dir>/dx/` does not exist and the old location has data, so all processes use the same docs lock.
- Only ADRs have numbers. To allocate a number, the caller holds the exclusive docs lock. Then `numbering.Allocator.Next`:
  1. Calculates `max(counter, highest number in the current checkout) + 1`.
  2. Creates the file with exclusive create (`O_EXCL`).
  3. Writes the counter, only if the create succeeded.
- `adr_list` shows a warning when two files have the same number.

## Alternatives
- Search all worktrees and branches for the highest number: Rejected. The shared counter gives the same result on one machine with less work.
- Date IDs: Rejected. They are hard to read.
- Keep the counters in the cache: Rejected. `clean` removes them, and two worktrees can get the same number.
- A separate numbering lock: Rejected. The docs lock protects the counter and the documents, with no extra lock.

## Consequences
- Positive: no duplicate numbers on one machine, also after `clean`. All worktrees share the data. One docs lock protects the counter and the documents.
- Negative: numbers can have gaps. A teammate with an unpushed branch can still use the same number, and the warning shows it. `project.json` has two writers, so both must hold the docs lock.
- Follow-up: none.

## Evidence
- internal/numbering/numbering.go:16-43 (`Allocator`, `Next`), :67-69 (`project.json`), :73-98 (read, other keys kept), :102-113 (write of `counters` only).
- internal/project/project.go:24-38 (`Open`, key and common directory), :72-75 (key).
- internal/project/data.go:26-34 (`DataDir`), :40-66 (`MoveData`), :70-82 (pending move).
- internal/launcher/launcher.go:41-45 and :68-77 (move at start, while no other process runs).
- internal/mcpserver/server.go:38-41 (data directory and docs lock), :103-122 (project file of the cache).
- internal/docs/adr/store.go:96 (duplicate warning), :153 (`Alloc.Next` for `adr`).
- internal/numbering/numbering_test.go:138-156: the counter file of an old build is not read.
