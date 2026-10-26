# Glossary

This file defines the terms of the project. Use these terms with these meanings in docs, code, tasks, ADRs and agent briefs. Use `/dx:glossary` to add or change a term.

- **Term**: the approved word. Use it.
- **Do not use**: words with the same meaning. Do not use them, to prevent confusion.
- **Area**: `project`, `process`, `build` or the name of an area of the project.

Sort the terms in alphabetical order.

| Term | Definition | Area | Do not use | Source |
| :--- | :--- | :--- | :--- | :--- |
| ADR | Architecture Decision Record. A file in `docs/decisions/` that records one architecture decision. | process | design doc | `docs/decisions/` |
| ASD-STE100 | Simplified Technical English. The language rule for dx documents, tool arguments, tool results, skills and commit messages. | process | simple English, STE | ADR-009 |
| Cache | The directory `os.UserCacheDir()/claude-dx`. It holds the extracted plugin, LSP servers, `update.json`, `unmapped.log`, the cache lock `cache.lock`, and the counters and locks of each project in `projects/<key>/`. | cli | data dir, home dir | ADR-001 |
| Cache lock | The read-write OS lock on `<cache>/cache.lock`. Each claude-dx process that uses the cache holds a shared lock, and `cache clean` needs the exclusive lock. No process removes the file. | cli | sessions lock, session lock | ADR-015 |
| Definition of Done | The checklist in the last section of a task. All items must be checked, with evidence, before the task is `Done`. | process | DoD, done criteria | `docs/tasks/` |
| Docs lock | The read-write OS lock on `<cache>/projects/<key>/docs.lock`, with a read-write mutex of the dx server. Each docs tool that reads holds a shared lock. Each tool that changes a task, ADR, learning, the glossary or a README index holds the exclusive lock for the full read, change and write. It also protects the number counter. | docs-tools | edit lock | ADR-007 |
| dx plugin | The Claude Code plugin named `dx` that the binary embeds and the launcher loads with `--plugin-dir`. It holds the skills, the scout agent, the hooks and the MCP server declaration. | plugin | extension, add-on | ADR-001 |
| dx server | The MCP server that `claude-dx mcp` starts on stdio. Claude Code starts one for each session. | mcp | daemon, backend | ADR-002 |
| Elicitation | An MCP request from the dx server to the user, through Claude Code, for a direct yes or no answer. The model does not answer it. | mcp | confirmation prompt | ADR-008 |
| Launcher | The part of `claude-dx` that prepares the dx plugin and starts `claude` with the dx flags and the user arguments. | cli | wrapper, shim | ADR-003 |
| Learning | A file in `docs/learnings/<topic>.md` about one topic. It holds one or more lessons. | docs-tools | note, tip | ADR-006 |
| Lesson | A verified, reusable and not obvious technical fact from finished work. `/dx:eureka` records it. | process | insight | ADR-009 |
| Main branch | The branch that `origin/HEAD` names, else `main`, else `master`. | git | trunk, default branch | TASK-011 |
| Milestone | The release version of a task, example `v0.1.0`. A frontmatter field of each task. | process | target version, release field | ADR-013 |
| Owner | The git `user.name` of the person who creates a task. A frontmatter field of each task. | process | assignee, reviewer | ADR-013 |
| Project data directory | The directory of the local data of one project: `<git-common-dir>/dx/` in a git repository, else `<cache>/projects/<key>/`. It holds the docs lock, the Project file (data) and the local tasks. All worktrees of a repository share it. | docs-tools | dx folder, project cache | ADR-015 |
| Project file (cache) | The file `<cache>/projects/<key>/project.json`. It holds the project path and the last use date, so `claude-dx clean` can show each project. | cli | — | ADR-018 |
| Project file (data) | The file `project.json` in the project data directory. It holds the number counters. | docs-tools | — | ADR-018 |
| Project key | The first 16 hex characters of the SHA-256 of the git common directory (or of the project root outside git). All worktrees of one repository have the same key. | mcp | project id | ADR-007 |
| Project root | The git top level of the current directory, or the current directory outside git. Tools refuse paths outside it. | mcp | workspace, base dir | ADR-005 |
| Ready task | An open task that is not `Blocked` and has all dependencies `Done`. `task_list(status: ready)` gives these tasks. | process | next task, unblocked task | ADR-013 |
| Scout | The read-only agent `dx:scout` on Haiku. Skills give it fact searches of more than 3 files. | plugin | explorer, searcher | ADR-009 |
| Shared rules | The short dx rules text (`rules.md` in the dx plugin) that the launcher gives to Claude Code with `--append-system-prompt-file`. | plugin | RULES.md | ADR-009 |
| Task | A spec file in `docs/tasks/` for one unit of work. | process | ticket, issue | `docs/tasks/` |
| Unmapped operation | A git operation that no dx git tool covers. The model runs it with `git_exec`, and the dx server logs it to `unmapped.log`. | git | unknown command | ADR-008 |
| Work batch | A group of steps of a task that one agent does in one run. | process | step, chunk | `docs/tasks/` |
