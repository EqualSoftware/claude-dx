# ADR-017: ADRs do not change after a push

- Status: Proposed
- Date: 2026-10-26
- Owners: Maintainer

## Context
An ADR records a decision at one time. If a pushed ADR changes, readers and other ADRs that link to it see a different decision than the one that was made. Now `adr_update` can change all sections of an ADR at any time.

## Decision
- `adr_update` refuses an ADR whose file exists in the upstream branch (`git cat-file -e @{upstream}:<path>`). A changed decision is a new ADR and `adr_supersede`. With no upstream, `adr_update` can change the ADR.
- `adr_supersede` is the only write to a pushed ADR, and it changes only the status line to `Superseded by ADR-NNN`. The new ADR gets a `Supersedes: ADR-NNN` line.
- One exception: the migration of ADR-016 changes TASK-NNN references in ADRs before the v0.1.0 release. After the v0.1.0 release, the rule has no exception.

## Alternatives
- Full immutability: only the new ADR gives Supersedes: A person who reads the old ADR on GitHub does not see that it is replaced.
- A rule in the skill text only: The model can break a rule in text. A tool check cannot be passed.

## Consequences
- Positive: A pushed decision stays as it was made. The history of a decision is a chain of ADRs.
- Negative: A typo in a pushed ADR stays. adr_update needs a git command.
- Follow-up: Add the upstream check to adr_update and the Supersedes line to adr_supersede in a task of v0.1.0.

## Evidence
- ADR-006 lists `adr_update` with no limit on pushed ADRs.
- Commit 79bfc69 changed ADR-004, which `origin/main` (bb0f52e) has.
