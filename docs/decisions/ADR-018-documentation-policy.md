# ADR-018: Documentation policy

- Status: Accepted
- Date: 2026-11-02
- Owners: Equal
- Supersedes: ADR-012
- Supersedes: ADR-009
- Supersedes: ADR-008

## Context
The number of ADRs increases quickly. The rule "record each architecture decision as an ADR" makes most small decisions into ADRs. An ADR does not change after a push (ADR-009). Thus an addition to an ADR needs a new ADR. For example, ADR-017 exists only because ADR-003 cannot get a new tool. Also, each change of Requirements, Design or Acceptance Criteria sets an Approved spec to In Review (ADR-008). Each new requirement needs a new approval. The Document History of a spec repeats the data of git log and the Refs trailers. These rules cause friction in dynamic projects, and exploration has no place to keep its findings. This ADR replaces ADR-008, ADR-009 and ADR-012 with one policy for all project docs.

## Decision
Where each type of knowledge goes, by how long it stays correct:
- Spec: the current behavior of one subject. It changes when the behavior changes. Git keeps the old versions.
- ADR: only a decision that is expensive to reverse, or that affects more than one spec. Examples: language, data format, rebase only, security rules.
- Design line in a spec: each other decision, as one line "Decision: X, because Y". It can change.
- Learning: a verified fact about tools, platforms or the environment. It can change or be deleted.
- Glossary: the terms of the project.
- Spike: the local findings of an exploration, before a spec exists.
- Model memory: only personal preferences and continuity. Project knowledge stays in files of the repository. dx has no memory tool.

ADR:
- An ADR is Proposed, Accepted or Superseded by ADR-NNN.
- adr_update reads the ADR in the upstream branch (git show @{upstream}:<path>). If the status there is Accepted or Superseded by ADR-NNN, adr_update refuses the change. If the file is not there, or its status there is Proposed, adr_update can change the ADR.
- With no upstream, outside a git repository or with no git command, adr_update can change the ADR. The check uses only the local ref of the upstream. It does not run git fetch.
- A pushed ADR has no amendments. An addition that does not reverse the decision goes to the Design of a spec. A changed decision is a new ADR and adr_supersede.
- adr_supersede changes only the status line of the old ADR to Superseded by ADR-NNN. It adds the line "- Supersedes: ADR-NNN" after the "- Owners" line of the new ADR, also when the new ADR is pushed.
- The new ADR must be Accepted. An ADR that is already superseded cannot get a second replacement. An ADR cannot replace itself.
- adr_update cannot change the status of a superseded ADR, but it can change its sections.
- There is no delete tool for ADRs.

Spec:
- A spec is the current definition of one feature or scope. It is in git: docs/specs/<slug>/SPEC.md, and other files of the spec (diagrams, designs) in the same folder. The ID is the slug, with no number. The slug general is reserved.
- Each spec has one subject: one command or one mechanism. A mechanism spec does not name a command. Only command specs give the user-facing behavior, and they say which mechanisms they use. Each claude-dx command has its own spec cli-<command>. The behavior that all commands share is in cli-common. The MCP server is one command spec, mcp-server. The docs-* specs give its tool areas.
- Frontmatter: slug, title, status, milestone, priority, area, related (ADRs).
- Sections: Summary, Problem, Goals, Out of Scope, Requirements, Design, Acceptance Criteria, Open Questions. A spec has no Document History. Git log of the file and the Refs trailers are the history.
- Requirements is a table ID | Requirement | Status. A requirement ID is R-N and does not change. Its status is Planned or Done. spec_req_add gives the highest current ID + 1. spec_req_delete deletes only a requirement that is Planned, has no commits and has no task, so an ID that comes back has no commit.
- Status: Draft, In Review, Approved. The allowed changes are Draft to In Review, In Review to Approved or Draft, and Approved to In Review. In Review needs the leaf sections filled. Approved needs no Open question.
- A change of Goals or Out of Scope of an Approved spec sets it to In Review. A change of Requirements, Design or Acceptance Criteria keeps the status. The user decides each such change in the conversation before the change.
- Only the spec tools change Requirements and Open Questions.
- A spec does not depend on another spec.
- A spec that is removed from the product, or split, is deleted. Git keeps the old version. The commit message gives the reason and, for a split, the old slug.
- Design links to the ADRs and does not copy them.
- A spec does not contain task IDs. A commit gives the requirements that it implements in a trailer Refs: <slug>#R-N. spec_get finds the commits of each requirement with git log.
- spec_delete refuses while tasks/<slug>/ has a task.

