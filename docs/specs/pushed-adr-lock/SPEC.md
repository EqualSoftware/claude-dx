---
slug: pushed-adr-lock
title: Lock pushed ADRs in adr_update
status: Draft
milestone: v0.1.0
priority: P1
area: [mcp, docs]
related: [ADR-017, ADR-006]
---

# Lock pushed ADRs in adr_update

## 1. Summary

`adr_update` refuses an ADR that the upstream branch has, and `adr_supersede` writes a `Supersedes` line (ADR-017).

## 2. Problem

A pushed ADR can change now, so readers can see a decision that was not made (ADR-017).

## 3. Goals

- G-1: A pushed ADR changes only through `adr_supersede`.

## 4. Out of Scope

- The one-time migration exception (TASK-067 (`tasks-to-specs`)).

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | `adr_update` refuses an ADR whose file exists in `@{upstream}`. With no upstream or outside git, it allows the change. | Planned |
| R-2 | `adr_supersede` changes only the status line of the old ADR and adds `Supersedes: ADR-NNN` to the new ADR. | Planned |

## 6. Design
[How the feature works: modules, data flow, formats, user-facing behavior. Link to the ADRs; do not copy them. A new architecture approach needs an ADR.]

## 7. Acceptance Criteria

- R-1: test repository with a bare remote: a pushed ADR is refused, a local ADR is changed.
- R-2: test of both lines.
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