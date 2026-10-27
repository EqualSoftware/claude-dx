# ADR-008: Specs and tasks model

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
A project needs the current definition of each feature, and a way to track the work on it. One task file in git for each unit of work gives too many files in a large project. A task number in git collides when two persons make tasks on two branches. Task files also do not give the current behavior of a feature: the data is in many closed tasks. Tasks of one spec must move to another spec, or be deleted, with no lost or broken dependency. A task must be a full brief, so that a subagent can do it with no other context.

## Decision
Spec:
- A spec is the current definition of one feature or scope. It is in git: `docs/specs/<slug>/SPEC.md`, and other files of the spec (diagrams, designs) in the same folder. The ID is the slug, with no number, so two branches do not get the same ID. The slug `general` is reserved.
- Frontmatter: `slug`, `title`, `status`, `milestone`, `priority`, `area`, `related` (ADRs).
- Sections: Summary, Problem, Goals, Out of Scope, Requirements, Design, Acceptance Criteria, Open Questions, Document History. Requirements is a table `ID | Requirement | Status`. A requirement ID is `R-N` and does not change. Its status is `Planned` or `Done`.
- Status: `Draft`, `In Review`, `Approved`. The allowed changes are `Draft` to `In Review`, `In Review` to `Approved` or `Draft`, and `Approved` to `In Review`. `In Review` needs the leaf sections filled. `Approved` needs no `Open` question. A change of Goals, Out of Scope, Requirements, Design or Acceptance Criteria of an `Approved` spec sets it to `In Review`.
- Only the spec tools change Requirements, Open Questions and Document History, so a free text edit cannot pass a gate.
- A spec does not depend on another spec, so each spec can be done alone and in parallel.
- A spec that is removed from the product, or split, is deleted. Git keeps the old version. A new spec from a split gives the old slug in its Document History. The reason of a delete goes to the commit message.
- Design links to the ADRs and does not copy them. A spec change that selects a new architecture approach needs an ADR.
- A spec does not contain task IDs. A commit gives the requirements that it implements in a trailer `Refs: <slug>#R-N`. A trailer stays the same after a rebase. `spec_get` finds the commits of each requirement with `git log`.
- Tools: `spec_list`, `spec_get`, `spec_create`, `spec_update_fields`, `spec_update_section`, `spec_set_status`, `spec_req_add`, `spec_req_set`, `spec_question_add`, `spec_question_answer`, `spec_delete`. `docs/specs/README.md` is the index for persons.
- `spec_delete` refuses while `tasks/<slug>/` has a task. It gives the number of tasks and the tools to move or delete them. It does not delete tasks.

Task:
- A task is a local work item of one developer. It is in `tasks/<spec>/<task-slug>.md` of the project data directory (ADR-007), or in `tasks/general/<task-slug>.md` for work that has no spec. Git does not track it.
- The ID is `<spec>/<task-slug>`. There is no counter.
- A task points to one spec. A spec can have many tasks. `depends_on` gives tasks of the same spec only, because a dependency on a task of another spec makes a dependency between specs.
- Frontmatter: `id`, `title`, `status`, `priority`, `covers` (requirement IDs), `depends_on`. The milestone is the milestone of the spec.
- Sections: Brief, Plan (a checklist of steps), Notes, Definition of Done. The Brief has what to do, the files, the constraints and the check commands. Notes is for progress.
- `task_create` requires a brief and a plan with one or more steps.
- `task_update_section` changes only Brief, Plan and Notes. Only the Definition of Done tools (`task_dod_add`, `task_dod_check`, `task_dod_uncheck`, `task_dod_remove`) change the section Definition of Done. `task_dod_check` writes its evidence as a sub-item. `task_dod_remove` needs a reason, and the tool writes the item text and the reason to Notes.
- Status: `Todo`, `In Progress`, `Blocked`, `Done`. There is no approval, because the spec has the approved scope. The allowed changes are `Todo` to `In Progress` or `Blocked`, `In Progress` to `Todo`, `Blocked` or `Done`, and `Blocked` to `Todo` or `In Progress`. `Done` is final.
- `In Progress` needs an `Approved` spec and `Done` dependencies. `Blocked` needs the blocker in the note. `Done` needs all Plan steps and Definition of Done items checked.
- When a task goes to `Done`, the tool sets the requirements in `covers` to `Done` in the spec and adds a Document History row with no task ID.
- `task_move(ids, to)` moves tasks of one spec to another spec. The source and the target can be `general`. The tool checks all tasks before any change. It refuses a dependency between specs and a slug that the target spec has. It removes `covers` and adds a Notes line: `- Moved from <old>. Old covers: R-2, R-5.` When the task has no `covers`, the line is `- Moved from <old>.` It removes an empty `tasks/<spec>/` folder. A retry accepts a task that is already in the target with its "Moved from" line.
- `task_delete(ids)` deletes tasks. It refuses a missing ID and a task outside the set that depends on a task of the set.
- `task_move` and `task_delete` accept `Done` tasks. The other task tools refuse a change of a `Done` task.
- The docs lock of the project data directory also protects the tasks, so a task change and its spec change need only one lock (ADR-006).
- Other tools: `task_list`, `task_get`, `task_update_section`, `task_update_fields`, `task_set_status`, `task_step_check`, `task_dod_check`, `task_dod_uncheck`, `task_dod_add`, `task_dod_remove`.

