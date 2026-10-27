---
slug: git-tools
title: Git tools and the Bash git hook
status: Draft
milestone: v0.2.0
priority: P1
area: [git, mcp, plugin]
related: [ADR-003, ADR-010]
---

# Git tools and the Bash git hook

## 1. Summary

claude-dx gives the agent git tools in place of the git command in the Bash tool. The tools use the system git binary. A Bash hook blocks git in Bash, and a git_exec tool records each operation that no tool maps. The skills /dx:commit and /dx:eureka use the tools.

## 2. Problem

The agent runs git in Bash, and it can rewrite history, push a force or discard changes with no check. There is no rule for commit messages (ADR-003). The unmapped git operations have no record, so the project cannot see which tools are missing.

## 3. Goals

- G-1: Each git operation of the agent goes through a tool that checks its input.
- G-2: A tool refuses a change that is unsafe, and it asks the user before a change to the remote history.
- G-3: A git operation with no tool is recorded for a report.
- G-4: The commit messages follow the Conventional Commits rules of ADR-003.

## 4. Out of Scope

- A git GUI or a web interface for the repository.
- Git operations that change the remote without a user request (for example a push of a tag that the user did not ask for).
- Signing commits.
- Hosting-service APIs (issues, pull requests). Only the issue form of claude-dx report is opened in the browser.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The git runner runs the system git in the project root with LC_ALL=C, GIT_TERMINAL_PROMPT=0, no pager and no color. A git error gives a short tool error with the first lines of stderr. | Planned |
| R-2 | The read tools git_status, git_diff, git_log, git_show and git_blame return grouped status, file stats or diffs, one line for each commit, and one line for each source line. | Planned |
| R-3 | doctor checks that git is in PATH and that its version is 2.30 or later. It reports FAIL with the minimum version otherwise. | Planned |
| R-4 | The local tools git_stage, git_unstage, git_discard, git_commit, git_amend, git_undo, git_fixup, git_revert, git_cherry_pick, git_tag and git_stash run their git operation. git_discard is marked destructive in its description. | Planned |
| R-5 | git_commit builds the commit message of ADR-003 and refuses a message that is not valid, such as a summary that is not lowercase imperative or that ends with a period. It refuses when nothing is staged, and it adds one Co-Authored-By trailer for each co-author. | Planned |
| R-6 | A commit is on the remote when a remote-tracking branch contains it (git branch -r --contains). | Planned |
| R-7 | A tool that rewrites a commit on the remote asks the user with elicitation first. If the client has no elicitation or the user says no, the tool refuses with the reason and does not change the repository. | Planned |
| R-8 | git_branch runs the actions list, create, switch, rename and delete. delete uses git branch -d, and a branch that is not merged needs elicitation. | Planned |
| R-9 | git_sync runs fetch --prune and rebase --autostash onto the upstream, and keeps the local changes. On a conflict it returns the result of git_conflicts. | Planned |
| R-10 | git_rebase runs the actions start, continue, abort and skip. continue uses GIT_EDITOR=true. | Planned |
| R-11 | git_conflicts returns each conflicted path with its conflict hunks and line numbers. | Planned |
| R-12 | git_push uses --force-with-lease and --force-if-includes when the remote branch is not an ancestor. A push to the main branch in that case needs elicitation first, and nothing is pushed on a no. | Planned |
| R-13 | Network git commands stop after 120 seconds, with GIT_TERMINAL_PROMPT=0, so a credential prompt never blocks the server. | Planned |
| R-14 | A PreToolUse hook denies a Bash call when a git command runs in it, also after ;, &&, \|\|, pipes, subshells, sh -c, eval and wrappers such as env and xargs. The deny reason says to use the git_* tools. | Planned |
| R-15 | If the hook cannot parse a Bash command, it allows the command. | Planned |
| R-16 | git_exec runs git with the given arguments for an operation that no tool maps. Each call appends one JSON line to <cache>/unmapped.log. | Planned |
| R-17 | claude-dx report groups the lines of unmapped.log by git subcommand, opens the issue form in the browser with the values filled in, and prints the URL. report --clear empties the log. | Planned |
| R-18 | The shared rules say to use the git_* tools for git and not to run git in Bash. | Planned |
| R-19 | The /dx:commit skill splits the changes into logical groups, makes one commit for each group with git_commit, and does not push unless the user asks. | Planned |
| R-20 | The /dx:eureka skill keeps only verified, reusable and undocumented lessons, and records each one as an ADR, a learning or a glossary term. It reports No new lessons and changes no file when there is none. | Planned |

