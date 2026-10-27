---
slug: tasks-to-specs
title: Migrate tasks to specs
status: Draft
milestone: v0.1.0
priority: P1
area: [docs, plugin]
related: [ADR-016, ADR-017, ADR-013]
---

# Migrate tasks to specs

## 1. Summary

Move the content of docs/tasks to specs, delete docs/tasks and HANDOFF.md, and change the skills, rules and ADR references (ADR-016, ADR-017).

## 2. Problem

After TASK-063 to TASK-066 (`resume-command`) the project still has 60 or more task files in git and text that gives TASK-NNN IDs.

## 3. Goals

- G-1: No `docs/tasks/` folder and no HANDOFF.md.
- G-2: Each feature of the project has a spec with its requirements and their status.
- G-3: No ADR, learning, glossary row, skill or rule gives a TASK-NNN ID.

## 4. Out of Scope

- New features. The open tasks of v0.2.0 and later become `Draft` specs, and the user plans the milestones again after this task.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Group the `Done` tasks into specs by area. Their requirements are `Done`. The sequence diagram of TASK-064 section 7.1 goes to the spec of local tasks. | Planned |
| R-2 | The open tasks of v0.2.0 and later become `Draft` specs with their milestone. | Planned |
| R-3 | Change TASK-NNN references in ADRs, learnings and the glossary to `<slug>#R-N` (the one exception of ADR-017). | Planned |
| R-4 | Change the task, adr and grill skills, `rules.md`, the scout agent, `CLAUDE.md` and `README.md` to the spec and local task model. | Planned |
| R-5 | Delete `docs/tasks/` and HANDOFF.md, and remove `/HANDOFF.md` from `.gitignore`. | Planned |
| R-6 | Set ADR-016 and ADR-017 to `Accepted` with the user approval, then supersede ADR-013 with ADR-016. | Planned |

## 6. Design
[How the feature works: modules, data flow, formats, user-facing behavior. Link to the ADRs; do not copy them. A new architecture approach needs an ADR.]

## 7. Acceptance Criteria

- G-1: `ls docs/tasks HANDOFF.md` finds nothing.
- G-3: `grep -rn TASK-[0-9] docs internal/plugin CLAUDE.md` finds nothing.
- The user approves the spec list.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which specs do the 62 tasks become? Proposal: launcher-and-cli, cache-and-update, mcp-server, document-tools, git-tools, lsp-servers, release. | Maintainer | Yes: launcher-and-cli, cache-and-update, mcp-server, document-tools, git-tools, lsp-servers, release. |
| Q-2 | The glossary rows Task and the task terms describe the old model. Do we change them in this task, with the new rows Spec, Requirement and Resume? | Maintainer | Yes. Change the task terms and add Spec, Requirement and Resume in this task. |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Added R-1. |
| 2026-10-26 | Equal | Added R-2. |
| 2026-10-26 | Equal | Added R-3. |
| 2026-10-26 | Equal | Added R-4. |
| 2026-10-26 | Equal | Added R-5. |
| 2026-10-26 | Equal | Added R-6. |