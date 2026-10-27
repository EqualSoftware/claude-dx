---
slug: resume-command
title: Resume command and SessionStart hook
status: Approved
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
- A history of the work for projects with no git repository. It needs a new ADR and its own spec (Q-4).
- Spec and task tools (docs-spec, docs-task).

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | `claude-dx resume` prints a header line, then the sections Changed, In Progress, Next, Blocked and Unpushed, in this order. It does not print an empty section. Header: "dx: branch <branch>, <N> ahead of <upstream>, <M> changed files"; with no upstream, "no upstream" replaces "<N> ahead of <upstream>"; with a detached HEAD, "dx: detached at <short hash>" and no Unpushed section. | Done |
| R-2 | It uses no network and no MCP server. | Done |
| R-3 | The plugin has a `SessionStart` hook with the matchers `startup`, `resume`, `clear` and `compact` that runs it. The binary path in the hook command is in double quotes. | Done |
| R-4 | With no In Progress, Next or Blocked task, no unpushed commit and a clean tree, it prints only the line "dx: branch <branch>, clean tree, no active tasks." Done tasks do not count. | Done |
| R-5 | Changed: one line with the `git status` short codes and paths, separated by commas, max 10 files. | Done |
| R-6 | In Progress: all In Progress tasks, newest modification time first. Each task has its ID, priority and title, the first unchecked Plan step, and the last 3 non-empty lines of Notes (sub-items included). | Done |
| R-7 | Next: Todo tasks of Approved specs and of `general` with all `depends_on` tasks Done, ordered by priority and then ID, max 3. Blocked: Blocked tasks with ID, priority, title and the last non-empty line of Notes, max 5. | Done |
| R-8 | Unpushed: the commits of the current branch after its upstream or, with no upstream, the commits that no remote branch has. Each line has the short hash, the subject and the `Refs:` trailers in brackets, max 10. The command uses the local remote refs and does not fetch. | Done |
| R-9 | The output has 40 lines or less. A cut list ends with "+N more". When the output is still more than 40 lines, the command cuts the Notes lines first. | Done |
| R-10 | In a directory that is not in a git repository, the header is "dx: no git repository" and the output has only the In Progress, Next and Blocked sections. A project with no data directory has no tasks and gets no error line. For the other errors (git not installed, a data directory that cannot be read, a task file that is not valid), the command skips the part that fails and prints one line "dx: skipped <part>: <error>" for each error. The part is git, tasks or the task ID. | Done |
| R-11 | The command always exits with code 0. It only reads: it does not create the data directory, does not change `project.json` and writes no other file. | Done |
| R-12 | The help of `claude-dx` has a line for `claude-dx resume`, aligned with the other commands, with the text "Print the work state (tasks and git) for a new session." | Done |

## 6. Design

### Planned modules

- `internal/resume/`: reads the task files, their modification time and git, and writes the state text.
- `internal/cli/` and `cmd/claude-dx/`: the `resume` command and its help line.
- `internal/plugin/assets/hooks/hooks.json`: the `SessionStart` hook with the matchers `startup`, `resume`, `clear` and `compact`, with `{{BINARY}} resume`. The plugin has no hook file now. `internal/plugin/plugin.go:79` replaces `{{BINARY}}` in the plugin files, as for `.mcp.json`.

### Data flow (planned)

1. The hook runs `claude-dx resume` in the project directory.
2. The command reads the tasks in the project data directory (ADR-007) and the git state. It uses no network and no MCP server.
3. It prints the state. Claude Code adds the output to the context of the session.

### Example output

```
dx: branch feat-x, 2 ahead of origin/feat-x, 3 changed files
Changed: M internal/a.go, ?? internal/b.go, M go.mod
In Progress:
- resume-command/cli (P1) Resume command
  Next step: Add the hook file
  Notes: Done: parser and help.
Next:
- git-tools/hook (P1) Bash git hook
Unpushed:
- abc1234 feat(cli): add resume [resume-command#R-1]
```

### Decisions

- ADR-008: the resume command, the hook, and the rule that no journal is kept (Q-4).

## 7. Acceptance Criteria

