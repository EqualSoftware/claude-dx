# ADR-009: ADRs do not change after a push

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
An ADR records a decision at one time. If an accepted ADR changes after a push, readers and other ADRs that link to it see a different decision than the one that was made. A lock on every pushed ADR file is too strict: a pushed ADR with the status `Proposed` must be open to review, change and acceptance. The new ADR of a replacement must also get a link to the old ADR, also when it is pushed.

## Decision
- `adr_update` reads the ADR in the upstream branch (`git show @{upstream}:<path>`). If the status there is `Accepted` or `Superseded by ADR-NNN`, `adr_update` refuses the change. If the file is not there, or its status there is `Proposed`, `adr_update` can change the ADR.
- With no upstream, outside a git repository or with no `git` command, `adr_update` can change the ADR. The check uses only the local ref of the upstream. It does not run `git fetch`.
- A changed decision is a new ADR and `adr_supersede`.
- `adr_supersede` changes only the status line of the old ADR to `Superseded by ADR-NNN`. It adds the line `- Supersedes: ADR-NNN` after the `- Owners` line of the new ADR, also when the new ADR is pushed.
- The new ADR must be `Accepted`. An ADR that is already superseded cannot get a second replacement. An ADR cannot replace itself. These rules also stop cycles.
- `adr_update` cannot change the status of a superseded ADR, but it can change its sections.
- There is no delete tool for ADRs.

## Alternatives
- Lock each ADR file that is in the upstream branch: Rejected. A pushed `Proposed` ADR cannot become `Accepted`, and `adr_supersede` cannot write the `Supersedes` line in a pushed new ADR.
- Check the remote main branch (`origin/main`) and not `@{upstream}`: Rejected. It locks an ADR on a feature branch that has no push yet, and it needs the name of the main branch.
- Run `git fetch` before the check: Rejected. It needs the network, it is slow and it can ask for credentials.
- A rule in the skill text only: Rejected. The model can break a rule in text. A tool check cannot be passed.
- Full immutability, where only the new ADR gives `Supersedes`: Rejected. A person who reads the old ADR on GitHub does not see that it is replaced.

## Consequences
- Positive: a pushed `Proposed` ADR can be reviewed, changed and accepted. A decision that was accepted and pushed stays as it was. The history of a decision is a chain of ADRs. The `Supersedes` line is in the header, where a tool or a person finds it.
- Negative: the check uses the local upstream ref, so a push from another clone is not known before a fetch. An ADR that was accepted only on the local branch can still change. A typo in a pushed ADR stays.
- Follow-up: none.

## Evidence
- internal/docs/adr/upstream.go:13-27 (`checkUpstream`).
- internal/docs/adr/store.go:207-214 (`Update` calls the check), :246-275 (`supersede`), :279-291 (a superseded ADR keeps its status), :295-311 (replacement rules), :327-343 (`Supersedes` line).
- internal/docs/adr/tools.go:44-66 (the tool list has no delete tool).
- internal/docs/adr/upstream_test.go and internal/docs/adr/store_test.go cover each rule.
