---
slug: pushed-adr-lock
title: Lock pushed ADRs in adr_update
status: Approved
milestone: v0.1.0
priority: P1
area: [mcp, docs]
related: [ADR-017, ADR-006]
---

# Lock pushed ADRs in adr_update

## 1. Summary

`adr_update` refuses an ADR that is `Accepted` or superseded in the upstream branch, and `adr_supersede` writes a `Supersedes` line (ADR-021).

## 2. Problem

An accepted and pushed ADR can change now, so readers can see a decision that was not made (ADR-021).

## 3. Goals

- G-1: An ADR that is `Accepted` or superseded in the upstream branch changes only through `adr_supersede`.
- G-2: A pushed `Proposed` ADR can change and become `Accepted`.

## 4. Out of Scope

- The one-time migration exception of ADR-021 (spec `tasks-to-specs`).
- `git fetch` before the check (ADR-021).
- A change of the `Replaces ADR-NNN.` text in older ADRs.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | `adr_update` refuses an ADR whose status in `@{upstream}` is `Accepted` or `Superseded by ADR-NNN`. If the file is not in `@{upstream}`, its status there is `Proposed`, there is no upstream, the root is outside git or `git` is missing, it allows the change. | Planned |
| R-2 | `adr_supersede` changes only the status line of the old ADR, and adds the line `- Supersedes: ADR-NNN` after `- Owners` in the new ADR, also when the new ADR is pushed. It does not write `Replaces ADR-NNN.` in the Context section. | Planned |

## 6. Design

The rules are in ADR-021. ADR-006 gives the ADR tools.

### Upstream check (R-1)

1. `adr_update` runs `git -C <root> show @{upstream}:<path>`, where `<path>` is the path of the ADR file relative to the repository root, with `/` separators.
2. If the command fails (no upstream, no file in the upstream, not a git repository, no `git` command), the change is allowed.
3. Else the tool reads the `- Status:` line of the output with the existing status pattern (`internal/docs/adr/store.go:28`). If the status is `Accepted` or starts with `Superseded by`, the tool refuses. Else the change is allowed.
4. The check runs before the docs lock is taken, and it changes no file. It uses only the local ref of the upstream.

Refusal text: `ADR-NNN is <status> in the upstream branch. A pushed decision does not change. Record a new ADR and use adr_supersede.`

### Supersede (R-2)

- The old ADR: only the `- Status:` line changes, to `Superseded by ADR-NNN`. There is no upstream check, because this is the only write that ADR-021 allows.
- The new ADR: the tool adds `- Supersedes: ADR-NNN` after the `- Owners:` line. The function `addReplaces` (`internal/docs/adr/store.go:309-316`) is replaced. The tool does not write in the Context section.

### Code

- The git call goes in the package `internal/docs/adr`, in one small function, so it is easy to delete. It uses `exec.Command` like `internal/docs/spec/commits.go:24-50`.
- Tests make a bare remote and a clone with `git` in `t.TempDir()`, like the helper `git` in `internal/docs/spec/spec_test.go:38-45`.

## 7. Acceptance Criteria

- R-1: test repository with a bare remote and an upstream: an ADR that is `Accepted` in the upstream is refused; an ADR that is `Proposed` in the upstream is changed and accepted; a local ADR that is not pushed is changed; a repository with no upstream and a directory outside git allow the change.
- R-2: test that the old ADR changes only on its status line, that the new ADR has `- Supersedes: ADR-NNN` after `- Owners`, that the Context section of the new ADR does not change, and that a pushed new ADR also gets the line.
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
| 2026-10-26 | Equal | R-1: text changed. |
| 2026-10-26 | Equal | R-2: text changed. |
| 2026-10-26 | Equal | Design and acceptance criteria for ADR-021. |
| 2026-10-26 | Equal | Approved by the user. |