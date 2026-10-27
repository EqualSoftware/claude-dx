# ADR-014: A lock for the plugin directory repair

- Status: Accepted
- Date: 2026-10-27
- Owners: Equal

## Context
claude-dx extracts the dx plugin into `<cache>/plugin/<DirName>`. `plugin.Ensure` runs under the shared cache lock (ADR-006), so many claude-dx processes can run it at the same time. If a file of the plugin directory is missing, `Ensure` must repair the directory. Two processes that repair at the same time can move aside or delete the good directory of the other process. A running Claude session then loses its `--plugin-dir` for a short time. ADR-006 lists the locks of the project and their order, but it does not cover this case. ADR-009 does not let us change ADR-006 after a push.

## Decision
Add a third OS lock: an exclusive lock on the file `<cache>/plugin.repair.lock`, with the `internal/filelock` package.

- The normal path does not use this lock. When the plugin directory has all 4 required files (`.claude-plugin/plugin.json`, `.mcp.json`, `hooks/hooks.json`, `rules.md`), `Ensure` uses it with no lock.
- The repair path takes the lock. Then it checks the directory again. If the directory is complete, it releases the lock and uses the directory. If not, it moves the bad directory aside, renames the new extracted directory into its place, and releases the lock.
- The repair never deletes a directory before the rename. The cleanup of the moved-aside directories stays in `removeUnused`, under the exclusive cache lock.
- Lock order: first the cache lock (shared), then the repair lock (exclusive). No code takes the cache lock while it holds the repair lock.
- The lock file is outside `<cache>/plugin/`, so the cleanup and `claude-dx clean` do not remove it.

## Alternatives
- Take the exclusive cache lock for the repair: Rejected. Other sessions hold the shared cache lock while they run, so the repair would wait until all sessions stop.
- Re-check the directory before the move, with no lock: Rejected. It makes the race window smaller, but it does not close it.
- Delete the bad directory, then extract again: Rejected. A second process can delete the good directory that the first process just put in place.

## Consequences
- Positive: No caller of `Ensure` gets a missing or incomplete plugin directory after `Ensure` returns. The normal start has no extra lock.
- Negative: The project has one more lock and one more lock order rule. On Windows, if a program holds a file in an incomplete plugin directory, the move aside fails and claude-dx stops with an error. The user must start it again. This case is rare and is accepted for 0.1.0.
- Follow-up: None.

## Evidence
- `internal/plugin/plugin.go`: `install` and `checkBinary` (branch fix-release).
- `internal/plugin/plugin_test.go`: `TestEnsureConcurrentRepair` runs 8 callers on an incomplete directory and passes with `-race -count=20`. With the lock or the re-check removed, the test fails (independent review, two mutations, `-count=100`).
- `internal/session/session.go:36-42`: `Ensure` runs after the shared cache lock. `removeUnused` runs only under the exclusive cache lock.
- ADR-006 gives the other locks and their order.
