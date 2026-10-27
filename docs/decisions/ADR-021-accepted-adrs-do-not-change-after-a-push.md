# ADR-021: Accepted ADRs do not change after a push

- Status: Accepted
- Date: 2026-10-26
- Owners: Maintainer

## Context

Replaces ADR-017.

ADR-017 locks each ADR whose file exists in the upstream branch. This also locks a pushed ADR with the status Proposed, so the team cannot accept it or change it after a review. ADR-017 also allows no write to a pushed new ADR, so `adr_supersede` cannot add the `Supersedes` line to it. Today `adr_supersede` writes `Replaces ADR-NNN.` in the Context section of the new ADR (internal/docs/adr/store.go:309-316). The lock must apply only to decisions that were made and pushed.

## Decision
- `adr_update` reads the ADR in the upstream branch (`git show @{upstream}:<path>`). If the status there is `Accepted` or `Superseded by ADR-NNN`, `adr_update` refuses the change. If the file is not there, or its status there is `Proposed`, `adr_update` can change the ADR.
- With no upstream, outside a git repository or with no `git` command, `adr_update` can change the ADR. The check uses only the local ref of the upstream. It does not run `git fetch`.
- A changed decision is a new ADR and `adr_supersede`.
- `adr_supersede` changes only the status line of the old ADR to `Superseded by ADR-NNN`. It adds the line `- Supersedes: ADR-NNN` after the `- Owners` line of the new ADR, also when the new ADR is pushed. It does not write `Replaces ADR-NNN.` in the Context section.
- One exception: the migration of ADR-016 changes TASK-NNN references in ADRs before the v0.1.0 release. After the v0.1.0 release, the rule has no exception.

## Alternatives
- Keep ADR-017: lock each ADR file in the upstream branch: Rejected. A pushed Proposed ADR cannot become Accepted, and adr_supersede cannot write the Supersedes line in a pushed new ADR.
- Check the remote main branch (origin/main) and not @{upstream}: Rejected. It locks an ADR on a feature branch that has no push yet, and it needs the name of the main branch.
- Run git fetch before the check: Rejected. It needs the network, it is slow and it can ask for credentials.

## Consequences
- Positive: A pushed Proposed ADR can be reviewed, changed and accepted. A decision that was accepted and pushed stays as it was. The Supersedes line is in the header, where a tool or a person finds it.
- Negative: The check uses the local upstream ref, so a push from another clone is not known before a fetch. An ADR that was accepted only on the local branch can still change.
- Follow-up: Spec pushed-adr-lock implements the check in adr_update and the Supersedes line in adr_supersede. Older ADRs keep their `Replaces ADR-NNN.` text.

## Evidence
docs/decisions/ADR-017-adrs-do-not-change-after-a-push.md:11-13; internal/docs/adr/store.go:195-251 (update and supersede); internal/docs/adr/store.go:309-316 (addReplaces); internal/docs/spec/commits.go:24-50 (existing git calls)
