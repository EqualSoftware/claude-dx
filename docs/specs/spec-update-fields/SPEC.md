---
slug: spec-update-fields
title: Change the fields of a spec
status: Approved
milestone: v0.1.0
priority: P1
area: [mcp, docs]
related: [ADR-016, ADR-006]
---

# Change the fields of a spec

## 1. Summary

Add the MCP tool `spec_update_fields`, which changes the frontmatter fields of a spec, like `task_update_fields` does for a task.

## 2. Problem

After `spec_create`, no tool changes the frontmatter of a spec. A person must not edit spec files by hand, so a wrong field stays. Example: spec `pushed-adr-lock` has `related: [ADR-017]`, but ADR-021 supersedes ADR-017.

## 3. Goals

- G-1: Each frontmatter field of a spec, except the slug and the status, can change through a dx tool.

## 4. Out of Scope

- A change of the slug (the ID does not change, ADR-016).
- A change of the status (`spec_set_status`).
- Add and remove parameters for the lists.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The MCP tool `spec_update_fields` changes one or more of the frontmatter fields `title`, `milestone`, `priority`, `area` and `related` of a spec. It does not change `slug` or `status`. | Planned |
| R-2 | A list that the tool gets (`area`, `related`) replaces the full list. A field that the tool does not get does not change. | Planned |
| R-3 | The tool validates each field with the rules of `spec_create`: the title has 40 characters or less and no other spec has it, the milestone is `vMAJOR.MINOR.PATCH`, the priority is P0 to P3, and each `related` ID is an ADR that exists. On an error it changes no field. | Planned |
| R-4 | The tool adds one Document History row that names the changed fields, and updates `docs/specs/README.md`. It does not change the status of the spec, also when the spec is `Approved`. | Planned |

## 6. Design

The frontmatter fields are in ADR-016. The tool follows the pattern of `task_update_fields` and of the other spec tools in the package of the spec tools.

- Parameters: `slug` (required), and the optional `title`, `milestone`, `priority`, `area`, `related`. The tool refuses a call with no field to change.
- The tool reuses the validation functions of `spec_create`. The title check ignores the spec itself.
- The tool holds the docs lock for the read, the write of `SPEC.md` and the write of `README.md`, like the other spec tools.
- Document History row: `Changed fields: <field>, <field>.`
- The tool does not change sections 1 to 9, so the status rules of ADR-016 do not apply.
- The `dx:spec` skill names the tool.

## 7. Acceptance Criteria

- R-1: test that each field changes, and that `slug` and `status` are not parameters.
- R-2: test that a list replaces the full list, and that a field that is not given stays.
- R-3: one test for each validation error, with no change to the file. Test that the spec can keep its own title.
- R-4: test of the Document History row and of `README.md`, and that an `Approved` spec stays `Approved`.
- Tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Added R-1. |
| 2026-10-26 | Equal | Added R-2. |
| 2026-10-26 | Equal | Added R-3. |
| 2026-10-26 | Equal | Added R-4. |
| 2026-10-26 | Equal | Requirements, design and acceptance criteria. |
| 2026-10-26 | Equal | Approved by the user. |