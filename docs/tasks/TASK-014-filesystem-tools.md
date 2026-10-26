---
id: TASK-014
title: Filesystem tools (v0.3)
status: Draft
priority: P2
type: feat
size: ""
owner: Equal
platforms: [macOS, Linux, Windows]
area: [fs]
depends_on: [TASK-012]
related: [ADR-010, TASK-004]
---

# TASK-014: Filesystem tools (v0.3)

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Placeholder. Run /dx:grill when this phase starts (design interview Q-47). |

## 1. Summary
Add filesystem tools that replace shell file operations (`mv`, `rm`, `cp`, `mkdir`) and add operations that save tool calls. Extend the Bash hook to block the shell file operations (ADR-010).

## 2. Problem and Motivation
Open. See Open Questions.

## 3. Goals
Open. See Open Questions.

## 4. Out of Scope
Open. See Open Questions.

## 5. Context
### 5.1 Current State
Open. See Open Questions.

### 5.2 Assumptions
Open. See Open Questions.

### 5.3 Constraints
- ADR-010.

## 6. Requirements
### 6.1 Functional
Open. See Open Questions.

### 6.2 Non-Functional
Open. See Open Questions.

## 7. Design
### 7.1 Architecture
Open. See Open Questions.

### 7.2 Public API
Open. See Open Questions.

### 7.3 Data and File Formats
Open. See Open Questions.

### 7.4 User-facing Behavior
Open. See Open Questions.

### 7.5 Alternatives
Open. See Open Questions.

## 8. Implementation Plan
Open. See Open Questions.

## 9. Acceptance Criteria
Open. See Open Questions.

## 10. Validation and Metrics
Open. See Open Questions.

## 11. Risks and Mitigations
Open. See Open Questions.

## 12. Removal and Compatibility
Open. See Open Questions.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Full tool list and parameters (start: `fs_move`, `fs_copy`, `fs_delete`, `fs_mkdir`, `fs_tree`, `fs_read_many`). | Maintainer | Open |
| Q-2 | Does `fs_delete` need elicitation, or a move to a trash directory? | Maintainer | Open |
| Q-3 | Which shell commands does the hook block (also `rmdir`, `touch`, `ln`)? | Maintainer | Open |
| Q-4 | `fs_move` with import updates: in this task or after TASK-015? | Maintainer | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
