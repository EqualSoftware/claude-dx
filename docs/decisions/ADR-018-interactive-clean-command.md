# ADR-018: Interactive clean command

- Status: Proposed
- Date: 2026-10-26
- Owners: Maintainer

## Context
`cache clean` removes all of the cache (`internal/cache/cache.go:24`). After ADR-016, the cache also has the local tasks of projects outside git, and `<git-common-dir>/dx/` has the local tasks of each git project. A full delete can lose work of the user. The cache folder of a project has a hash as its name (ADR-007), so a person cannot see which project it is. `cache clean` also needs the exclusive cache lock, so it cannot remove anything while a session runs.

## Decision
- The command `claude-dx clean` replaces `cache clean`. It shows an interactive multi-select with the `charmbracelet/huh` library, then asks for a confirmation. Nothing is deleted without a selection and a confirmation of the user.
- Items:
  - Global: old plugin versions, and each LSP server.
  - Projects: for each known project, its cache data and its local tasks (by spec and status). A project that is not on the disk is shown as missing.
- Lock files are never shown and never deleted (ADR-015).
- Each item has its lock. Global items need the exclusive cache lock: while a session runs, they are disabled with the reason. Project items use the docs lock or the task lock of that project, so a person can clean one project while another project has a session.
- `claude-dx clean --list` prints the same tree with sizes and status and deletes nothing.
- With no interactive terminal, `claude-dx clean` refuses. Only `--list` works. There are no flags that delete with no selection.
- Each project has `<cache>/projects/<key>/project.json`. It contains the project path, the last use date and the number counters that `counters.json` has now. The MCP server writes the path and the date at start. All writes hold the docs lock and are atomic.

## Alternatives
- Numbered list with no new dependency: The user wants an interactive and good-looking selection.
- Flags --only, --all and --yes for scripts: A script must never delete user data. The user does not want this capability.
- A separate command for .git/dx: One command shows all data of claude-dx in one place.
- One cache lock for all items: Nothing can be cleaned while any session runs.
- Keep cache clean as an alias: v0.1.0 is not released, so no user depends on it.

## Consequences
- Positive: The user sees all data of claude-dx, with sizes, and selects what to delete. No work is lost by accident. Missing projects are easy to find and clean.
- Negative: New dependency `charmbracelet/huh` and its terminal libraries. Terminals that cannot run it (possibly Git Bash on Windows) cannot clean, only list. project.json has two writers, so both must hold the docs lock.
- Follow-up: Test the selection by hand in Windows Terminal, PowerShell and Git Bash, and on macOS and Linux.

## Evidence
- `internal/cache/cache.go:24`: `Clean` removes the cache directory.
- `internal/numbering/numbering.go:79`: the counters are in `counters.json`.
- ADR-007: the project folder name is a hash of the git common directory or the project root.