## 6. Design

### Planned modules

- `internal/git/`: the runner (environment, timeout, error text) and the tools of each group: read, local, branch and remote.
- `internal/hook/`: the Bash command parser for the PreToolUse hook. It parses the lists, pipes, subshells, `sh -c` strings and the wrappers. It fails open (allows) on a parse error.
- `internal/cli/`: the `hook pre-tool-use` and `report` commands.
- `internal/plugin/assets/hooks/hooks.json`: the PreToolUse hook for the Bash tool, with `{{BINARY}} hook pre-tool-use`.
- `internal/plugin/assets/skills/commit/` and `skills/eureka/`: the two skills.

### Data flow

1. An agent calls a git_* tool. The tool runs the git runner in the project root.
2. A tool that rewrites the remote history asks the user with elicitation. If the answer is no, the tool stops.
3. The Bash hook checks each Bash command. A command with git is denied with the reason of R-14.
4. git_exec appends one line to `<cache>/unmapped.log`. `claude-dx report` reads that file.

### Formats

- `<cache>/unmapped.log`: one JSON line for each git_exec call.
- The commit message: the type, the scope, the summary and the trailers (ADR-003).

### Decisions

- ADR-003: the git tools.
- ADR-010: the skills and the rules.

## 7. Acceptance Criteria

Done requirements: none.

Planned requirements (tests to write):
- R-1: test the runner environment (LC_ALL=C, GIT_TERMINAL_PROMPT=0) and the error text with the first lines of stderr.
- R-2: tests on a temporary repository with staged and untracked files, for the groups of git_status and the outputs of git_diff, git_log, git_show and git_blame.
- R-3: test `doctor` with a git below 2.30 (FAIL) and with a valid git (OK).
- R-4: a test that each local tool runs its git operation on a temporary repository.
- R-5: a summary `Added X.` is refused, and each co-author gives one trailer.
- R-6: a test that the remote check finds a pushed commit with `git branch -r --contains`.
- R-7: a test with a pushed HEAD and a user that says no (git_undo does not change HEAD), and a client without elicitation.
- R-8 and R-10 to R-13: tests on a temporary repository with a remote, for branch, sync, rebase, conflicts and push (an edit that is rebased keeps the local changes, and a push with no answer changes nothing), and a test of the 120 second timeout (R-13).
- R-14 and R-15: hook tests with a command that runs git inside `cd sub && bash -c "git push"` (denied), and `gh pr list` (allowed), and a test that a command that the parser cannot read is allowed.
- R-16: a test that git_exec adds one line to unmapped.log.
- R-17: a test of the grouping and the output of `claude-dx report`, with the browser call replaced by a function.
- R-18: a test that the shared rules have the git line.
- R-19 and R-20: a manual check of the transcript of /dx:commit (two commits with valid headers) and of /dx:eureka (No new lessons).

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | Added R-10. |
| 2026-10-27 | Equal | Added R-11. |
| 2026-10-27 | Equal | Added R-12. |
| 2026-10-27 | Equal | Added R-13. |
| 2026-10-27 | Equal | Added R-14. |
| 2026-10-27 | Equal | Added R-15. |
| 2026-10-27 | Equal | Added R-16. |
| 2026-10-27 | Equal | Added R-17. |
| 2026-10-27 | Equal | Added R-18. |
| 2026-10-27 | Equal | Added R-19. |
| 2026-10-27 | Equal | Added R-20. |
| 2026-10-27 | Equal | v0.2.0 work is planned again after the migration. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | R-5: text changed. |