# ADR-005: One MCP tool for each operation, with compact text results

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The objective of claude-dx is fewer tokens, better quality and the same result each time. Each tool definition and each tool result uses context.

## Decision
- One tool for each operation, with the name `<area>_<verb>`. Examples: `adr_create`, `task_set_status`, `git_commit`.
- Results are compact text for the model, not JSON. A list has one line for each item. Example: `ADR-001 Accepted  Load dx with launch flags`.
- Errors and messages use ASD-STE100.
- The project root is the git top level (`git rev-parse --show-toplevel`). If the directory is not in a git repository, the root is the current directory.
- Each tool refuses paths outside the project root.
- The launcher passes `--settings` with permission rules for the session. These tools are allowed with no prompt: read-only tools (`*_list`, `*_get`, `glossary_lookup`, `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`) and document tools (`adr_*`, `task_*`, `learning_*`, `glossary_*`). Git tools that change state ask the user, unless the user allows them in Claude Code.

## Alternatives
- One tool for each area with an `op` parameter: rejected. The schemas become unclear and permissions apply to all operations of the area. Claude Code defers MCP tool schemas when there are many tools, so the cost of many tools is small.
- JSON results: rejected. JSON repeats keys for each item.

## Consequences
- Positive: clear schemas, permission for each tool, small results.
- Negative: many tools. Each tool needs a good description.
- Follow-up: TASK-002 verifies the permission rule names of plugin MCP tools.

## Evidence
- Design interview (2026-10-25): Q-27, Q-28, Q-29, Q-30.
