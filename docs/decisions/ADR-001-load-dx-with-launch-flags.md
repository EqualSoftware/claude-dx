# ADR-001: Load dx with launch flags, not with project configuration

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
claude-dx adds skills, an agent, an MCP server, hooks and rules to Claude Code. These additions must not change the configuration of the user or of the project. The skills must use the `/dx:` prefix. Claude Code gives the `<plugin>:` prefix only to plugin skills. Claude Code 2.1.308 has the flags `--plugin-dir`, `--mcp-config`, `--settings` and `--append-system-prompt`.

## Decision
- The binary embeds the `dx` plugin (`go:embed`): skills, the `scout` agent, hooks and the MCP server declaration.
- At start, the launcher extracts the plugin to `<cache>/plugin/<version>/`, if it is not there. It writes the absolute path of the binary into the MCP and hook declarations of the extracted plugin.
- The launcher starts Claude Code with `--plugin-dir <cache>/plugin/<version>`, `--settings <json>` (permissions, see ADR-005) and `--append-system-prompt <dx rules>` (see ADR-009).
- The launcher writes no file in the project and no file in the Claude Code configuration of the user.
- `<cache>` is `os.UserCacheDir()/claude-dx`: `~/Library/Caches/claude-dx` (macOS), `~/.cache/claude-dx` (Linux), `%LocalAppData%\claude-dx` (Windows).
- `claude-dx cache clean` removes `<cache>`. It does not touch the docs in the project.
- Only the documents that the tools make (`docs/decisions/`, `docs/tasks/`, `docs/learnings/`, `docs/GLOSSARY.md`) go into the project. The project keeps them in git.

## Alternatives
- Write `.mcp.json`, `.claude/skills/dx-*` and `.claude/settings.local.json` with a manifest for removal: rejected. It changes files of the user, the skill prefix becomes `/dx-`, and removal must merge JSON files.
- Extract the plugin into `.claude/dx/` in the project: rejected. It adds files to each project and a line to `.gitignore`. The cache gives the same result with no project files.

## Consequences
- Positive: no change to user or project configuration. Removal is `cache clean` plus deletion of the binary. Native `/dx:` prefix.
- Negative: `claude` without the launcher does not load dx.
- Follow-up: TASK-002 verifies the tool names and skill prefix of a plugin loaded with `--plugin-dir`.

## Evidence
- `claude --help` of Claude Code 2.1.308 lists `--plugin-dir`, `--mcp-config`, `--settings`, `--append-system-prompt`.
- Design interview (2026-10-25): Q-1, Q-2.
