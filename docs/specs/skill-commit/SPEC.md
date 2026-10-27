---
slug: skill-commit
title: Commit skill
status: Draft
milestone: v0.4.0
priority: P2
area: [skills, git, plugin]
related: [ADR-003, ADR-010]
---

# Commit skill

## 1. Summary

The `/dx:commit` skill splits the changes of the working tree into logical groups. It makes one commit for each group. It uses the git tools of git-tools to run the git operations.

## 2. Problem

There is no rule for commit messages in the agent work. ADR-003 gives the rule. The commit skill needs a fixed procedure to apply that rule to each group of changes.

## 3. Goals

- The skill makes commits with messages that follow the Conventional Commits rules of ADR-003. The git tools of git-tools check each message (G-4 of git-tools).

Goals for the skill itself are not set in git-tools. See Open Questions.

## 4. Out of Scope

From git-tools, these items apply to this skill:

- Pushing to the remote without a user request. The skill does not push unless the user asks.
- Signing commits.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The /dx:commit skill splits the changes into logical groups, makes one commit for each group with git_commit from the git tools of git-tools, and does not push unless the user asks. | Planned |

## 6. Design

### Planned module

- `internal/plugin/assets/skills/commit/`: the skill. It uses the git tools of git-tools (`git_status`, `git_diff`, `git_stage`, `git_commit`).

### Decisions

- ADR-003: the git tools and the commit message rules.
- ADR-010: the skills and the shared rules.

## 7. Acceptance Criteria
- [R-N: the test or check that proves the requirement.]

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Define the goals of the Commit skill. git-tools gives only requirement R-19 for this skill. | Equal | Open |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |