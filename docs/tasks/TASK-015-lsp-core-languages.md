---
id: TASK-015
title: LSP manager and core language servers
status: Draft
priority: P2
milestone: v0.4.0
type: feat
size: ""
owner: Equal
platforms: [macOS, Linux, Windows]
area: [lsp]
depends_on: []
related: [ADR-011, TASK-004]
---

# TASK-015: LSP manager and core language servers

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Placeholder. Run /dx:grill when this phase starts (design interview Q-47). |

## 1. Summary
Add the LSP manager (download, checksum, start, warm cache) and the LSP tools, with servers for Go, TypeScript, JavaScript, Python, C and C++ (ADR-011).

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
| Q-1 | Server for each language (example: gopls, typescript-language-server, pyright or basedpyright, clangd) and pinned versions. | Maintainer | Open |
| Q-2 | Start policy: start at session start for detected languages, or at first use. | Maintainer | Open |
| Q-3 | Tool list (references, definition, rename, symbols, read one symbol, diagnostics, hover). | Maintainer | Open |
| Q-4 | Download source and checksum source for each server and for the portable Node.js runtime. | Maintainer | Open |
| Q-5 | C and C++ without `compile_commands.json`: what does the tool do? | Maintainer | Open |
| Q-6 | Also declare the native LSP of the plugin for automatic diagnostics? | Maintainer | Open |
| Q-7 | `doctor` LSP checks and `cache clean` effect on running servers. | Maintainer | Open |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
