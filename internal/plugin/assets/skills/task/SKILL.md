---
name: task
description: Create, refine or update a task spec in docs/tasks with the dx task tools. Use when planning work, when the user gives requirements or documents for a feature, or when a task changes status.
---

# Make or update a task

The dx task tools read and write the task files in `docs/tasks/`. They set the numbers, the dates, the Document History rows and `README.md`. Do not change task files with Edit or Write. Do not read `docs/tasks/README.md`: it is for persons. Use `task_list`.

Read only what you need: `task_get` with `section` returns one section.

## Flow

The task tools check each step of the flow of ADR-013:

1. Create: `task_create` makes a `Draft` task.
2. Refine: `/dx:grill`, then `In Review`. The tools warn about other open tasks on the same files.
3. Approve: `Approved` only when the user approves. A change of section 3, 4, 6, 7 or 9 of an `Approved` task sets it to `In Review`, and the user approves it again.
4. Select: `task_list` with `status: ready` gives the next work.
5. Develop: `In Development`. A new finding during work is a new task, not a change of the scope.
6. Close: check each Definition of Done item with `task_dod_check`, then `Done`.

Merge rule: tasks with the same milestone, the same files and a similar goal are one task, if the size stays `S` or `M`.

## List tasks

Run `task_list` with the filters of the request. Each frontmatter field is a filter, example `status`, `priority`, `milestone`, `area`, `owner` or `depends_on` (the tasks that wait for one task). To select the next work, use `status: ready`: it gives the `Approved` tasks that have all dependencies `Done`, in milestone order, then priority order. Give its lines to the user and stop. Do not run other tools: `task_list` has all the data of a list. Run `task_get` only when the user asks for the content of a task.

## Create a task

Each unit of work of the project gets a task. Changes to the agent setup (`.claude/` and the agent rules in `CLAUDE.md`) do not get a task.

1. Collect the inputs: the request of the user, documents, related code, ADRs (`adr_list`) and tasks (`task_list` with `status: all`). Do not repeat an investigation that is already in the conversation. Give a search of more than 3 files to the `dx:scout` agent, with a closed list of questions, and use its `path:line` findings.
2. If a task has the same objective, refine it. Do not make a duplicate.
3. Run `task_create` with the title (40 characters or less, not the title of another task), priority, milestone (`vMAJOR.MINOR.PATCH`, example `v0.1.0`), type, areas, dependencies, size, platforms, related IDs and the sections that you know, but not the Document History, section 13 or section 14. The tool sets the owner to the git `user.name`. The task starts as `Draft`.
4. `depends_on` gives only open tasks that must be done first. The tools refuse an ID with no task file, the same task, a cycle, a `Done` or `Cancelled` task (put it in `related`) and a task of a later milestone. They warn when a dependency has a lower priority. A change of `depends_on` checks only the new IDs, but the cycle check uses all IDs: an old dependency that became `Done` can stay.
5. Write only verified facts. Write "Not applicable" for a section that does not apply.
6. Add each unknown item with `task_question_add`. Do not guess.
7. Use the terms of `docs/GLOSSARY.md`. Run `/dx:glossary` for a new or unclear term.

## Refine a task

Run `/dx:grill` on the task. Use `task_question_add` and `task_question_answer` for the questions. Cover all levels:

1. Goals and motivation.
2. Scope and Out of Scope.
3. Constraints, platforms and related ADRs.
4. Architecture, ownership and threads.
5. Public API and data formats.
6. User-facing behavior (if applicable).
7. Acceptance criteria and metrics.
8. Risks, removal and compatibility.

Change the sections with `task_update_section` and the frontmatter fields with `task_update_fields`. Only their own tools change these sections, so the flow checks stay true:

- Document History: `task_set_status` and the Definition of Done tools add the rows.
- Section 13 (Open Questions): `task_question_add` and `task_question_answer`. A task with no questions keeps the `None.` row.
- Section 14 (Definition of Done): `task_dod_add` and `task_dod_remove` change the items. `task_dod_check` checks an item and adds its evidence. `task_dod_uncheck` unchecks an item. Removal and uncheck need a reason. The template items are valid.

A task is ready for review when:
- each goal has an acceptance criterion and a metric or a check,
- Out of Scope is not empty,
- the Implementation Plan has exact files, steps and checks for each work batch,
- the user confirms the same understanding.

Then run `task_set_status` with `In Review`. The tool also checks that no question is `Open`, that no section has the template text and that the size (`S` or `M`) and the platforms are set. Split a larger task into smaller tasks with `depends_on`.

For `In Review` and `Approved`, the tool warns when another open task gives the same file, or a directory that contains it, in its Implementation Plan, and no `depends_on` path connects the two tasks. Then set the order with `depends_on` or merge the tasks, and tell the user.

## Change the status

Run `task_set_status`. The tool adds the Document History row and refuses a change that breaks a rule. Read its reasons and fix them. Do not work around a refusal.

The allowed changes are:

| From | To |
| :--- | :--- |
| `Draft` | `In Review`, `Cancelled` |
| `In Review` | `Approved`, `Draft`, `Cancelled` |
| `Approved` | `In Development`, `In Review`, `Blocked`, `Cancelled` |
| `In Development` | `Done`, `Blocked`, `Cancelled` |
| `Blocked` | `Approved`, `In Development`, `Cancelled` |
| `Cancelled` | `Draft` |
| `Done` | None. A new finding is a new task. |

- `Approved`: only when the user approves.
- `In Development`: when work starts.
- `Blocked`: give the blocker in `note`.
- `Done`: first check each Definition of Done item with `task_dod_check` and its evidence. A `Done` task is final: the Definition of Done tools refuse it.
- `Cancelled`: give the reason in `note`.

If `task_list` warns that two files have one number, run `task_renumber` with the file name of the newer task.

## Rules

- Write in ASD-STE100. The user-facing text in section 7.4 can use other styles.
- Keep the spec short and exact. Use `path:line` references, not copies of code.
- Link to ADRs. Do not copy them.
- Add a Mermaid diagram to section 7.1 when the task has a call flow, an ownership or lifetime rule, a state change, or data that moves between modules.
- Give the reason of each rule in Requirements and Design that is not obvious: "..., because ...". A rule without a reason is often a hidden decision. Ask it with `/dx:grill`.

## Report

After a create, a refine or a status change, give the task path, its status, its priority and the open questions. The path is in the `task_create` result, and `task_list` gives the ID. Do not run other tools to find the path.
