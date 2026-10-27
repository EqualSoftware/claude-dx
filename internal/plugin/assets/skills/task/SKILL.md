---
name: task
description: Create or update a local task with the dx task tools. Use to make the work of an Approved spec into tasks, when work on a task starts, or when a task changes status.
---

# Make or update a task

A task is a local file, not a file of the repository. The dx task tools read and write it in the project data directory, shared by all worktrees. The task ID is `<spec>/<task-slug>`, example `docs-lock/shared-reads`. A task with no spec has the spec `general`. Do not change task files with Edit or Write.

Read only what you need: `task_get` with `section` returns one section (`Brief`, `Plan`, `Notes` or `Definition of Done`).

Refinement, grill and approval of the scope belong to `/dx:spec`. A task does not change the scope of its spec.

## List tasks

Run `task_list` with `spec` and `status` as the request needs. Each line has the ID, the status, the priority and the title. Give the lines to the user and stop. Run `task_get` only when the user asks for the content of a task.

## Create a task

1. Read the spec (`spec_get`). Make tasks for the requirements of the spec. The spec can be in any status, but a task starts only when the spec is `Approved`.
2. Run `task_list` for the spec. Do not make a duplicate.
3. Run `task_create` with:
   - `spec` and `slug` (kebab-case, 40 characters or less), `title` (40 characters or less) and `priority` (`P0` to `P3`).
   - `brief`: a full brief, so a subagent can do the task with no other context. Give what to do, the files, the constraints and the check commands. The reason is in the requirements of the spec.
   - `plan`: one or more steps. Each step is one action with its check, example `Add SetDoneLocked. Check: go test ./internal/docs/spec`.
   - `covers`: the requirement IDs of the spec that the task implements. A `general` task has none.
   - `depends_on`: only tasks of the same spec that must be `Done` first. The tools refuse a missing task, the same task, a cycle and a task of another spec.
4. Write only verified facts. Use the terms of `docs/GLOSSARY.md`.

The task starts as `Todo`. Change the Brief, the Plan or the Notes with `task_update_section`. Change the title, the priority, `covers` and `depends_on` with `task_update_fields`.

## Change the status

Run `task_set_status`. The tool refuses a change that breaks a rule. Read its reason and fix it. Do not work around a refusal.

| From | To |
| :--- | :--- |
| `Todo` | `In Progress`, `Blocked` |
| `In Progress` | `Todo`, `Blocked`, `Done` |
| `Blocked` | `Todo`, `In Progress` |
| `Done` | None. |

- `In Progress`: needs all `depends_on` tasks `Done` and an `Approved` spec (not for a `general` task).
- `Blocked`: give the blocker in `note`.
- `Done`: first check each Plan step with `task_step_check` and each Definition of Done item with `task_dod_check` and its evidence. The tool sets the `covers` requirements to `Done` in the spec. A `Done` task is final: all task tools refuse a change of it, except `task_move` and `task_delete`.
- For the other statuses, `note` is optional. The tool adds it to Notes.

A choice that the task does not decide is a question to the user before you write it (rules). A new finding during work is a new task, not a change of the scope.

## Plan and Definition of Done

- `task_step_check` checks one Plan step by number. Change the steps with `task_update_section`.
- `task_dod_check` checks an item and adds its evidence. `task_dod_uncheck` unchecks an item.
- `task_dod_add` adds an item. `task_dod_remove` removes an item and needs a reason, which the tool adds to Notes.
- Section `Definition of Done` changes only with the tools of this list.

## Move or delete tasks

- `task_move(ids, to)` moves tasks of one spec to another spec, or to or from `general`. The tool removes `covers` and writes the old `covers` in Notes, because the requirement IDs have no meaning in the new spec. Set the new `covers` with `task_update_fields`.
- `task_delete(ids)` deletes tasks. A task out of the set cannot depend on a task of the set.
- `spec_delete` refuses while the spec has tasks. Move or delete them first.
- Both tools check all tasks before any change. After a failure, run the same call again.

## Rules

- Write in ASD-STE100.
- Keep the Brief short and exact. Use `path:line` references, not copies of code.

## Report

After a create or a status change, give the task ID, its status and its priority.
