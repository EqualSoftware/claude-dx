# ADR-003: Git through MCP tools, no merge

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
Git commands through Bash change between shells and platforms (Windows uses PowerShell or Git Bash). The project must use rebase and Conventional Commits. `go-git` does not support rebase.

## Decision
- The tools run the `git` binary of the system. `claude-dx doctor` will check the minimum version (git-tools#R-3). Today it checks only the Claude Code version.
- Tools: `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, `git_stage`, `git_unstage`, `git_discard`, `git_commit`, `git_amend`, `git_undo`, `git_fixup`, `git_branch`, `git_sync`, `git_rebase`, `git_conflicts`, `git_push`, `git_stash`, `git_cherry_pick`, `git_revert`, `git_tag`, `git_exec`.
- Read-only tools (`git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`) run with no prompt. Tools that change state ask the user, unless the user allows them in Claude Code.
- `git_commit` takes `type`, `scope`, `summary`, `body`, `breaking`, `paths` and `co_authors`. It builds and validates a Conventional Commit:
  - Types: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `build`, `ci`, `style`, `chore`, `revert`.
  - Scope is optional. Summary is imperative, lowercase, with no period. The header has a maximum of 72 characters.
  - `breaking: true` adds `!` and a `BREAKING CHANGE:` footer.
  - `co_authors` adds `Co-Authored-By:` trailers. The model sends them when the project uses them.
- `git_undo` runs `reset --soft HEAD~n` and keeps the changes staged.
- `git_sync` runs fetch and rebase onto the upstream, with autostash.
- Safety rules:
  - No merge. No tool runs `git merge`.
  - No plain `--force`. `git_push` uses `--force-with-lease` only after a history change.
  - `git_push` refuses a force push to the main branch (from `origin/HEAD`).
  - `git_undo`, `git_amend` and `git_fixup` refuse to change commits that are on the remote, and suggest `git_revert`.
  - The user can authorize an exception. The server asks the user directly with MCP elicitation. If the client does not support elicitation, the server refuses.
- A `PreToolUse` hook (`claude-dx hook pre-tool-use`) blocks all `git` commands in the Bash tool, also inside `sh -c`, `bash -c`, `&&`, `;`, `|` and subshells. It does not block `gh`. The block message tells the model to use the `git_*` tools, or `git_exec` if no tool fits.
- `git_exec` runs a git command that no tool covers. It takes `args` and `reason`. Its result tells the model to tell the user that the operation is not mapped, with the URL to open an issue. It adds the args and the reason to `<cache>/unmapped.log`.
- `claude-dx report` opens a prefilled GitHub issue in the browser from `unmapped.log`. Nothing is sent without the user.

## Alternatives
- `go-git`: Rejected. No rebase.
- Permission deny rules (`Bash(git:*)`): Rejected. `cd x && git` and `sh -c` bypass them.
- No block: Rejected. The model uses Bash git and the result is not consistent.

## Consequences
- Positive: same git behavior on all platforms. Safe defaults. Unmapped cases are visible and can become new tools.
- Negative: git must be installed. This is the only external dependency.
- Follow-up: spec git-tools verifies MCP elicitation support in Claude Code (git-tools#R-7).

## Evidence
- Not built yet. The spec is docs/specs/git-tools/SPEC.md, and its status is Draft.
- cmd/claude-dx/main.go:66 answers `report` and `hook` with "is not available yet". internal/cli/cli.go:20-22 reserves the names.
- internal/launcher/launcher.go:79-81: the settings allow all tools of the server now. The ask list for the git tools comes with them.
- internal/project/project.go:58-69 and internal/docs/spec/commits.go:25-30 run `git` directly until the git runner of the spec exists.