- R-1, R-4 to R-9: golden tests of the output in a temporary git repository with a temporary data directory: clean tree, changed files, upstream, no upstream, detached HEAD, each task status, dependencies, the limits and "+N more", and the 40-line cut.
- R-2: the command does not start the MCP server and does not run `git fetch`; the tests run with no network.
- R-3: plugin test that the extracted plugin has `hooks/hooks.json` with the 4 matchers and the absolute binary path.
- R-10: tests for a directory with no git repository, no data directory, and a task file that is not valid.
- R-11: test that the exit code is 0 in each error case and that the data directory is not created.
- R-12: test of the help text.
- Tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Output format: which layout does the command print? Recommended: plain text with one header line (branch, upstream, ahead count, number of changed files) and then the sections Changed, In Progress, Next, Blocked and Unpushed. It does not print an empty section. | Equal | Yes. Plain text: one header line (branch, upstream, ahead count, number of changed files), then the sections Changed, In Progress, Next, Blocked and Unpushed. The command does not print an empty section. |
| Q-2 | No upstream: what does the command print when the branch has no upstream? Recommended: the header shows "no upstream", and the Unpushed section lists the commits that no remote branch has (git log --branches --not --remotes, current branch only). | Equal | Yes. The header shows "no upstream". The Unpushed section lists the commits of the current branch that no remote branch has. |
| Q-3 | Limits for 40 lines: how does the command shorten the output? Recommended: Changed files max 10, Next max 3, Blocked max 5, Unpushed max 10, Notes of an In Progress task: the last 3 lines. Each cut list ends with "+N more". When the total is still more than 40 lines, the command cuts the Notes first. | Equal | Yes. Changed max 10, Next max 3, Blocked max 5, Unpushed max 10, Notes of an In Progress task: the last 3 lines. Each cut list ends with "+N more". When the total is more than 40 lines, the command cuts the Notes first. |
| Q-4 | Errors and side effects: what does the command do when git is missing, the directory is not a git repository, the project has no data directory, or a task file is not valid? Recommended: it skips the part that fails, prints one line for each error, and always exits with code 0, so the hook does not stop the session. It only reads: it does not create the data directory and does not change project.json. | Equal | Keep ADR-008: no journal. In a project with no git repository, the header is "dx: no git repository" and the output has only the task sections (In Progress, Next, Blocked). For the other errors (git not installed, no data directory, task file not valid), the command skips the part that fails and prints one line for each error. It always exits with code 0. It only reads: it does not create the data directory and does not change project.json. A history for projects with no git is a possible later spec with a new ADR. |
| Q-5 | Next tasks: which Todo tasks are in Next, and in which order? Recommended: Todo tasks of Approved specs and general with all depends_on Done, ordered by priority and then ID (as task_list), max 3. | Equal | Yes. Todo tasks of Approved specs and of general, with all depends_on Done. Order: priority, then ID. Max 3. |
| Q-6 | R-4 text: what is the one line when there are no tasks to show and the tree is clean? Recommended: "dx: branch <branch>, clean tree, no active tasks." And does "no tasks" mean no In Progress, Next and Blocked tasks (Done tasks do not count)? Recommended: yes. | Equal | Yes. The line is "dx: branch <branch>, clean tree, no active tasks." No tasks means no In Progress, Next and Blocked tasks. Done tasks do not count. |
| Q-7 | Line formats: Recommended: (a) Changed is one line with the git status short codes, separated by commas. (b) In Progress is ordered by modification time, newest first. (c) A Blocked line has the ID, the priority, the title and the last line of Notes (the blocker). (d) An Unpushed line has the short hash, the subject and the Refs trailers in brackets. (e) A Notes line is one non-empty line of the Notes section, sub-items included. | Equal | Yes. (a) Changed is one line with the git status short codes, separated by commas. (b) In Progress is ordered by modification time, newest first. (c) A Blocked line has the ID, the priority, the title and the last line of Notes. (d) An Unpushed line has the short hash, the subject and the Refs trailers in brackets. (e) A Notes line is one non-empty line of the Notes section, sub-items included. |
| Q-8 | Detached HEAD: what does the header show when HEAD is not on a branch (for example during a rebase)? Recommended: "dx: detached at <short hash>", with no upstream and no Unpushed section. | Equal | Yes. The header is "dx: detached at <short hash>", with no upstream and no Unpushed section. |
| Q-9 | Help text: which line does the help of claude-dx show for the resume command? Recommended: "resume   Print the work state (tasks and git) for a new session." | Equal | Yes. The help line is "resume   Print the work state (tasks and git) for a new session." |
| Q-10 | Error line text (R-10): which text does the command print for a part that fails? Recommended: "dx: skipped <part>: <error>", where part is git, tasks or the task ID (example: "dx: skipped tasks: no data directory"). | Equal | Yes. The text is "dx: skipped <part>: <error>". The part is git, tasks or the task ID. Example: "dx: skipped tasks: no data directory". |

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
| 2026-10-27 | Equal | R-1: text changed. |
| 2026-10-27 | Equal | R-4: text changed. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | Added R-10. |
| 2026-10-27 | Equal | Added R-11. |
| 2026-10-27 | Equal | Added R-12. |
| 2026-10-27 | Equal | R-10: text changed. |
| 2026-10-27 | Equal | Refined with the user: Q-1 to Q-10 answered, R-1 to R-12. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | R-4: text changed. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 5. |
| 2026-10-27 | Equal | R-10: text changed. |
| 2026-10-27 | Equal | R-12: text changed. |
| 2026-10-27 | Equal | R-3: text changed. |
| 2026-10-27 | Equal | Approved again by the user after the changes of R-3, R-4, R-10 and R-12. |
| 2026-10-27 | Equal | R-1, R-2, R-3, R-4, R-5, R-6, R-7, R-8, R-9, R-10, R-11, R-12: Done. |