Task:
- A task is a local work item of one developer, in tasks/<spec>/<task-slug>.md of the project data directory (ADR-007), or tasks/general/<task-slug>.md for work that has no spec. Git does not track it. The ID is <spec>/<task-slug>.
- A task points to one spec. depends_on gives tasks of the same spec only.
- Frontmatter: id, title, status, priority, covers (requirement IDs), depends_on. Sections: Brief, Plan, Notes, Definition of Done.
- Status: Todo, In Progress, Blocked, Done, with the current allowed changes. Done is final. In Progress needs an Approved spec and Done dependencies. Blocked needs the blocker in the note. Done needs all Plan steps and Definition of Done items checked.
- When a task goes to Done, the tool sets the requirements in covers to Done in the spec. It writes no history row.
- The docs lock of the project data directory protects the tasks and the spikes (ADR-006).

Spike:
- A spike is a local record of one exploration, in spikes/<slug>.md of the project data directory. Git does not track it.
- Sections: Question, Notes, Result. Status: Open or Closed. Closed needs a Result.
- Tools: spike_create, spike_list, spike_get, spike_note (adds one finding to Notes), spike_close, spike_delete.
- A spike has no tasks and needs no spec. Code of an exploration does not need an Approved spec.
- When the code is stable, /dx:spec makes the spec from the spike and the code. There is no conversion tool. The spike then goes to Closed.

Resume:
- claude-dx resume prints the tasks In Progress, the next tasks, the Blocked tasks, the unpushed commits with their Refs trailers and the changed files. A SessionStart hook runs it on startup, resume, clear and compact.

## Alternatives
- Keep all architecture decisions as ADRs: Rejected. The number of ADRs increases with each small interface change, and git already keeps the history.
- An Amendments section in an Accepted ADR: Rejected. With the new ADR criterion, additions go to the spec Design. Amendments add a second change path.
- Editable ADRs, with git as the history: Rejected. An expensive-to-reverse decision must stay stable after a push.
- A parameter that keeps an Approved spec Approved: Rejected. The section rule is simpler: only a scope change needs a new approval.
- Exploration without a record, or a Spike status of the spec: Rejected. Findings are lost, or the spec model gets a new status. A separate local spike keeps specs and tasks unchanged.
- A memory tool of dx for project knowledge: Rejected. Model memory is not reviewed, not shared with the team and becomes old with no sign.

## Consequences
- Positive: Fewer ADRs. Small decisions change in one place. An Approved spec does not need a new approval for each new requirement. Specs have less text. Exploration keeps its findings and does not change specs or tasks.
- Negative: The Design of a spec holds more decisions, and an agent must read it. A change of Requirements of an Approved spec has no new review. Spikes are local and the team does not see them. A one-time migration changes many files.
- Follow-up: Supersede ADR-008, ADR-009 and ADR-012 with this ADR. A one-time migration: move ADR-014 to the Design of spec session, ADR-015 and ADR-017 to the Design of spec git-tools, ADR-016 to the Design of spec filesystem-tools; delete ADR-004, ADR-008, ADR-009, ADR-012, ADR-014, ADR-015, ADR-016 and ADR-017; change the references in code, specs and related fields; remove Document History from all specs. New spec docs-spike. New requirements in docs-spec, docs-task, docs-adr and plugin. Change the ADR rule in CLAUDE.md and the skills.

## Evidence
ADR-017 Decision: "This ADR adds to ADR-003 and does not replace it." ADR-008 Decision: a change of Goals, Out of Scope, Requirements, Design or Acceptance Criteria of an Approved spec sets it to In Review. ADR-009 Decision: adr_update refuses an ADR whose upstream status is Accepted or Superseded. docs-spec R-8: spec_req_add uses the history to not use an ID again, and spec_req_delete deletes only a Planned requirement with no commits and no task. Spec git-tools left Approved two times for R-23 to R-25.
