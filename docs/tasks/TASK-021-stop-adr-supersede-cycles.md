---
id: TASK-021
title: Stop ADR supersede cycles
status: Approved
priority: P1
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [ADR-006, TASK-005]
---

# TASK-021: Stop ADR supersede cycles

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |

## 1. Summary

`adr_supersede` and `adr_update` accept changes that make the ADR chain wrong. Add checks to refuse them.

## 2. Problem and Motivation

In the v0.1.0 QA session (scratch repository, binary from `7063947`):
- `adr_supersede(ADR-001, ADR-003)`, then `adr_supersede(ADR-003, ADR-001)` passed. ADR-001 has "Replaces ADR-003" and ADR-003 has "Replaces ADR-001" and "Superseded by ADR-001". This is a cycle.
- The second call used a superseded ADR (ADR-001) as the new ADR.
- `adr_update(ADR-001, status: Accepted)` changed a superseded ADR back to `Accepted`. The "Replaces" line stayed.

## 3. Goals

- G-1: The ADR chain has no cycles, and a superseded ADR stays superseded.

## 4. Out of Scope

- A tool to undo a supersede.
- Repair of ADR files that are already wrong.

## 5. Context
### 5.1 Current State

`Store.Supersede` (`internal/docs/adr/store.go:183`) refuses only the same ID (`:193`). `Store.Update` (`internal/docs/adr/store.go:162`) changes the status of any ADR, also a superseded one.

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-006 (ADR tools), ADR-005 (compact results).
- Tests pass on macOS, Linux and Windows.

## 6. Requirements

### 6.1 Functional
- FR-1: `adr_supersede` refuses a new ADR that is superseded, because the replacement must be the current decision.
- FR-2: `adr_supersede` refuses an old ADR that is already superseded, because one ADR has one replacement.
- FR-3: `adr_supersede` refuses a new ADR that is not `Accepted`, because a `Proposed` ADR is not a decision yet (Q-1).
- FR-4: `adr_update` refuses a status change of a superseded ADR, because only `adr_supersede` sets that status. A section change stays allowed, because a person can fix a typo or add evidence.
- FR-5: Each refusal gives the reason and the current status, in ASD-STE100.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design

### 7.1 Architecture
Checks in `Store.Supersede` (`internal/docs/adr/store.go:183`) and `Store.Update` (`internal/docs/adr/store.go:162`).

### 7.2 Public API
No change to the tool inputs.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
Not applicable.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/adr/store.go`, `internal/docs/adr/store_test.go` | FR-1 to FR-5. One test for each AC. | `go test ./internal/docs/adr/` |

## 9. Acceptance Criteria

- AC-1: Given ADR-001 superseded by ADR-003, when `adr_supersede(ADR-003, ADR-001)` runs, then the tool refuses it and no file changes.
- AC-2: Given ADR-001 superseded, when `adr_supersede(ADR-001, ADR-002)` runs, then the tool refuses it.
- AC-3: Given ADR-001 superseded, when `adr_update(ADR-001, status: Accepted)` runs, then the tool refuses it; a section change passes.
- AC-4: Given ADR-002 `Proposed`, when `adr_supersede(ADR-001, ADR-002)` runs, then the tool refuses it.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Supersede cycles that the tool accepts | 1 in the QA session | 0 | AC-1 test |

- Tests: unit tests in `internal/docs/adr/store_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| An existing ADR chain is already wrong | Low | Out of Scope. The checks apply at the next change. |

## 12. Removal and Compatibility

Delete the checks and their tests. Files are not changed.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Must the new ADR be Accepted to supersede an old ADR? Recommended: yes, because a Proposed ADR is not a decision yet, so it cannot replace one. | Equal | Yes. Only an Accepted ADR can supersede an ADR. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`/dx:adr` skill, ADR-006: the supersede rules).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.