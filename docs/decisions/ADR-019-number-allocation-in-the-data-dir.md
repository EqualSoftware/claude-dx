# ADR-019: Number allocation in the data dir

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context

Replaces ADR-007.

ADR-007 keeps the counters and a numbering lock in `<cache>/projects/<key>/`. Since then: TASK-059 removed the numbering lock, because the docs lock protects the counter (ADR-015); ADR-015 moved the docs lock to the project data directory; ADR-016 removes the task numbers; ADR-018 puts the counters in the Project file (data). `cache clean` also removed the counters of git projects, so two worktrees could get the same number. ADR-007 is pushed, so it does not change (ADR-017).

## Decision
- The project key is a hash of the absolute path of the git common directory (`git rev-parse --git-common-dir`). All worktrees of one repository have the same key. Outside git, the key is a hash of the project root.
- The counters are in `project.json` in the project data directory (ADR-015): `{"counters": {"adr": 18}}`. In git this is `<git-common-dir>/dx/project.json`, so all worktrees share it and `cache clean` does not remove it. Tasks have numbers until TASK-064; then only ADRs have numbers (ADR-016).
- To allocate a number, the caller holds the exclusive docs lock (ADR-015). Then `numbering.Allocator.Next`:
  1. Calculates `max(counter, highest number in the current checkout) + 1`.
  2. Creates the file with exclusive create (`O_EXCL`).
  3. Writes the counter with `docfile.Write`.
- `adr_list` and `task_list` show a warning when two files have the same number. `task_renumber` fixes it.
- Each document write goes to a temporary file in the same directory, which is then renamed to the file (`docfile.Write`). The file keeps its mode. A reader sees the old or the new content, never a partial file. On Windows, Go opens files with no `FILE_SHARE_DELETE`, so the rename tries again for 1 second while a reader has the file open.

## Alternatives
- Change ADR-007: ADR-007 is pushed, so it does not change (ADR-017).
- Search all worktrees and branches for the highest number: The shared counter gives the same result on one machine with less work (ADR-007).
- Date IDs (ADR-20261025-1432): They are hard to read (ADR-007).

## Consequences
- Positive: No duplicate numbers on one machine, also after cache clean. One lock protects the counter and the documents, with no lock order rule.
- Negative: Numbers can have gaps. A teammate with an unpushed branch can still use the same number; the warning and task_renumber handle this. A counter in the cache of an old build is not read; the highest number in the checkout replaces it.
- Follow-up: None.

## Evidence
- `internal/numbering/numbering.go`: `project.json` with `counters`, written with `docfile.Write` (TASK-068).
- `internal/project/data.go`: `DataDir` (TASK-068).
- TASK-059 (commit 4929598): the numbering lock is removed.
