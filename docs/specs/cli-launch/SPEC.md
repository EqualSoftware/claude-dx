---
slug: cli-launch
title: Launch command of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [cli, launcher]
related: [ADR-001, ADR-005, ADR-006, ADR-012]
---

# Launch command of claude-dx

## 1. Summary

claude-dx with no dx command starts claude with the dx plugin. The launcher extracts the plugin, builds the flags, runs claude as a child process, and returns its exit code. It also prints a notice when a newer release exists.

## 2. Problem

The developer must start Claude Code with dx in one command, with no change to the project or user configuration (ADR-001). The session keeps the terminal, the signals and the exit code of claude. The locks live as long as the session (ADR-006).

## 3. Goals

- G-1: `claude-dx <args>` starts claude with the dx flags and the user arguments, with the same terminal, signals and exit code.

## 4. Out of Scope

- The parser, the help and the version: see the spec cli-common.
- The deletion of the file .old and the update command: see the spec cli-update.
- The cache lock and the removal of old plugins: see the spec session.
- The plugin extraction: see the spec plugin.
- The background update check: see the spec mcp-server.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | claude-dx starts claude with the dx flags and the user arguments, with the same terminal, environment and exit code. | Done |
| R-2 | If claude is not in PATH, the launcher prints the install message and exits with code 1. | Done |
| R-3 | On Unix, the launcher runs claude as a child, forwards SIGTERM and SIGHUP to it, and drops SIGINT and SIGQUIT. | Done |
| R-4 | On all platforms, the launcher returns the exit code of claude, or 128 plus the signal number when a signal ended claude. | Done |
| R-5 | At start, the launcher reads update.json and prints one line to stderr when a newer release exists. The launcher makes no network call. | Done |

## 6. Design

### Modules

- `internal/launcher/launcher.go`: runs the update notice, extracts the plugin, and builds the arguments of claude. It calls `session.Start` and then the start function.
- `internal/launcher/start.go`: starts claude as a child process with `exec.Command`. It uses `signal.Notify` for the ignored signals and one channel for each forwarded signal. `signals_unix.go` and `signals_windows.go` give the signal lists. Windows forwards no signal (ADR-006).
- `internal/update/update.go`: `Notice` reads `<cache>/update.json`.

### Data flow

1. `main` gives the arguments to `launcher.Run`.
2. The launcher reads `update.json` for the notice, extracts the plugin (see plugin), and calls `start`.
3. `start` runs claude as a child process. The exit code of claude becomes the exit code of claude-dx.

### Formats

- `<cache>/update.json`: `{"checked_at": "2026-10-25T10:00:00Z", "latest": "v0.2.0"}`.

### Decisions

- ADR-001: launch with flags and the passthrough of arguments.
- ADR-006: claude runs as a child of the launcher, so the locks live as long as the session.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestLaunchPassesArgsAndExitCode` (`cmd/claude-dx/main_test.go`), with a fake claude.
- R-2: `TestLaunchWithoutClaude` (`cmd/claude-dx/main_test.go`).
- R-3: `TestStartForwardsSIGTERM`, `TestStartIgnoresSIGINT` (`internal/launcher/start_unix_test.go`).
- R-4: `TestStartExitCode` (`internal/launcher/start_test.go`) and `TestStartSignalExitCode` (`internal/launcher/start_unix_test.go`).
- R-5: `TestLaunchShowsUpdateNotice` (`cmd/claude-dx/main_test.go`).

All tests pass on macOS. Linux and Windows run in CI.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification. Split from cli. |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |