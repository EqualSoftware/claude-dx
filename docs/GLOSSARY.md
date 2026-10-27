# Glossary

This file defines the terms of the project. Use these terms with these meanings in docs, code, tasks, ADRs and agent briefs. Use `/dx:glossary` to add or change a term.

- **Term**: the approved word. Use it.
- **Do not use**: words with the same meaning. Do not use them, to prevent confusion.
- **Area**: `project`, `process`, `build` or the name of an area of the project.

Sort the terms in alphabetical order.

| Term | Definition | Area | Do not use | Source |
| :--- | :--- | :--- | :--- | :--- |
| ADR | Architecture Decision Record. A file in `docs/decisions/` that records one architecture decision. | process | design doc | `docs/decisions/` |
| ASD-STE100 | Simplified Technical English. The language rule for dx documents, tool arguments, tool results, skills and commit messages. | process | simple English, STE | ADR-010 |
| Brief | The section of a task that says what to do, the files, the constraints and the check commands. | process | — | docs-task#R-2 |
| Cache | The directory `os.UserCacheDir()/claude-dx`. It holds the extracted plugin, `update.json`, `unmapped.log`, the cache lock `cache.lock`, and in `projects/<key>/` the Project file (cache) of each project and the project data directory of each project outside git. | cli | data dir, home dir | ADR-001 |
| Cache lock | The read-write OS lock on `<cache>/cache.lock`. Each claude-dx process that uses the cache holds a shared lock, and the global items of `clean` need the exclusive lock. No process removes the file. | cli | sessions lock, session lock | ADR-006 |
| Definition of Done | The section of a task with the checklist. All items must be checked before the task is `Done`. | process | DoD, done criteria | docs-task#R-13 |
| Docs lock | The read-write OS lock on `docs.lock` in the project data directory, with a read-write mutex of the dx server. Each docs tool that reads holds a shared lock. Each tool that changes a spec, task, ADR, learning, the glossary or a README index holds the exclusive lock for the full read, change and write. It also protects the number counter. | docs-tools | edit lock | ADR-006 |
| dx plugin | The Claude Code plugin named `dx` that the binary embeds and the launcher loads with `--plugin-dir`. It holds the skills, the shared rules and the MCP server declaration. | plugin | extension, add-on | ADR-001, ADR-010 |
| dx server | The MCP server that `claude-dx mcp` starts on stdio. Claude Code starts one for each session. | mcp | daemon, backend | ADR-002 |
| Elicitation | An MCP request from the dx server to the user, through Claude Code, for a direct yes or no answer. The model does not answer it. | mcp | confirmation prompt | ADR-003 |
| Launcher | The part of `claude-dx` that prepares the dx plugin and starts `claude` with the dx flags and the user arguments. | cli | wrapper, shim | ADR-001 |
| Learning | A file in `docs/learnings/<topic>.md` about one topic. It holds one or more lessons. | docs-tools | note, tip | ADR-002 |
| Lesson | A verified, reusable and not obvious technical fact from finished work. `/dx:eureka` records it. | process | insight | ADR-010 |
| Main branch | The branch that `origin/HEAD` names, else `main`, else `master`. | git | trunk, default branch | git-tools#R-8 to R-13 |
| Milestone | The release version of a spec, example `v0.1.0`. A frontmatter field of each spec. | process | target version, release field | docs-spec |
| Owner | The person who decides an open question of a spec, or who is responsible for an ADR. The default is the git user.name. | process | assignee, reviewer | docs-spec |
| Plan | The section of a task with a numbered list of steps. Each step has one action and its check. | process | — | docs-task#R-2, R-12 |
| Project data directory | The directory of the local data of one project: `<git-common-dir>/dx/` in a git repository, else `<cache>/projects/<key>/`. It holds the docs lock, the Project file (data) and the local tasks. All worktrees of a repository share it. | docs-tools | dx folder, project cache | ADR-007 |
| Project file (cache) | The file `<cache>/projects/<key>/project.json`. It holds the project path and the last use date, so `claude-dx clean` can show each project. | cli | — | ADR-007 |
| Project file (data) | The file `project.json` in the project data directory. It holds the number counters. | docs-tools | — | ADR-007 |
| Project key | The first 16 hex characters of the SHA-256 of the git common directory (or of the project root outside git). All worktrees of one repository have the same key. | mcp | project id | ADR-007 |
| Project root | The git top level of the current directory, or the current directory outside git. `Project.Resolve` refuses a path outside it. The filesystem tools will use it. | mcp | workspace, base dir | ADR-002 |
| Ready task | A Todo task whose spec is Approved and whose depends_on tasks are all Done. | process | next task, unblocked task | docs-task#R-8 |
| Requirement | One behavior that a spec requires. Its ID R-N does not change. Its status is Planned or Done. | process | — | ADR-008 |
| Shared rules | The short dx rules text (`rules.md` in the dx plugin) that the launcher gives to Claude Code with `--append-system-prompt-file`. | plugin | RULES.md | ADR-010 |
| Spec | The current definition of one feature or scope, in docs/specs/<slug>/SPEC.md. Its ID is the slug. It has requirements and no task IDs. | process | — | ADR-008, `docs/specs/` |
| Task | A local work item of one spec, with ID `<spec>/<task-slug>`. It is stored in the project data directory, not in git. | process | ticket, issue | docs-task |
| Task status | The state of a task: Todo, In Progress, Blocked or Done. Done is final. | process | — | docs-task#R-5 |
| Unmapped operation | A git operation that no dx git tool covers. The model runs it with `git_exec`, and the dx server logs it to `unmapped.log`. | git | unknown command | ADR-003 |
| Work batch | A group of steps of a task that one agent does in one run. | process | step, chunk | docs-task#R-2 |
