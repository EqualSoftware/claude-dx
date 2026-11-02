---
slug: skill-commit
title: Commit skill
status: Approved
milestone: v0.4.0
priority: P2
area: [skills, git, plugin]
related: [ADR-003, ADR-010, ADR-017]
---

# Commit skill

## 1. Summary

The `/dx:commit` skill splits the changes of the working tree into logical groups. It makes one commit for each group. It uses the git tools of git-tools to run the git operations.

## 2. Problem

There is no rule for commit messages in the agent work. ADR-003 gives the rule. The commit skill needs a fixed procedure to apply that rule to each group of changes.

## 3. Goals

- The skill splits all changes into logical groups by file or by hunk.
- The skill shows the plan and waits for the approval of the user.
- The skill makes one Conventional Commit for each group.
- The skill adds the Refs trailers of the requirements.
- The skill does not push.
- The skill makes commits with messages that follow the Conventional Commits rules of ADR-003. The git tools of git-tools check each message (G-4 of git-tools).

## 4. Out of Scope

From git-tools, these items apply to this skill:

- Pushing to the remote without a user request. The skill does not push unless the user asks.
- Signing commits.
- Running the acceptance check or other checks before the commit.
- Keeping the staged state of the user: the skill groups all changes again.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The /dx:commit skill unstages all changes and splits all changes (staged, not staged and untracked) into logical groups. A group holds whole files or some hunks of a file (git_diff with hunks: true, git_stage_hunks of git-tools). | Done |
| R-2 | Before it commits, the /dx:commit skill shows each group with its files, its hunks and its full commit message. It makes no commit until the user approves the plan. It applies the corrections of the user. | Done |
| R-3 | The /dx:commit skill makes one commit for each group with git_commit of git-tools. It does not run the acceptance check and does not push unless the user asks. | Done |
| R-4 | The /dx:commit skill adds Refs: <slug>#R-N trailers to the body from the spec and the task In Progress. It asks the user when the requirements of a group are not clear. A group with no related spec gets no trailer. | Done |

## 6. Design

### Planned module

- `internal/plugin/assets/skills/commit/`: the skill. It uses the git tools of git-tools: `git_status`, `git_diff` (with `hunks: true`), `git_unstage`, `git_stage`, `git_stage_hunks`, `git_commit`.

### Flow

1. Read the changes with `git_status` and `git_diff` (`hunks: true`).
2. Unstage all changes with `git_unstage`.
3. Group the changes by file or by hunk, and add the Refs trailers.
4. Show the plan: each group with its files, hunks and full message.
5. Wait for the approval or the corrections of the user.
6. For each group, stage it (`git_stage` or `git_stage_hunks`) and commit it (`git_commit`).

### Decisions

- ADR-003: the git tools and the commit message rules.
- ADR-010: the skills and the shared rules.

## 7. Acceptance Criteria

- R-1: a test that the embedded skill file `internal/plugin/assets/skills/commit/SKILL.md` has the unstage-all step and names `git_diff` with `hunks: true` and `git_stage_hunks`. Plus a manual run of the skill on a test repository with staged, unstaged and untracked changes: each change is in exactly one group.
- R-2: a read check of the embedded skill file: the plan is shown and the skill waits for the approval before the first `git_commit`. Plus a manual run: no commit exists before the approval.
- R-3: a read check of the embedded skill file: the skill names `Refs: <slug>#R-N` and the rule for a group with no related spec. Plus a manual run: each commit body has the expected trailer.
- R-4: a read check of the embedded skill file: the skill does not run the acceptance check and does not call `git push`. Plus a manual run: `git log` shows the commits and no push is sent.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Define the goals of the Commit skill. git-tools gives only requirement R-19 for this skill. | Equal | Goals: the skill splits all changes into logical groups by file or by hunk, shows the plan and waits for the approval of the user, makes one Conventional Commit for each group, adds the Refs trailers of the requirements, and does not push. |
| Q-2 | How does the skill split the changes: by file or by hunk? | Equal | By hunk. A group can hold whole files or some hunks of a file. The skill uses git_diff with hunks: true and the new tool git_stage_hunks of git-tools. |
| Q-3 | Does the skill show the plan before it commits? | Equal | Yes. It shows each group with its files, hunks and full message, and waits for the approval or the corrections of the user. It makes no commit before the approval. |
| Q-4 | How does the skill set the Refs trailer? | Equal | It finds the requirements from the spec and the task In Progress and adds Refs: <slug>#R-N to the body. It asks the user when the requirement of a group is not clear. A group with no related spec gets no trailer. |
| Q-5 | What does the skill do with changes that are staged before it starts? | Equal | It unstages all changes and groups all of them again. |
| Q-6 | Does the skill run the acceptance check? | Equal | No. The author of the change runs it before. The skill only groups and commits. |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-11-02 | Equal | R-1: text changed. |
| 2026-11-02 | Equal | Added R-2. |
| 2026-11-02 | Equal | Added R-3. |
| 2026-11-02 | Equal | Added R-4. |
| 2026-11-02 | Equal | Changed fields: related. |
| 2026-11-02 | Equal | Status: In Review. |
| 2026-11-02 | Equal | The user approved the spec. |
| 2026-11-02 | Equal | R-1, R-2, R-3, R-4: Done. |