---
name: task
description: Create, refine or update a task spec in docs/tasks with the dx task tools. Use when planning work, when the user gives requirements or documents for a feature, or when a task changes status.
---

# Make or update a task

The dx task tools read and write the task files in `docs/tasks/`. They set the numbers, the dates, the Document History rows and `INDEX.md`. Do not change task files with Edit or Write. Do not read `docs/tasks/INDEX.md`: it is for persons. Use `task_list`.

Read only what you need: `task_get` with `section` returns one section.

## List tasks

Run `task_list` with the filters of the request. Each frontmatter field is a filter, example `status`, `priority`, `area`, `owner` or `depends_on` (the tasks that wait for one task). Give its lines to the user and stop. Do not run other tools: `task_list` has all the data of a list. Run `task_get` only when the user asks for the content of a task.

## Create a task

Each unit of work of the project gets a task. Changes to the agent setup (`.claude/` and the agent rules in `CLAUDE.md`) do not get a task.

1. Collect the inputs: the request of the user, documents, related code, ADRs (`adr_list`) and tasks (`task_list` with `status: all`). Do not repeat an investigation that is already in the conversation. Give a search of more than 3 files to the `dx:scout` agent, with a closed list of questions, and use its `path:line` findings.
2. If a task has the same objective, refine it. Do not make a duplicate.
3. Run `task_create` with the title, priority, type, areas, dependencies, size, platforms, related IDs and the sections that you know. The tool sets the owner to the git `user.name`. The task starts as `Draft`.
4. Write only verified facts. Write "Not applicable" for a section that does not apply.
5. Add each unknown item with `task_question_add`. Do not guess.
6. Use the terms of `docs/GLOSSARY.md`. Run `/dx:glossary` for a new or unclear term.

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

Change the sections with `task_update_section` and the frontmatter fields with `task_update_fields`.

A task is ready for review when:
- each goal has an acceptance criterion and a metric or a check,
- Out of Scope is not empty,
- the Implementation Plan has exact files, steps and checks for each work batch,
- the user confirms the same understanding.

Then run `task_set_status` with `In Review`. The tool also checks that no question is `Open`, that no section has the template text and that the size (`S` or `M`) and the platforms are set. Split a larger task into smaller tasks with `depends_on`.

## Change the status

Run `task_set_status`. The tool adds the Document History row and refuses a change that breaks a rule. Read its reasons and fix them. Do not work around a refusal.

- `Approved`: only when the user approves.
- `In Development`: when work starts.
- `Blocked`: give the blocker in `note`.
- `Done`: first check each Definition of Done item with `task_dod_check` and its evidence.
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
