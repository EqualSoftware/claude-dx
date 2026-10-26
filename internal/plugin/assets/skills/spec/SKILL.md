---
name: spec
description: Create, refine or change a spec in docs/specs with the dx spec tools. Use for a new feature, or when the current behavior of a feature changes.
---

# Make or change a spec

A spec is the current definition of one feature or scope (ADR-016). It is in `docs/specs/<slug>/SPEC.md`, with the other files of the spec (diagrams, designs) in the same folder. The ID is the slug. The dx spec tools set the dates, the Document History rows and `README.md`. Do not change spec files with Edit or Write. Do not read `docs/specs/README.md`: it is for persons. Use `spec_list`.

Read only what you need: `spec_get` with `section` returns one section. The full spec and section 5 also give the commits of each requirement.

## When to use

- A new feature: make a spec.
- A change of the behavior of a feature: change its spec. Do not make a second spec for the same feature.
- A spec does not depend on another spec, and it does not contain task IDs.

## Create a spec

1. Collect the inputs: the request of the user, documents, related code, ADRs (`adr_list`) and specs (`spec_list`). Give a search of more than 3 files to the `dx:scout` agent, with a closed list of questions.
2. Run `spec_create` with the slug (kebab-case, 40 characters or less, not `general`), the title, the milestone (`vMAJOR.MINOR.PATCH`), the priority, the areas, the related ADRs and, for a split, `split_from`. The spec starts as `Draft`.
3. Refine it (next section).

## Refine a spec

Run `/dx:grill` on the spec. Use `spec_question_add` and `spec_question_answer` for the questions. Cover all levels:

1. Goals and the problem.
2. Scope and Out of Scope.
3. Requirements: one behavior for each `R-N`.
4. Design: link to the ADRs, do not copy them. A new architecture approach needs an ADR (`/dx:adr`).
5. Acceptance criteria: a test or a check for each requirement.

Change the sections with `spec_update_section`. Write "Not applicable" for a section that does not apply. Use the terms of `docs/GLOSSARY.md`.

## Protected sections

Only their own tools change these sections, so the flow checks stay true:

- Section 5 (Requirements): `spec_req_add` and `spec_req_set`.
- Section 8 (Open Questions): `spec_question_add` and `spec_question_answer`.
- Section 9 (Document History): the tools add the rows.

## Status

`Draft` -> `In Review` -> `Approved`. `In Review` can go back to `Draft`, and `Approved` to `In Review`.

- `In Review` needs no `Open` question, no section with the template text and one requirement or more.
- `Approved` only when the user approves. It needs no `Open` question.
- A change of section 3, 4, 5, 6 or 7 of an `Approved` spec, a new requirement, a requirement text change and a new question set it to `In Review`. The user approves it again.

## Requirements

- `spec_req_add` adds `R-N` with the status `Planned`. An ID does not change, and the tools do not use it again.
- `spec_req_set` changes the text, the status (`Planned` or `Done`), or both. A status change does not change the scope.
- A commit gives the requirements that it implements in a trailer: `Refs: <slug>#R-N`. One commit can have many `Refs:` lines, or a comma list: `Refs: docs-lock#R-1, docs-lock#R-2`.

## Delete or split a spec

- A spec that is removed from the product, or split, is deleted with `spec_delete`. Git keeps the old version, so the folder must have no changes that are not committed.
- The other files of the folder need `move_to` (the slug of a spec that gets them) or `discard_files: true`. For a split, make the new spec first, then delete the old spec with `move_to`.
- A new spec from a split gives the old slug in `split_from` of `spec_create`. The tool writes it in the Document History.
- Give the reason of the delete in the commit message.
