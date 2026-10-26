---
id: TASK-016
title: LSP servers for other languages
status: Draft
priority: P3
type: feat
size: ""
owner: Equal
platforms: [macOS, Linux, Windows]
area: [lsp]
depends_on: [TASK-015]
related: [ADR-011]
---

# TASK-016: LSP servers for other languages

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Placeholder. Run /dx:grill when this phase starts (design interview Q-47). |

## 1. Summary
Add LSP servers for HTML, CSS, JSON, Bash, CMake, Lua, Make and Ninja to the LSP manager of TASK-015 (ADR-011).

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
- ADR-011.

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
| Q-1 | Server for each language and pinned versions. | Maintainer | Open |
| Q-2 | Make and Ninja have no mature LSP server. What support do we give (example: symbols and targets from a dx parser)? | Maintainer | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
