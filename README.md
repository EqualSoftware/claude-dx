# claude-dx

[![CI](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml/badge.svg)](https://github.com/EqualSoftware/claude-dx/actions/workflows/ci.yml)

`claude-dx` starts [Claude Code](https://docs.claude.com/claude-code) with the dx tools: skills and an MCP server for specs, ADRs, tasks, learnings and a glossary. The tools do the mechanical work, so the model uses fewer tokens and gives the same result each time.

`claude-dx` writes no configuration into your project or into your Claude Code settings. It loads its plugin with launch flags. It keeps the local data of a project (tasks and counters) in the git directory of the project, in `.git/dx/`.

> Status: in development. See [docs/specs/](docs/specs/README.md).

## Why

An AI agent spends many tool calls and tokens on mechanical work: read a template, find the next number, write a file, update an index, adapt a shell command to the platform. Each call sends the full context again, and each run can give a different result.

claude-dx moves this work into Go tools:

- **Fewer tokens.** One tool call does what took many. Results are compact text.
- **Consistency.** The same operation gives the same result each time, on macOS, Linux and Windows.
- **Safe defaults.** Rebase, Conventional Commits, no force push to the main branch without your direct approval.
- **Clear documents.** Specs, ADRs, tasks, learnings and the glossary use [ASD-STE100](https://www.asd-ste100.org/) Simplified Technical English. You talk to the agent in your language.

## Roadmap

Each spec has a `milestone` field. The [spec index](docs/specs/README.md) has one row for each spec, with its milestone.

| Milestone | Content |
| :--- | :--- |
| v0.1.0 | Launcher, `update`, `clean`, `doctor`. Tools and skills for specs, ADRs, tasks, learnings and the glossary. `/dx:grill`. |
| v0.2.0 | Git tools: safe defaults for commits, rebase and push. |
| v0.3.0 | Filesystem tools. |
| v0.4.0 | More skills, for example for commits and to record the lessons of a session. |

## Requirements

claude-dx works on macOS, Linux and Windows, on amd64 and arm64. It is one binary with no runtime. It needs:

| Tool | Version | Why |
| :--- | :--- | :--- |
| Claude Code | 2.1.308 or later | claude-dx starts it with the dx plugin |
| git | 2.30 or later | claude-dx keeps the project data in the git directory. The git tools use git |

## Install

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/EqualSoftware/claude-dx/main/scripts/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/EqualSoftware/claude-dx/main/scripts/install.ps1 | iex
```

The scripts verify the SHA-256 checksum of the release. They install to `~/.local/bin` on macOS and Linux, and to `%LocalAppData%\Programs\claude-dx` on Windows. To change the install directory, set `CLAUDE_DX_INSTALL_DIR`.

Each release file has a GitHub build provenance attestation. It proves that the release workflow of this repository made the file. To verify a file that you downloaded, use the [GitHub CLI](https://cli.github.com/). The version has no `v` (for example `0.1.0`). On Windows, the file is a `.zip`:

```sh
gh attestation verify claude-dx_<version>_<os>_<arch>.tar.gz --repo EqualSoftware/claude-dx
```

## Use

```sh
claude-dx [claude arguments]     # Start Claude Code with dx.
claude-dx -- [claude arguments]  # Pass all arguments to Claude Code.
claude-dx update                 # Install the latest release.
claude-dx clean                  # Select the data of claude-dx to delete.
claude-dx clean --list           # Show the data of claude-dx and its sizes.
claude-dx doctor                 # Check the setup.
claude-dx --version              # Show the version.
claude-dx --help                 # Show the commands of claude-dx.
```

`claude-dx clean` needs an interactive terminal. Use `claude-dx clean --list` in scripts.

The dx server checks for a new release one time in 24 hours, in the background. When a new release exists, `claude-dx` shows one line at start. It never updates without `claude-dx update`.

Each `claude-dx` process holds a shared lock on `cache.lock` in the cache. At start, if no other process uses the cache, `claude-dx` removes the plugin files of old versions. `claude-dx clean` asks you to select the data to delete and to confirm. While a session runs, it disables the global items. It never shows or deletes the lock files.

claude-dx gives the dx rules to Claude Code with `--append-system-prompt-file`. Your `--append-system-prompt` also applies. Do not give `--append-system-prompt-file`: Claude Code uses only the last one, so the dx rules stop.

## Data and uninstall

claude-dx keeps its data in two places:

- The cache: the plugin files and the global data. It is `claude-dx` in the cache directory of the system: `~/Library/Caches/claude-dx` on macOS, `~/.cache/claude-dx` on Linux, `%LocalAppData%\claude-dx` on Windows.
- The project data: `.git/dx/` in each git project. Outside git, it is `projects/<key>/` in the cache.

`claude-dx clean --list` shows each item and its size. To uninstall, delete the `claude-dx` binary from the install directory, the cache directory, and `.git/dx/` in your projects.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md). To report a security problem, read [SECURITY.md](SECURITY.md). This project has a [Code of Conduct](CODE_OF_CONDUCT.md).

## Design

- Decisions: [docs/decisions/](docs/decisions/README.md)
- Specs: [docs/specs/](docs/specs/README.md)
- Learnings: [docs/learnings/](docs/learnings/README.md)
- Terms: [docs/GLOSSARY.md](docs/GLOSSARY.md)

## Credits

The `dx:grill` skill is based on the [grilling skill](https://github.com/mattpocock/skills/blob/main/skills/productivity/grilling/SKILL.md) of [mattpocock/skills](https://github.com/mattpocock/skills). Copyright (c) 2026 Matt Pocock. MIT License.

## License

[MIT](LICENSE)
