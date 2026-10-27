# ADR-001: Launcher and CLI

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
claude-dx adds skills, rules and an MCP server to Claude Code. These additions must not change the configuration of the user or of the project. The skills must use the `/dx:` prefix, and Claude Code gives the `<plugin>:` prefix only to plugin skills. claude-dx is a proxy for `claude`, and some dx commands have the same name as `claude` commands (`--help`, `--version`, `update`, `doctor`). Claude Code 2.1.308 has the flags `--plugin-dir`, `--settings` and `--append-system-prompt-file`.

## Decision
- The binary embeds the `dx` plugin (`go:embed`): skills, the rules file and the MCP server declaration.
- At start, the launcher extracts the plugin to `<cache>/plugin/<version>-<key>/`, if it is not there. `<key>` is the first 8 bytes of the SHA-256 of the absolute binary path, in hex. A `dev` build also puts the embedded files in the key. The launcher writes the absolute path of the binary into the JSON files of the extracted plugin. So each binary path has its own plugin directory.
- The launcher starts Claude Code with `--plugin-dir <cache>/plugin/<version>-<key>`, `--settings <json>` and `--append-system-prompt-file <plugin>/rules.md`, then the arguments of the user. The settings allow all tools of the dx MCP server with no prompt.
- The launcher writes no file in the project and no file in the Claude Code configuration of the user.
- `<cache>` is `os.UserCacheDir()/claude-dx`: `~/Library/Caches/claude-dx` (macOS), `~/.cache/claude-dx` (Linux), `%LocalAppData%\claude-dx` (Windows).
- Only the documents that the dx tools make (`docs/decisions/`, `docs/specs/`, `docs/learnings/`, `docs/GLOSSARY.md`) go into the project. The project keeps them in git.
- dx owns these commands: `--help`, `-h`, `--version`, `update`, `clean`, `doctor`, and the internal command `mcp`. The names `report` and `hook` are reserved for the git tools. The parser reads only the first argument.
- The launcher passes all other arguments to `claude` with no change.
- `--` stops dx parsing. The launcher passes all arguments after `--` to `claude`. Example: `claude-dx -- --help` shows the help of `claude`.
- The parser uses only the Go standard library.
- If `claude` is not in `PATH`, the launcher stops with an error and the install URL.
- The launcher runs `claude` as a child process (ADR-006).

## Alternatives
- Write `.mcp.json`, `.claude/skills/dx-*` and `.claude/settings.local.json`, with a manifest for removal: Rejected. It changes files of the user, the skill prefix becomes `/dx-`, and removal must merge JSON files.
- Extract the plugin into `.claude/dx/` in the project: Rejected. It adds files to each project and a line to `.gitignore`. The cache gives the same result with no project files.
- Strict mode, with all `claude` arguments only after `--`: Rejected. Daily use (`claude-dx --resume`) becomes longer.
- A `dx` namespace (`claude-dx dx doctor`): Rejected. `--` solves the conflicts with less typing.
- cobra or urfave/cli: Rejected. They do not pass unknown flags through easily.

## Consequences
- Positive: no change to user or project configuration. Native `/dx:` prefix. `claude-dx` replaces `claude` in daily use.
- Negative: `claude` without the launcher does not load dx. A future `claude` command with a dx name needs `--`.
- Follow-up: none.

## Evidence
- internal/plugin/plugin.go:19 (embed), :25-44 (extract), :49-65 (plugin directory name), :78-97 (binary path in JSON files).
- internal/launcher/launcher.go:25-62 (start), :81 (settings), :86-93 (flags), :57-60 (claude not found).
- internal/cache/cache.go:10-16 (cache directory).
- internal/cli/cli.go:12-23 (dx commands), :25-39 (parse, `--`).
- cmd/claude-dx/main.go:26-39 (help text), :46-68 (command dispatch).
- internal/doctor/doctor.go:16 (`MinClaude = "2.1.308"`, the oldest Claude Code version on which the flags are verified).
