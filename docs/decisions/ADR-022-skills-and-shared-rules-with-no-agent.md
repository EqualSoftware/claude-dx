# ADR-022: Skills and shared rules with no agent

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context

Replaces ADR-009.

ADR-009 adds the agent `dx:scout`, and the skills and `rules.md` tell the model to give fact searches to it. When the session runs with an agent plugin, for example `dx-agents`, the hook of that plugin allows only its own agents. The rule of dx then causes an agent call that fails. The orchestration of agents is the work of the agent plugin of the session, not of dx.

## Decision
- The dx plugin has no agent.
- The skills and `rules.md` do not tell the model which agent to use for a search.
- Skills of v0.1: `/dx:adr`, `/dx:task`, `/dx:learning`, `/dx:glossary`, `/dx:grill`. Skills of v0.2: `/dx:commit`, `/dx:eureka`.
- `/dx:eureka` records lessons only as an ADR, a learning or a glossary term.
- The launcher passes the shared rules file `rules.md` of the dx plugin with `--append-system-prompt-file`, because Claude Code uses only the last `--append-system-prompt`, and the user can give one. The rules are short.
- ASD-STE100 always applies to dx documents, commit messages, tool arguments, tool results and skill files. There is no `free` language option.
- The chat with the user uses the language of the user.
- Skills are written in ASD-STE100.

## Alternatives
- Keep the dx:scout agent: Rejected. The hook of the dx-agents architect blocks it, and the rule causes a failed agent call.
- A generic rule: give a search to a read-only search agent of the session, if there is one: Rejected by the user. The agent plugin of the session gives its own rules for searches.
- A RULES.md that each skill reads: Rejected, as in ADR-009. One file read for each skill. The system prompt is cached.

## Consequences
- Positive: No conflict with an agent plugin. dx has fewer files to keep.
- Negative: With no agent plugin, fact searches run in the main session and cost more.
- Follow-up: Remove agents/scout.md and the scout rule from rules.md and the skills. Remove the glossary term Scout. Change README.md, the dx plugin glossary term, R-4 of spec tasks-to-specs and the learning subagent-tool-names.

## Evidence
docs/decisions/ADR-009-skills-agent-and-rules.md:15; internal/plugin/assets/rules.md:7; internal/plugin/assets/agents/scout.md; hook refusal in the session of 2026-10-26: "Agent type 'dx:scout' is not a worker agent."
