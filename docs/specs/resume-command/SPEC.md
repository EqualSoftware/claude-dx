---
slug: resume-command
title: Resume command and SessionStart hook
status: Draft
milestone: v0.1.0
priority: P1
area: [cli, plugin]
related: [ADR-008]
---

# Resume command and SessionStart hook

## 1. Summary

Add the command `claude-dx resume` and a `SessionStart` hook of the plugin that runs it (ADR-008).

## 2. Problem

After a session start, `/clear` or compact, the model has no work state. A handoff file that a person writes needs the model to update it, and it repeats the rules and the task status that other files already have (ADR-008).

## 3. Goals

- G-1: At each session start, `/clear` and compact, the model gets the work state with no tool call.
- G-2: The output is short: 40 lines or less.

## 4. Out of Scope

- A journal of events (ADR-008 rejects it).
- Spec and task tools (docs-spec, docs-task).

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | `claude-dx resume` prints: branch, commits ahead of upstream, changed files; tasks `In Progress` by modification time, with the first unchecked Plan step and the Notes; next `Todo` tasks with all dependencies `Done`; `Blocked` tasks; unpushed commits with their `Refs:` trailers. | Planned |
| R-2 | It uses no network and no MCP server. | Planned |
| R-3 | The plugin has a `SessionStart` hook with the matchers `startup`, `resume`, `clear` and `compact` that runs it. | Planned |
| R-4 | With no tasks and a clean tree, it prints one line. | Planned |

## 6. Design

### Planned modules

- `internal/resume/`: reads the task files, their modification time and git, and writes the state text.
- `internal/cli/`: the `resume` command.
- `internal/plugin/assets/hooks/hooks.json`: the `SessionStart` hook with the matchers `startup`, `resume`, `clear` and `compact`, with `{{BINARY}} resume`.

### Data flow (planned)

1. The hook runs `claude-dx resume` in the project directory.
2. The command reads the tasks in the project data directory and the git state. It uses no network and no MCP server.
3. It prints the state. Claude Code adds the output to the context of the session.

### Decisions

- ADR-008: the resume command, the hook, and the rule that no journal is kept.

## 7. Acceptance Criteria

- R-1, R-4: golden tests of the output in a test repository.
- R-3: plugin test that the hook file is in the extracted plugin.
- Tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Added R-1. |
| 2026-10-26 | Equal | Added R-2. |
| 2026-10-26 | Equal | Added R-3. |
| 2026-10-26 | Equal | Added R-4. |
| 2026-10-27 | Equal | Changed fields: related. |