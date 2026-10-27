# ADR-010: Skills and shared rules, no agent

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
The tools do the mechanical work. The skills tell the model when and how to use the tools. A short set of shared rules must hold in each session. When the session runs with an agent plugin, for example `dx-agents`, the hook of that plugin allows only its own agents. A dx rule that tells the model to call a dx agent then causes an agent call that fails. The orchestration of agents is the work of the agent plugin of the session, not of dx.

## Decision
- The dx plugin has no agent.
- The skills and `rules.md` do not tell the model which agent to use for a search.
- Skills: `/dx:adr`, `/dx:spec`, `/dx:task`, `/dx:learning`, `/dx:glossary`, `/dx:grill`. The spec git-tools plans `/dx:commit` and `/dx:eureka`. `/dx:eureka` records lessons only as an ADR, a learning or a glossary term.
- The launcher passes the shared rules file `rules.md` of the dx plugin with `--append-system-prompt-file`, because Claude Code uses only the last `--append-system-prompt`, and the user can give one. One flag of each kind applies together. The rules are short.
- ASD-STE100 always applies to dx documents, commit messages, tool arguments, tool results and skill files. There is no `free` language option.
- The chat with the user uses the language of the user.
- Skills are written in ASD-STE100.

## Alternatives
- An agent in the dx plugin for fact searches: Rejected. The hook of an agent plugin can block it, and the rule causes a failed agent call.
- A generic rule: give a search to a read-only search agent of the session, if there is one: Rejected by the user. The agent plugin of the session gives its own rules for searches.
- A `RULES.md` that each skill reads: Rejected. One file read for each skill. The system prompt is cached.

## Consequences
- Positive: no conflict with an agent plugin. dx has fewer files to keep. The rules apply once for each session.
- Negative: the rules are in each session, also when no dx skill runs. With no agent plugin, fact searches run in the main session and cost more.
- Follow-up: none.

## Evidence
- internal/plugin/assets has `rules.md` and `skills/` only: no agents folder (`ls internal/plugin/assets`).
- internal/plugin/assets/rules.md:1-9 (the shared rules).
- internal/launcher/launcher.go:83-93 (`--append-system-prompt-file`).
- internal/plugin/assets/skills/{adr,glossary,grill,learning,spec,task}/SKILL.md (the six skills).
- Claude Code 2.1.308 (2026-10-25): with two `--append-system-prompt` flags, or two `--append-system-prompt-file` flags, only the last applies. One of each applies together.
