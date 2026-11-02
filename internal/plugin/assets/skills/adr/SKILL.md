---
name: adr
description: Record an architecture decision of this project as an ADR in docs/decisions with the dx ADR tools. Use when a plan or change selects an approach that is expensive to reverse or that affects more than one spec.
---

# Record an architecture decision

## When to use

Use this skill only when a decision:
- is expensive to reverse, or
- affects more than one spec.

Write each other decision as one line in the Design of the spec: "Decision: X, because Y." Do not make an ADR for it.

Do not record small implementation details or decisions that the code already makes obvious.

Record only decisions of the project. Do not record decisions about the agent setup (`.claude/` and the agent rules in `CLAUDE.md`).

## Procedure

1. Run `/dx:grill` on the decision. Cover: the problem and constraints, the options, the selected option, the consequences and the follow-up. Skip this step only if the conversation already has all answers and the user confirmed them.
2. Run `adr_list`. If an ADR already records the same decision, read it with `adr_get` and change it with `adr_update`. Do not make a duplicate.
3. Run `adr_create` for a new decision. The tool sets the number, the date and the file name.
   - The title has 40 characters or less, and no other ADR has it. The tool refuses other titles.
   - Set `status` to `Accepted` only when the user approved the decision. Else omit it (`Proposed`).
   - In `evidence`, write only verified facts: `path:line`, commits, benchmarks or issues.
   - Keep each field short. Write in ASD-STE100.
4. If the new ADR replaces an older ADR, set the new ADR to `Accepted`, then run `adr_supersede` with the old and the new ID. The tool refuses a new ADR that is not `Accepted` and an old ADR that is already superseded. The status of a superseded ADR cannot change.
5. When the user approves a `Proposed` ADR, run `adr_update` with `status: Accepted`.

Do not change files in `docs/decisions` with Edit or Write. Use the ADR tools. The tools also write the index `docs/decisions/README.md`. Do not read it: it is for persons. Use `adr_list`.

## Report

Give the ADR path, its status and one sentence that states the decision.
