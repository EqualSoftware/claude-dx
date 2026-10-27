---
slug: docs-learnings
title: Learning documents and learning tools
status: Approved
milestone: v0.1.0
priority: P1
area: [docs, mcp]
related: [ADR-002]
---

# Learning documents and learning tools

## 1. Summary

A learning is one file, `docs/learnings/<topic>.md`, about one topic. The learning tools list, read, create, change and delete learnings. The learning skill tells when a lesson is valid to record.

## 2. Problem

Later work needs verified facts about tools, platforms and libraries. Without a list tool and a duplicate check, a person or the model can record the same topic twice, or miss a lesson.

## 3. Goals

- G-1: `learning_create` refuses a topic that exists.
- G-2: One bad learning file does not hide the other learnings.
- G-3: The README index is not a learning.

## 4. Out of Scope

- The index format: docs-common.
- Lesson review after a record: not part of this spec.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | learning_list returns one line for each learning as topic and description, and skips the README.md index. | Done |
| R-2 | learning_list gives one warning line for each learning file that it cannot read or parse, and lists the other learnings. | Done |
| R-3 | learning_get returns the full file of one learning. | Done |
| R-4 | learning_create writes docs/learnings/<topic>.md with the frontmatter topic and description, and refuses a topic that is not kebab-case, a topic that exists, and the topic readme. | Done |
| R-5 | learning_update changes the description or the body of a learning, and learning_delete removes it. | Done |
| R-6 | The learning skill records a lesson only when it is verified, reusable, not obvious and not documented, and it runs learning_list before learning_create. | Done |

## 6. Design

### Modules

- `internal/docs/learning/learning.go`: list, get, create, update and delete.
- `internal/docs/learning/index.go`: the generated `README.md` of `docs/learnings`.
- `internal/plugin/assets/skills/learning/SKILL.md`: the learning skill.

### Data flow

Each write holds the docs lock and uses the write rule of docs-common. Then the tool writes the index. `learning_list` skips `README.md`.

### Formats

- File `docs/learnings/<topic>.md`. The topic is kebab-case.
- Frontmatter: `topic` and `description`.
- Index table: `| Topic | Description |`, in name order.

### Decisions

- ADR-002: document tools and paths.

## 7. Acceptance Criteria

- R-1: `TestToolsFlow` and `TestIndex` (`internal/docs/learning`).
- R-2: `TestListBadFrontmatter` (`internal/docs/learning`).
- R-3: `TestToolsFlow` (`internal/docs/learning`).
- R-4: `TestCreate`, `TestMissingDir` and `TestCreateRefusesReadme` (`internal/docs/learning`).
- R-5: `TestUpdateNeedsAChange` and `TestToolsFlow` (`internal/docs/learning`).
- R-6: Not automated. The skill text is checked by review.

The tests pass on macOS, Linux and Windows.

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
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |