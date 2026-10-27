---
slug: docs-spec
title: Spec documents and spec tools
status: Approved
milestone: v0.1.0
priority: P1
area: [docs, mcp]
related: [ADR-002, ADR-008]
---

# Spec documents and spec tools

## 1. Summary

A spec is one file, `docs/specs/<slug>/SPEC.md`. It gives the current definition of one feature, with its requirements and their status. The `spec_*` tools read, create and change specs and their requirements.

## 2. Problem

Task files in git make many files, and they do not give the current behavior of a feature. A spec gives the current behavior, so a reader sees one definition for each feature (ADR-008).

## 3. Goals

- G-1: A spec gives the current definition of one feature, with requirements R-N and their status.
- G-2: The status rules of ADR-008 apply to specs.
- G-3: `spec_get` gives the commits of each requirement from the `Refs: <slug>#R-N` trailers of HEAD.

## 4. Out of Scope

- Local tasks: the task spec.
- Removal of a requirement or a Removed status: it needs a change of ADR-008.
- Renaming of a slug: the slug does not change (ADR-008).

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | A spec has one file docs/specs/<slug>/SPEC.md, made from the template by spec_create with status Draft. | Done |
| R-2 | The tool spec_get gives the commits of each requirement from the Refs trailers of HEAD. | Done |
| R-3 | The spec tools list, read, create, change status and add requirements of a spec, with the rules of ADR-008. | Done |
| R-4 | The tool spec_update_fields changes one or more of the frontmatter fields title, milestone, priority, area and related of a spec. It does not change slug or status. | Done |
| R-5 | A list that spec_update_fields gets (area, related) replaces the full list. A field that the tool does not get does not change. | Done |
| R-6 | spec_update_fields validates each field with the rules of spec_create, and changes no field on an error. | Done |
| R-7 | spec_update_fields adds one Document History row that names the changed fields, and does not change the status of the spec. | Done |

## 6. Design

### Modules

- `internal/docs/spec`: the store (create, list, get, delete, status, requirements, `spec_update_fields`), the index writer `index.go` and the commit reader `commits.go`. `commits.go` runs `git log` and reads the `Refs: <slug>#R-N` trailers.

### Data flow

`spec_create` writes `SPEC.md` from the template with status `Draft`, and then writes the README index. The requirement tools change section 5 of `SPEC.md`. `spec_get` adds the commits of each requirement from the trailers of HEAD.

### Formats

- Frontmatter: `slug`, `title`, `status`, `milestone`, `priority`, `area`, `related`.
- Requirements: a table `| ID | Requirement | Status | Commits |`, with IDs R-1, R-2 and so on, and status `Planned` or `Done`.

### Decisions

- ADR-008: specs in git, and the gates of the status changes.
- ADR-002: the document tools.

## 7. Acceptance Criteria

- R-1: `TestCreateAndList` (`internal/docs/spec`).
- R-2: `TestCommits` and `TestCommitsGitFailure` (`internal/docs/spec`).
- R-3: `TestGates` and `TestRequirementsAndSetDoneLocked` (`internal/docs/spec`).
- R-4, R-5: `TestUpdateFieldsChanges` and `TestUpdateFieldsSchema` (`internal/docs/spec`).
- R-6: `TestUpdateFieldsErrors` (`internal/docs/spec`).
- R-7: `TestStoreUpdateFields` (`internal/docs/spec`).

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
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | R-3: text changed. |