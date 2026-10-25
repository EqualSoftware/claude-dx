# ADR-003: Pass arguments through to Claude Code, with a small set of dx commands

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
- macOS and Linux: the launcher replaces its process with `claude` (`syscall.Exec`). Signals, TTY and exit code are the same as a direct start.
- Windows: the launcher starts `claude` as a child process with the same stdio, forwards Ctrl+C and returns the same exit code.
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
