---
slug: cli-common
title: Command parser of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [cli]
related: [ADR-001, ADR-012]
---

# Command parser of claude-dx

## 1. Summary

claude-dx parses its own commands in one place. The first argument selects a dx command. Other arguments go to claude. The parser also gives the help text and the version. This spec is the entry of the family of the cli-* specs, one for each command.

## 2. Problem

The developer must start Claude Code with dx in one command, with no change to the project or user configuration (ADR-001). The arguments of claude must pass through, and a dx command must not take an argument that is meant for claude.

## 3. Goals

- G-1: Only the first argument selects a dx command, and `--` passes all later arguments to claude.
- G-2: `claude-dx --help` shows the dx help and does not start claude.

## 4. Out of Scope

This spec is the scope of the family of the cli-* specs (ADR-012). Each command has its own spec:
- The start of claude: see the spec cli-launch.
- The update command: see the spec cli-update.
- The doctor command: see the spec cli-doctor.
- The clean command: see the spec cli-clean.
- The MCP server (command mcp): see the spec mcp-server.
- The report and hook commands: see the spec git-tools.
- The resume command: see the spec resume-command.
- The cache lock and the plugin extraction: see the specs session and plugin.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Only the first argument selects a dx command (--help, -h, --version, update, clean, doctor, report, mcp, hook). The argument -- passes all later arguments to claude. | Done |
| R-2 | claude-dx --help and claude-dx --version print the dx help or the version and do not start claude. | Done |

## 6. Design

### Modules

- `internal/cli/cli.go`: the parser. It maps the first argument to a dx command. Other arguments go to claude.
- `cmd/claude-dx/main.go`: runs the command that the parser gives. It prints the help and the version.

### Data flow

1. `main` calls `cli.Parse`.
2. A dx command runs its own code (see the spec of the command).
3. Otherwise the arguments go to the launcher (see cli-launch).

### Decisions

- ADR-001: launch with flags, no change to the configuration, and the passthrough of arguments.
- ADR-012: one subject for each spec.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestParse` (`internal/cli/cli_test.go`).
- R-2: `TestRunHelpDoesNotStartClaude` and `TestRunVersion` (`cmd/claude-dx/main_test.go`).

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
| 2026-10-27 | Equal | R-1: text changed. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |