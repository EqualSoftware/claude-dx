---
name: learning
description: Record a verified, reusable technical lesson in docs/learnings with the dx learning tools. Use when work proves a fact about tools, platforms, libraries or the domain that later work needs.
---

# Record a lesson

A learning is a file `docs/learnings/<topic>.md` about one topic. It holds one or more lessons. The dx learning tools read and write these files. They also write the index `docs/learnings/README.md`. Do not change files in `docs/learnings` with Edit or Write. Do not read `docs/learnings/README.md`: it is for persons. Use `learning_list`.

## When a lesson is valid

Record a lesson only if it is:
- **Verified**: a command, a test, a build or the code proves it.
- **Reusable**: it helps later work, not only this task.
- **Not obvious**: a person cannot easily see it in the code or in the history.
- **Not documented**: `CLAUDE.md`, `docs/` and the code comments do not have it. Give a search of more than 3 files to the `dx:scout` agent.

If no lesson is valid, report "No new lessons" and stop. This is a correct result.

Put a lesson in a different place if it fits better:
- It applies to one function or file: a short code comment there.
- It is a project term: `/dx:glossary`.
- It is an architecture decision with alternatives: `/dx:adr`.

## Procedure

1. Run `learning_list`. If a topic has the same subject, read it with `learning_get` and add the lesson with `learning_update`. The new body replaces all of the old body, so keep the old lessons.
2. Else run `learning_create`:
   - `topic`: kebab-case, example `go-embed-paths`.
   - `description`: one line: what the lessons are about and when to use them.
   - `body`: the lesson.
3. Remove a lesson that new evidence proves wrong. Delete a file with no valid lesson with `learning_delete`.

## Rules

- One lesson is one or two sentences. Write in ASD-STE100.
- Do not record guesses, plans or a history of the task.

## Report

For each lesson: the topic and one line with the lesson. Or "No new lessons".
