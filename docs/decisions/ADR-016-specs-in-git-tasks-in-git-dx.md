# ADR-016: Specs in git, tasks in .git/dx

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context
Each unit of work is a task file in docs/tasks, in git. QA, explorations and small fixes make many task files: 62 tasks before v0.1.0. In a large project this is too many files. A task ID in git also has a number collision when two persons make tasks on two branches, because the counter of ADR-007 is local to one machine. The task files also do not give the current behavior of a feature: the data is in many closed tasks. The hand-written HANDOFF.md repeats rules, user preferences and task status that other files already have, and the model must remember to update it.

## Decision

Spec:
- A spec is the current definition of one feature or scope. It is in git: `docs/specs/<slug>/SPEC.md`, and other files of the spec (diagrams, designs) in the same folder. The ID is the slug, with no number, so two branches do not get the same ID. The slug `general` is reserved.
- Frontmatter: `slug`, `title`, `status`, `milestone`, `priority`, `area`, `related` (ADRs).
- Sections: Summary, Problem, Goals, Out of Scope, Requirements, Design, Acceptance Criteria, Open Questions, Document History. Requirements is a table `ID | Requirement | Status`. A requirement ID is `R-N` and does not change. Its status is `Planned` or `Done`.
- Status: Accepted
- A spec does not depend on another spec, so each spec can be done alone and in parallel.
- A spec that is removed from the product, or split, is deleted. Git keeps the old version. A new spec from a split gives the old slug in its Document History. The reason of a delete goes to the commit message.
- Design links to the ADRs and does not copy them. A spec change that selects a new architecture approach needs an ADR.
- A spec does not contain task IDs. A commit gives the requirements that it implements in a trailer `Refs: <slug>#R-N`. A trailer stays the same after a rebase. `spec_get` finds the commits of each requirement with `git log`.
- The tools run git directly, as `internal/project/project.go:57` does. The git runner of TASK-009 can replace this later.

Task:
- A task is a local work item of one developer. It is in `<git-common-dir>/dx/tasks/<spec>/<task-slug>.md`, or in `<git-common-dir>/dx/tasks/general/<task-slug>.md` for work that has no spec. Git does not track the git common directory, and all worktrees of a repository share it.
- Outside a git repository, the tasks are in `<cache>/projects/<key>/tasks/`, with the key of ADR-007. When the project becomes a git repository, the first task tool call moves them to `<git-common-dir>/dx/tasks/`.
- The ID is `<spec>/<task-slug>`. There is no counter.
- A task points to one spec. A spec can have many tasks. `depends_on` gives tasks of the same spec only, because a dependency on a task of another spec makes a dependency between specs.
- Frontmatter: `id`, `title`, `status`, `priority`, `covers` (requirement IDs), `depends_on`. The milestone is the milestone of the spec.
- Sections: Plan (a checklist of steps), Notes, Definition of Done.
- Status: Accepted
- When a task goes to `Done`, the tool sets the requirements in `covers` to `Done` in the spec and adds a Document History row with no task ID.
- The task lock is `tasks.lock` in the task folder, shared by all worktrees.

Tools:
- Add: `spec_list`, `spec_get`, `spec_create`, `spec_update_section`, `spec_set_status`, `spec_req_add`, `spec_req_set`, `spec_question_add`, `spec_question_answer`, `spec_delete`, `task_step_check`.
- Change: all `task_*` tools use the local path and the slug ID.
- Remove: `task_renumber`, `task_question_add`, `task_question_answer`, `docs/tasks/` and its `README.md` index. `docs/specs/README.md` is the index for persons.

Resume:
- The command `claude-dx resume` prints a short state of the work: the tasks `In Progress` with the first unchecked Plan step and the Notes, the next tasks, the `Blocked` tasks, the unpushed commits with their `Refs:` trailers and the changed files. It gets all data from the task files, their modification time and git. It keeps no journal.
- A `SessionStart` hook of the plugin runs it on `startup`, `resume`, `clear` and `compact`, so the model gets the state with no tool call and no rule.
- HANDOFF.md is removed.

## Alternatives
- Keep tasks in docs/tasks in git: Too many files in a large project, number collisions between branches, and the current behavior of a feature is in many closed tasks.
- Tasks in .dx/ at the project root, in .gitignore: Each worktree gets its own tasks, so parallel work in worktrees splits the tasks.
- Tasks in the claude-dx cache: `cache clean` deletes the tasks.
- Spec ID SPEC-NNN with a counter: Two branches can get the same number.
- One task file for each requirement (<spec>/<R-N>.md): One task can cover many requirements, and one requirement can need many tasks.
- Status Deprecated for a removed or split spec: Git keeps the old version. A deleted spec is simpler.
- Commit hash column in the Document History: A commit cannot contain its own hash, and a rebase changes the hash.
- A journal of tool events for the resume: Task files, their modification time and git give the same data with no extra file and no extra write.
- Resume text in the MCP server instructions: The server sends the instructions one time at start, so they are old after /clear and compact.

## Consequences
- Positive: Few files in git. Specs give the current behavior of each feature. No ID collision between branches. Many processes can work on different specs at the same time. The model gets the work state at each session start with no tool call, and the user does not ask for a handoff.
- Negative: Tasks are local, so another person does not see them, and they are lost with the clone. Task tools need a git repository. ADR and learning text that gives old TASK-NNN IDs needs a change at migration. The spec and task tools need new code and a migration of 62 tasks.
- Follow-up: Supersede ADR-013 when this ADR is Accepted. Change the task part of ADR-006 and the task counter of ADR-007 in the implementation tasks. Migrate the tasks of docs/tasks to specs and delete docs/tasks. Add the glossary terms Spec, Requirement, Task and Resume.

## Evidence
- 62 task files in docs/tasks before v0.1.0 (`ls docs/tasks`, 2026-10-26).
- HANDOFF.md has 124 lines of rules, preferences and task status (2026-10-26).
- `internal/project/project.go:32` already gets the git common directory.
- ADR-007: the counter is in the local cache, so it is unique on one machine only.
