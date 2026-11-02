# ADR-017: Stage hunks by numbered diff token

- Status: Accepted
- Date: 2026-11-02
- Owners: Equal

## Context
git_stage stages whole files only. The /dx:commit skill (skill-commit#R-1) must split one file between commits. A model makes errors when it writes a unified patch: wrong line numbers and wrong context.

## Decision
git_diff takes hunks (default false). With true, it numbers the hunks of each file ([1], [2], ...) and gives a token, a short hash of the diff. It stays read only. Without the option, its output does not change. A new tool, git_stage_hunks, takes the token and a list of paths with hunk numbers, and stages only those hunks. It refuses when the current diff has another token, or when a number does not exist. This ADR adds to ADR-003 and does not replace it.

## Alternatives
- An option on git_stage: Rejected. The user prefers a separate tool. git_stage stays simple.
- A patch text input: Rejected. Errors in line numbers and context.
- A separate read tool git_hunks: Rejected. One more tool. git_diff already reads the diff.
- git_exec: Rejected. Each use is logged as unmapped and asks the user.
- No token check: Rejected. A wrong hunk can go into a commit.

## Consequences
- Positive: The skill can split one file between commits. The hunk numbers are checked against the diff.
- Negative: One more state-changing tool. It asks the user, like git_stage.
- Follow-up: git-tools R-23 and R-24 implement it. skill-commit R-1 uses it.

## Evidence
Verified: docs/specs/git-tools/SPEC.md R-23 and R-24 (added in this session). git_stage today stages whole files only: docs/specs/git-tools/SPEC.md section 6 lists no hunk tool.
