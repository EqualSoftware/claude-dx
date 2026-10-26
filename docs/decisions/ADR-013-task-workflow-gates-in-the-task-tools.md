# ADR-013: Task workflow gates in the task tools

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The v0.1.0 QA session of 2026-10-25 made 8 tasks and found these problems in the task process:
- Long titles gave long file names, and the 60-character slug cut left words such as "of" and "the" at the end.
- New tasks had `depends_on` with tasks that are `Done`. This sets no order.
- Open tasks change the same files with no order between them, so parallel work gives rebase conflicts.
- One task (TASK-026) needed a change of the scope of an `Approved` task (TASK-013).
- No field tells the release of a task. The README roadmap gives task number ranges that are not correct now.
- `task_create` copies the template comments into the frontmatter.
Rules in skill text did not stop these problems. ADR-005 puts mechanical work in tools, so the same input gives the same result.

## Decision

The task tools check the task process at each status step. A check refuses a change that is always wrong, and gives a warning for a change that can have a reason.

1. Create (`task_create`, `task_update_fields`):
   - Title: 40 characters or less, for tasks and ADRs. The slug is never longer than the title, so file names are `TASK-NNN-<slug>.md` and `ADR-NNN-<slug>.md` with a slug of 40 characters or less.
   - All task fields are in the frontmatter: `id`, `title`, `status`, `priority`, `milestone`, `type`, `size`, `owner`, `platforms`, `area`, `depends_on`, `related`. The Field/Value table under the title is removed, because two formats need two parsers and the tools can filter only frontmatter fields.
   - `blocks` and `reviewers` are removed. `blocks` is the inverse of `depends_on` and becomes wrong. `owner` is the git `user.name` of the person who creates the task.
   - `milestone` (example `v0.1.0`) is required.
   - `depends_on`: refuse an ID that does not exist, the same task, a cycle, a task that is `Done` (use `related`), and a task of a later milestone. Warn when a task depends on a task of lower priority.
   - The frontmatter has no template comments.
2. Refine (`task_set_status` to `In Review` or `Approved`): warn when another open task gives the same file in its Implementation Plan and no `depends_on` path connects the two tasks. Then set the order or merge the tasks.
3. Approve: a change of Requirements, Design or Acceptance Criteria of an `Approved` task sets it to `In Review`. The user approves it again. This is the only way to change the scope of an approved task.
4. Select: `task_list(status: ready)` gives the open tasks with all dependencies `Done`, in milestone order, then priority order. `task_list(milestone)` filters one milestone. `INDEX.md` groups tasks by milestone.
5. Develop: a new finding during work is a new task, not a scope change.
6. Merge rule (skill text): tasks with the same milestone, the same files and a similar goal are one task, if the size stays `S` or `M`.

Before v0.1.0, all titles longer than 40 characters get shorter. Only files with a slug longer than 40 characters get a new name, because the older files have short names that a person selected. Links change in the same commit.

## Alternatives

- Rules only in the /dx:task skill text: Rejected. The QA session followed the skill and still made all the problems.
- Remove stop words at the end of a cut slug: Rejected. It hides the cause: titles that are too long.
- Title limit of 60 or 50 characters: Rejected by the maintainer. 40 characters gives short file names and short task_list rows.
- Mark the release with an area, example v0.1: Rejected. It mixes two meanings in one field.
- Warning for a dependency on a Done task: Rejected. It is never necessary, so a refusal is correct.
- Rename all files to the slug of the new title: Rejected. The older file names are already 40 characters or less, so a rename only breaks links.
- Keep the Field/Value table: Rejected. Two formats need two parsers, and the tools cannot filter table values.
- Calculate `blocks` from `depends_on`: Rejected by the maintainer. `depends_on` has the same data.
- `owner` with a role, example Maintainer: Rejected by the maintainer. The file must record the person who creates the task.

## Consequences
- Positive: The tools stop each problem of the QA session at the step where it occurs. task_list(status: ready) gives the next work with one call. The release content comes from the milestone field, not from number ranges.
- Negative: The task file format gets a required field. The overlap check reads section 8 of all open tasks at each status change. Some links in docs and history change because of the renames.
- Follow-up: Tasks for each check. ADR-006 refers to this ADR. The /dx:task skill gives the flow and the merge rule. The README roadmap uses milestones.

## Evidence

- QA session of 2026-10-25, binary from `7063947`: TASK-021 to TASK-028 were created with `depends_on` on `Done` tasks and with template comments in the frontmatter.
- `internal/docs/docfile/docfile.go:118`: the slug is cut at 60 characters.
- `internal/markdown/frontmatter.go:50`: `SetField` keeps the comment of the line.
- `internal/docs/task/template.md:3`: the template has a comment on each frontmatter field.
- Before this ADR, 30 of 40 titles of tasks and ADRs had more than 40 characters (all 12 ADRs and 18 tasks).
- `internal/docs/task/create.go:27` writes the Field/Value table, and `internal/docs/task/change.go:20` reads `Size` from it with a second parser.
- After the QA cleanup commit `d9ce536`, `Blocks` was wrong in most tasks. Example: TASK-004 gives TASK-009, TASK-014 and TASK-015, and no one of them depends on TASK-004.