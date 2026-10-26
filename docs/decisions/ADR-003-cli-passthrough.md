# ADR-003: Pass arguments through to Claude Code

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
`claude-dx` is a proxy for `claude`. Some dx commands have the same name as `claude` commands (`--help`, `--version`, `update`, `doctor`).

## Decision

- dx owns these commands: `--help`, `--version`, `update`, `cache clean`, `doctor`, `report`, and the internal commands `mcp` and `hook`.
- The launcher passes all other arguments to `claude` with no change.
- `--` stops dx parsing. The launcher passes all arguments after `--` to `claude`. Example: `claude-dx -- --help` shows the help of `claude`.
- The parser uses only the Go standard library.
- All platforms: the launcher starts `claude` as a child process with the same stdio and environment, waits for it and returns its exit code. If a signal ended `claude`, the launcher returns 128 plus the signal number. The launcher is the parent of `claude` because its locks must live as long as the session (ADR-015, TASK-060).
- Signals: the terminal sends Ctrl+C (and SIGQUIT on Unix) to `claude` too, so the launcher catches these signals and does nothing. On Unix, the launcher sends SIGTERM and SIGHUP that it gets to `claude`. The launcher catches signals and does not ignore them, because an ignored signal stays ignored in `claude` after exec.
- If `claude` is not in `PATH`, the launcher stops with an error and the install URL.

## Alternatives
- Strict mode: all `claude` arguments only after `--`: rejected. Daily use (`claude-dx --resume`) becomes longer.
- A `dx` namespace (`claude-dx dx uninstall`): rejected. `--` solves the conflicts with less typing.
- cobra or urfave/cli: rejected. They do not pass unknown flags through easily.

## Consequences
- Positive: `claude-dx` replaces `claude` in daily use.
- Negative: a future `claude` command with a dx name needs `--`.
- Follow-up: `--help` tells the user about `claude-dx -- --help`.

## Evidence
- Design interview (2026-10-25): Q-9, Q-16, Q-40, Q-41.
