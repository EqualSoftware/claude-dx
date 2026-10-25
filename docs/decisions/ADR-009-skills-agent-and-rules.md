# ADR-009: Skills, the scout agent and the shared rules of the dx plugin

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The tools do the mechanical work. The skills tell the model when and how to use the tools. In the reference workflow, each skill reads a shared `RULES.md` file, and fact searches go to a cheap agent.

## Decision
- Skills of v0.1: `/dx:adr`, `/dx:task`, `/dx:learning`, `/dx:glossary`, `/dx:grill`.
- Skills of v0.2: `/dx:commit`, `/dx:eureka`.
- `/dx:eureka` records lessons only as an ADR, a learning or a glossary term.
- Agent `dx:scout`: model Haiku, read-only. Tools: `Read`, `Grep`, `Glob` and the read-only dx tools. Skills give it fact searches of more than 3 files.
- The launcher passes the shared rules file `rules.md` of the dx plugin with `--append-system-prompt-file`, because Claude Code uses only the last `--append-system-prompt`, and the user can give one (TASK-008 Q-2). They are short (approximately 10 lines). Examples: write dx documents, MCP messages and commit messages in ASD-STE100; use the `git_*` tools for git; do not read `docs/tasks/INDEX.md`.
- ASD-STE100 always applies to dx documents, commit messages, tool arguments, tool results and skill files. There is no `free` language option.
- The chat with the user uses the language of the user.
- Skills and the agent are written in ASD-STE100.

## Alternatives
- A `RULES.md` that each skill reads: rejected. One file read for each skill. The system prompt is cached.
- No scout agent: rejected. Fact searches in the main session use the full context for each tool call.

## Consequences
- Positive: rules apply once for each session. Cheap fact searches.
- Negative: the rules are in each session, also when no dx skill runs.
- Follow-up: none.

## Evidence
- Design interview (2026-10-25): Q-25, Q-38, Q-43, Q-45.
- Claude Code 2.1.308 (2026-10-25): with two `--append-system-prompt` flags, or two `--append-system-prompt-file` flags, only the last applies. One of each applies together.
