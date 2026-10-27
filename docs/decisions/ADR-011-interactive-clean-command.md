# ADR-011: Interactive clean command

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
The cache of claude-dx has old plugin versions and, for projects outside git, the local tasks. The data directory `<git-common-dir>/dx/` has the local tasks of each git project (ADR-007). A full delete of all data can lose work of the user. The cache folder of a project has a hash as its name, so a person cannot see which project it is. Global items need the exclusive cache lock, so they cannot be removed while a session runs (ADR-006).

## Decision
- The command `claude-dx clean` shows an interactive multi-select with the `charmbracelet/huh` library, then asks for a confirmation. Nothing is deleted without a selection and a confirmation of the user. The default answer is No.
- Items:
  - Global: old plugin versions, and each LSP server when the LSP servers exist (ADR-004).
  - Projects: for each known project, its cache data and its local tasks (by spec and status). A project that is not on the disk is shown as missing.
- Lock files are never shown and never deleted (ADR-006). The number counters of a project that exists are not items (ADR-007).
- Each item has its lock. Global items need the exclusive cache lock: while a session runs, they are disabled with the reason. Project items take the shared cache lock, then the docs lock of that project (ADR-006), so a person can clean one project while a session runs. `MoveData` holds the exclusive cache lock, so a project delete never runs during a move.
- `claude-dx clean --list` prints the same tree with sizes and status and deletes nothing.
- With no interactive terminal, `claude-dx clean` refuses. Only `--list` works. There are no flags that delete with no selection.
- The code is in `internal/clean`. The project file of the cache (ADR-007) gives the path and the last use date of each project.

## Alternatives
- A numbered list with no new dependency: Rejected. The user wants an interactive and good-looking selection.
- Flags `--only`, `--all` and `--yes` for scripts: Rejected. A script must never delete user data.
- A separate command for the data directory in `.git/dx`: Rejected. One command shows all data of claude-dx in one place.
- One cache lock for all items: Rejected. Nothing can be cleaned while any session runs.

## Consequences
- Positive: the user sees all data of claude-dx, with sizes, and selects what to delete. No work is lost by accident. Missing projects are easy to find and clean.
- Negative: new dependency `charmbracelet/huh` and its terminal libraries. Terminals that cannot run it (possibly Git Bash on Windows) cannot clean, only list.
- Follow-up: test the selection by hand in Windows Terminal, PowerShell and Git Bash, and on macOS and Linux.

## Evidence
- internal/clean/form.go:12-14 (`Interactive`), :18-51 (multi-select with `huh`; a note lists the disabled items), :54-58 (confirm).
- internal/clean/run.go:20-51 (select, confirm, delete), :92-104 (`List`).
- internal/clean/scan.go:81-105 (global item, disabled while a session runs), :108-118 (exclusive try before the delete), :183-200 (project groups, missing projects), :203-214 (`shared`), :230-232 and :311-312 (shared cache lock, then the docs lock of the project).
- cmd/claude-dx/main.go:113-127 (`--list`, refuses with no terminal), :152-154 (terminal check).
- cmd/claude-dx/main_test.go:273: the help text has `claude-dx clean --list`.
