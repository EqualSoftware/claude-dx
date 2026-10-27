---
slug: docs-adr
title: ADR documents and ADR tools
status: Approved
milestone: v0.1.0
priority: P1
area: [docs, mcp]
related: [ADR-002, ADR-006, ADR-007, ADR-009]
---

# ADR documents and ADR tools

## 1. Summary

An ADR is one file in `docs/decisions`. The ADR tools create, list, read, change and supersede ADRs. A pushed ADR that is accepted or superseded does not change, except through `adr_supersede`.

## 2. Problem

An accepted ADR can change after others read it. Then a reader sees a decision that was not made. The old tools also allowed a supersede cycle and a status change of a superseded ADR.

## 3. Goals

- G-1: An ADR that is Accepted or superseded in the upstream branch changes only through `adr_supersede`.
- G-2: The ADR chain has no cycles, and a superseded ADR stays superseded.
- G-3: One bad ADR file does not hide the other ADRs.

## 4. Out of Scope

- Delete of ADRs: ADR-009 has no delete tool.
- Repair of ADR files that are already wrong.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | adr_create makes a complete ADR file from the template in one call, allocates the number, sets the date and returns the path. | Done |
| R-2 | adr_list returns one line for each ADR in number order, and one warning line for each duplicate number. | Done |
| R-3 | With no docs/decisions folder, adr_list returns a message and makes no folder. | Done |
| R-4 | adr_get returns the full ADR, or one section when a section name is given. | Done |
| R-5 | adr_update changes the status or replaces one section of an ADR. | Done |
| R-6 | adr_supersede changes the status of the old ADR and the new ADR in one call, and refuses a new ADR that is superseded or not Accepted, and an old ADR that is already superseded. | Done |
| R-7 | adr_update refuses a status change of a superseded ADR, and allows a section change of it. | Done |
| R-8 | Each refusal of an ADR tool gives the reason and the current status. | Done |
| R-9 | adr_list gives one warning line for each ADR file that it cannot read or parse, and lists the other ADRs. | Done |
| R-10 | adr_update refuses an ADR whose status in the upstream branch is Accepted or Superseded. If the upstream check cannot run, it allows the change. | Done |
| R-11 | adr_supersede changes only the status line of the old ADR, and adds the line Supersedes: ADR-NNN to the new ADR after the Owners line, also when the new ADR is pushed. | Done |

## 6. Design

### Modules

- `internal/docs/adr/store.go`: create, list, get, update, supersede, and the upstream check of `adr_update`.
- `internal/docs/adr/index.go`: the generated `README.md` of `docs/decisions`.

### Data flow

- `adr_update` reads the status of the ADR in `@{upstream}` with `git show`. The check uses only the local ref and does not fetch. If the check cannot run, the change is allowed. If the upstream status is Accepted or Superseded, the change is refused.
- `adr_supersede` writes the new ADR first, then the status line of the old ADR. The new ADR gets the line `- Supersedes: ADR-NNN` after the Owners line.
- Each write holds the docs lock (ADR-006).

### Formats

- Status lines: `- Status: Proposed`, `Accepted` or `Superseded by ADR-NNN`.
- Owners line: `- Owners: <name>`. The Supersedes line follows it.

### Decisions

- ADR-002: MCP tool conventions and document tools.
- ADR-006: the docs lock.
- ADR-007: number allocation.
- ADR-009: accepted ADRs do not change after a push.

## 7. Acceptance Criteria

- R-1: `TestCreate` and `TestCreateNextNumber` (`internal/docs/adr`).
- R-2: `TestToolsFlow` (`internal/docs/adr`).
- R-3: `TestList` (`internal/docs/adr`).
- R-4: `TestGet` (`internal/docs/adr`).
- R-5, R-7: `TestUpdate` (`internal/docs/adr`).
- R-6, R-8: `TestSupersede` and `TestSupersedeRefusesBadChain` (`internal/docs/adr`).
- R-9: `TestListBadFile` (`internal/docs/adr`).
- R-10: `TestUpdateUpstream`, `TestUpdateUpstreamSuperseded`, `TestUpdateNoUpstream` and `TestUpdateOutsideGit` (`internal/docs/adr`).
- R-11: `TestSupersedePushed` and `TestSupersedeNoOwners` (`internal/docs/adr`).

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
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | Added R-10. |
| 2026-10-27 | Equal | Added R-11. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | R-9: status Done. |
| 2026-10-27 | Equal | R-10: status Done. |
| 2026-10-27 | Equal | R-11: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |