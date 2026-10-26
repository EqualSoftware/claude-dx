# claude-dx

[![CI](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml/badge.svg)](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml)

`claude-dx` starts [Claude Code](https://docs.claude.com/claude-code) with the dx tools: skills, an MCP server and hooks for ADRs, tasks, learnings, a glossary, git, files and LSP servers. The tools do the mechanical work, so the model uses fewer tokens and gives the same result each time.

`claude-dx` writes no configuration into your project or into your Claude Code settings. It loads its plugin with launch flags.

> Status: in development. See [docs/tasks/](docs/tasks/).

## Why

An AI agent spends many tool calls and tokens on mechanical work: read a template, find the next number, write a file, update an index, adapt a shell command to the platform. Each call sends the full context again, and each run can give a different result.

claude-dx moves this work into Go tools:

- **Fewer tokens.** One tool call does what took many. Results are compact text.
- **Consistency.** The same operation gives the same result each time, on macOS, Linux and Windows.
- **Safe defaults.** Rebase, Conventional Commits, no force push to the main branch without your direct approval.
- **Clear documents.** ADRs, tasks, learnings and the glossary use [ASD-STE100](https://www.asd-ste100.org/) Simplified Technical English. You talk to the agent in your language.

## Roadmap

Each task has a `milestone` field. The [task index](docs/tasks/) has one table for each milestone.

| Milestone | Content |
| :--- | :--- |
| v0.1.0 | Launcher, `update`, `cache clean`, `doctor`. Tools and skills for ADRs, tasks, learnings and the glossary. `/dx:grill` and the `scout` agent. |
| v0.2.0 | Git tools, the Bash git block, `git_exec`, `report`. `/dx:commit` and `/dx:eureka`. |
| v0.3.0 | Filesystem tools. |
| v0.4.0 | LSP servers with a warm cache for Go, TypeScript, JavaScript, Python, C and C++. No install by the user. |
| v0.5.0 | LSP servers for HTML, CSS, JSON, Bash, CMake, Lua, Make and Ninja. |

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

Each release file has a GitHub build provenance attestation. It proves that the release workflow of this repository made the file. To verify a file that you downloaded, use the [GitHub CLI](https://cli.github.com/):

```sh
gh attestation verify claude-dx_<version>_<os>_<arch>.tar.gz --repo EqualSoftware/claude-dx
```

## Use

```sh
claude-dx [claude arguments]     # Start Claude Code with dx.
claude-dx -- [claude arguments]  # Pass all arguments to Claude Code.
claude-dx update                 # Install the latest release.
claude-dx cache clean            # Remove all data of claude-dx but the lock files.
claude-dx doctor                 # Check the setup.
```

The dx server checks for a new release one time in 24 hours, in the background. When a new release exists, `claude-dx` shows one line at start. It never updates without `claude-dx update`.

Each `claude-dx` process holds a shared lock on `cache.lock` in the cache. At start, if no other process uses the cache, `claude-dx` removes the plugin files of old versions. `claude-dx cache clean` removes nothing while a session runs: close all `claude-dx` sessions first. It keeps the lock files.

claude-dx gives the dx rules to Claude Code with `--append-system-prompt-file`. Your `--append-system-prompt` also applies. Do not give `--append-system-prompt-file`: Claude Code uses only the last one, so the dx rules stop.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md). To report a security problem, read [SECURITY.md](SECURITY.md). This project has a [Code of Conduct](CODE_OF_CONDUCT.md).

## Design

- Decisions: [docs/decisions/](docs/decisions/)
- Tasks: [docs/tasks/](docs/tasks/)
- Learnings: [docs/learnings/](docs/learnings/)
- Terms: [docs/GLOSSARY.md](docs/GLOSSARY.md)

## Credits

The `dx:grill` skill is based on the [grilling skill](https://github.com/mattpocock/skills/blob/main/skills/productivity/grilling/SKILL.md) of [mattpocock/skills](https://github.com/mattpocock/skills). Copyright (c) 2026 Matt Pocock. MIT License.

## License

[MIT](LICENSE)
