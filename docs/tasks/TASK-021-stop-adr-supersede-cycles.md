---
id: TASK-021
title: Stop ADR supersede cycles
status: Draft
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
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: `adr_supersede` refuses a new ADR that is superseded, because the replacement must be the current decision.
- FR-2: `adr_supersede` refuses an old ADR that is already superseded, because one ADR has one replacement.
- FR-3: `adr_update` refuses a status change of a superseded ADR, because only `adr_supersede` sets that status. A section change stays allowed.
- FR-4: Each refusal gives the reason and the current status, in ASD-STE100.

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
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given ADR-001 superseded by ADR-003, when `adr_supersede(ADR-003, ADR-001)` runs, then the tool refuses it and no file changes.
- AC-2: Given ADR-001 superseded, when `adr_supersede(ADR-001, ADR-002)` runs, then the tool refuses it.
- AC-3: Given ADR-001 superseded, when `adr_update(ADR-001, status: Accepted)` runs, then the tool refuses it; a section change passes.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| [Response time (ms)] | [value] | [value] | [tool, setup, steps] |

- Tests: [unit tests, manual steps].
- Platforms not verified: [list, or None].

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| [Risk] | Low / Medium / High | [Action] |

## 12. Removal and Compatibility

Delete the checks and their tests. Files are not changed.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | [Question] | [Name] | [Answer, or Open] |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
