# ADR-015: Git tools in worktrees

- Status: Accepted
- Date: 2026-11-02
- Owners: Equal

## Context
The dx-agents workflow creates worktrees, and the workers commit in them. Related: ADR-003. ADR-003 runs git only in the project root. The Bash hook blocks git in Bash, so a worker has no git path for its worktree.

## Decision
Each git_* tool takes an optional path. The path must be the project root or a worktree of the same repository. The tool runs git in that directory. Another path gives a tool error (spec git-tools R-21). git_worktree runs the actions list, add and remove. add creates a new branch from a start point. remove refuses a worktree with changes that are not committed. With delete_branch it also deletes the branch with git branch -d (spec git-tools R-22). git_worktree add does not fix the .claude/worktrees/ directory, because that is a dx-agents convention. The caller gives the path. remove with delete_branch uses git branch -d, as R-8.

## Alternatives
- Use git_exec for worktrees: Rejected. Worktrees are frequent, and each git_exec call adds a line to unmapped.log (spec git-tools R-16).

## Consequences
- Positive: Workers commit in their worktrees through the git_* tools. The path check keeps each tool in the project repository.
- Negative: Each tool must validate the path.
- Follow-up: Implement R-21 and R-22 with their tests. ADR-003 stays unchanged.

## Evidence
Spec git-tools R-14: the Bash hook denies git in Bash. Spec git-tools R-16: git_exec appends one line to unmapped.log for each call. Spec git-tools R-21 and R-22 (added in this run).
