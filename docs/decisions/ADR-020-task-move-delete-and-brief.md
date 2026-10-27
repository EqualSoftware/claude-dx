# ADR-020: Task move, delete and Brief

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context
ADR-016 lists the task tools, but not task_move, task_delete, the refusal of spec_delete while a spec has tasks, or the Brief section. Tasks of one spec must move to another spec, or be deleted, without a lost or broken dependency. A task must be a full brief, so that a subagent can do it with no other context (TASK-064 Q-19, Q-32). ADR-016 is pushed and does not change (ADR-017), so this ADR adds to it and does not supersede it.

## Decision
Add these to ADR-016:
- task_move(ids, to) moves one or more tasks of one spec to another spec. The source and the target can be general. The tool checks all tasks before any change. It refuses a dependency between specs and a slug that the target spec has. It removes covers and adds a Notes line: "- Moved from <old>. Old covers: R-2, R-5." When the task has no covers, the line is "- Moved from <old>.". It removes an empty tasks/<spec>/ folder.
- A retry of task_move accepts a task that is already in the target with its "Moved from" Notes line. A retry after a crash between the new file and the old file removes the old file.
- task_delete(ids) deletes one or more tasks. It refuses a missing ID and a task outside the set that depends on a task of the set. It removes an empty tasks/<spec>/ folder.
- task_move and task_delete accept Done tasks. The other task tools refuse a change of a Done task.
- spec_delete refuses while tasks/<slug>/ has a .md file. It gives the number of tasks and the tools to move or delete them. It does not delete tasks.
- task_create requires brief and plan (one or more steps). The Brief has what to do, the files, the constraints and the check commands. The section order is Brief, Plan, Notes, Definition of Done. Notes is for progress.

## Alternatives
- A callback field in spec.Store that the MCP server sets to the task store, so spec_delete moves the tasks: Rejected: it inverts the dependency and mixes the two packages. The task package imports the spec package, and the spec package does not import the task package.
- A third package with the rules between specs and tasks: Rejected: it adds a layer that the task does not need.
- Owner references and a garbage collector: Rejected: the error comes late, and a move to a new spec is not possible.
- move_to and discard_tasks parameters of spec_delete: Replaced by task_move and task_delete, because the tasks must move or be deleted by the same rules in one tool each.

## Consequences
- Positive: A task can move between specs, or be deleted, with clear refusals before any change. A spec with tasks cannot be deleted by mistake. A subagent gets a full brief from the task file.
- Negative: task_create needs more input. A task that a retry touches can have two files for a short time after a crash.
- Follow-up: Accept this ADR when the user approves it. Implement the tools in TASK-064 batch 2 and 3, and the Brief section in the dx:task skill (batch 4).

## Evidence
TASK-064 answers Q-17, Q-18, Q-19, Q-23, Q-29, Q-32 and Q-43 (task_get TASK-064, 2026-10-26). ADR-016 is Accepted and lists the task tools without task_move, task_delete or the Brief section (adr_get ADR-016, 2026-10-26).
