---
name: commit
description: Split the changes of the working tree into logical groups and make one Conventional Commit for each group with the dx git tools. Use when the user asks to commit the changes.
---

# Commit the changes in groups

Use the dx git tools. Do not run git in Bash. Pass `path` when you work in a worktree.

## Steps

1. Read all changes (staged, not staged and untracked) with `git_status` and `git_diff`. Read the staged diff and the unstaged diff.
2. Unstage all changes with `git_unstage`. The skill groups all changes again and does not keep the staged state. Then run `git_diff` with `hunks: true` to see the hunks of each file. It works only on the unstaged diff, so run it after the unstage.
3. Split the changes into logical groups. A group holds whole files or some hunks of a file. Each change is in exactly one group.
4. Set the message of each group: type, scope, summary and body (rules: ADR-003, Conventional Commits). Add the Refs trailer (see Refs).
5. Show the plan. For each group give the files, the hunks and the full message (type, scope, summary, body and Refs).
6. Wait for the approval of the user. Make no commit before the approval. Apply the corrections of the user, show the new plan and wait again.
7. For each group, in order:
   - Whole files, also untracked files, binary files, mode-only changes, renames and deletes: stage them with `git_stage`. `git_stage_hunks` refuses these files.
   - Part of a file: run `git_diff` with `hunks: true` to get a new token. The first line is `Token: <12 hex>`. Each hunk line has a number, example `[2] @@ -10,3 +10,4 @@`. The numbers restart for each file. Run `git_stage_hunks` with the token and `hunks: [{path, numbers}]`. The token covers the full unstaged diff: get a new token before each `git_stage_hunks`, because each stage and each commit changes the diff.
   - Run `git_commit` with the message of the group.
8. Run `git_status` to check that no change is left. Give the user the commits.

## Refs

- Find the requirements from the spec and from the task In Progress: run `task_list` with the status `In Progress` and `spec_get` with section 5.
- Put `Refs: <slug>#R-N` in the body of the group, one for each requirement. Example: `Refs: skill-commit#R-1, skill-commit#R-2`.
- If the requirements of a group are not clear, ask the user.
- A group with no related spec gets no trailer.

## Rules

- Do not run the acceptance check. The author of the change runs it before.
- Do not push. Push only when the user asks.
- Write the messages in ASD-STE100.