Resume:
- The command `claude-dx resume` prints a short state of the work: the tasks `In Progress` with the first unchecked Plan step and the Notes, the next tasks, the `Blocked` tasks, the unpushed commits with their `Refs:` trailers and the changed files. It gets all data from the task files, their modification time and git. It keeps no journal.
- A `SessionStart` hook of the plugin runs it on `startup`, `resume`, `clear` and `compact`, so the model gets the state with no tool call and no rule.

## Alternatives
- Tasks in `docs/tasks` in git: Rejected. Too many files in a large project, number collisions between branches, and the current behavior of a feature is in many closed tasks.
- Tasks in `.dx/` at the project root, in `.gitignore`: Rejected. Each worktree gets its own tasks, so parallel work in worktrees splits the tasks.
- Tasks in the claude-dx cache: Rejected. The cache is not the place of user work.
- Spec ID `SPEC-NNN` with a counter: Rejected. Two branches can get the same number.
- One task file for each requirement: Rejected. One task can cover many requirements, and one requirement can need many tasks.
- Status `Deprecated` for a removed or split spec: Rejected. Git keeps the old version, and a deleted spec is simpler.
- A commit hash column in the Document History: Rejected. A commit cannot contain its own hash, and a rebase changes the hash.
- A journal of tool events for the resume: Rejected. Task files, their modification time and git give the same data with no extra file and no extra write.
- Resume text in the MCP server instructions: Rejected. The server sends the instructions one time at start, so they are old after `/clear` and compact.
- A callback in the spec store so `spec_delete` moves the tasks: Rejected. It inverts the dependency: the task package imports the spec package, and the spec package does not import the task package.

## Consequences
- Positive: few files in git. Specs give the current behavior of each feature. No ID collision between branches. Many processes can work on different specs at the same time. A subagent gets a full brief from the task file. A task can move or be deleted with a clear refusal before any change.
- Negative: tasks are local, so another person does not see them, and they are lost with the clone. `task_create` needs more input. A retry of `task_move` after a crash can leave two task files for a short time.
- Follow-up: build `claude-dx resume` and its hook (spec resume-command).

## Evidence
- internal/docs/spec/spec.go:1 and :27-35 (`SPEC.md`, statuses), :50-59 (frontmatter).
- internal/docs/spec/change.go:36-41 (protected sections), :64 (scope sections), :102-106 (spec status changes), :109-125 (`SetStatus`), :133-141 (`Approved` and `In Review` checks).
- internal/docs/spec/commits.go:24-30 (`Refs:` trailers from `git log`).
- internal/docs/spec/delete.go:60-82 (refuses a spec with tasks).
- internal/docs/task/task.go:1-3 and :25-29 (task path, `general`, statuses).
- internal/docs/task/change.go:100-105 (task status changes), :107-135 (`Done` sets the covers), :137-149 (checks).
- internal/docs/task/create.go:41-42 (brief required), internal/docs/task/move.go:22-52 (`Delete`), :94-121 (`Move`), :238-241 (the Notes line with and without old covers).
- internal/docs/task/change.go:21-31 (`task_update_section` accepts Brief, Plan and Notes only), :233-234 (evidence sub-item), :300-313 (`RemoveDoD` needs a reason and writes Notes).
- internal/mcpserver/server.go:44-47 (the task store is in the data directory and shares the docs lock).
- internal/docs/spec/tools.go:90-144 and internal/docs/task/tools.go:93-152 (tool list).
- Not built yet: `claude-dx resume`. The spec is docs/specs/resume-command/SPEC.md.
