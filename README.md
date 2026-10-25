# claude-dx

[![CI](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml/badge.svg)](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml)

`claude-dx` starts [Claude Code](https://docs.claude.com/claude-code) with the dx tools: skills, an MCP server and hooks for ADRs, tasks, learnings, a glossary, git, files and LSP servers. The tools do the mechanical work, so the model uses fewer tokens and gives the same result each time.

`claude-dx` writes no configuration into your project or into your Claude Code settings. It loads its plugin with launch flags.

> Status: in development. See [docs/tasks/INDEX.md](docs/tasks/INDEX.md).

## Why

An AI agent spends many tool calls and tokens on mechanical work: read a template, find the next number, write a file, update an index, adapt a shell command to the platform. Each call sends the full context again, and each run can give a different result.

claude-dx moves this work into Go tools:

- **Fewer tokens.** One tool call does what took many. Results are compact text.
- **Consistency.** The same operation gives the same result each time, on macOS, Linux and Windows.
- **Safe defaults.** Rebase, Conventional Commits, no force push to the main branch without your direct approval.
- **Clear documents.** ADRs, tasks, learnings and the glossary use [ASD-STE100](https://www.asd-ste100.org/) Simplified Technical English. You talk to the agent in your language.

## Roadmap

| Version | Content | Tasks |
| :--- | :--- | :--- |
| v0.1 | Launcher, `update`, `cache clean`, `doctor`. Tools and skills for ADRs, tasks, learnings and the glossary. `/dx:grill` and the `scout` agent. | [TASK-001](docs/tasks/TASK-001-project-skeleton-ci-and-release.md) to [TASK-008](docs/tasks/TASK-008-grill-skill-scout-agent-and-rules.md) |
| v0.2 | Git tools, the Bash git block, `git_exec`, `report`. `/dx:commit` and `/dx:eureka`. | [TASK-009](docs/tasks/TASK-009-git-runner-and-read-tools.md) to [TASK-013](docs/tasks/TASK-013-commit-and-eureka-skills.md) |
| v0.3 | Filesystem tools. | [TASK-014](docs/tasks/TASK-014-filesystem-tools.md) |
| v0.4 | LSP servers with a warm cache for Go, TypeScript, JavaScript, Python, C and C++. No install by the user. | [TASK-015](docs/tasks/TASK-015-lsp-core-languages.md) |
| v0.5 | LSP servers for HTML, CSS, JSON, Bash, CMake, Lua, Make and Ninja. | [TASK-016](docs/tasks/TASK-016-lsp-other-languages.md) |

## Install

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/EqualSoftware/claude-dx/main/scripts/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/EqualSoftware/claude-dx/main/scripts/install.ps1 | iex
```

The scripts verify the SHA-256 checksum of the release. To change the install directory, set `CLAUDE_DX_INSTALL_DIR`.

## Use

```sh
claude-dx [claude arguments]     # Start Claude Code with dx.
claude-dx -- [claude arguments]  # Pass all arguments to Claude Code.
```

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md). To report a security problem, read [SECURITY.md](SECURITY.md). This project has a [Code of Conduct](CODE_OF_CONDUCT.md).

## Design

- Decisions: [docs/decisions/](docs/decisions/)
- Tasks: [docs/tasks/INDEX.md](docs/tasks/INDEX.md)
- Terms: [docs/GLOSSARY.md](docs/GLOSSARY.md)

## License

[MIT](LICENSE)
