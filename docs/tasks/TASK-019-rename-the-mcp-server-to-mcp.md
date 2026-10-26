---
id: TASK-019
title: Rename the dx MCP server to mcp
status: Done
priority: P1
type: refactor
size: S
owner: Equal
platforms: [Not applicable]
area: [plugin, mcp]
depends_on: [TASK-004]
related: [ADR-002, ADR-005, TASK-002, TASK-008]
---

# TASK-019: Rename the dx MCP server to mcp

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial specification from the v0.1.0 QA of the maintainer |
| 2026-10-25 | Equal | Approved in the chat for v0.1.0. Work started. The maintainer does the real session test. |
| 2026-10-25 | Equal | Done. A-1 is verified. CI runs after the push. |

## 1. Summary
Rename the server key in the `.mcp.json` of the dx plugin from `dx` to `mcp`. Claude Code then shows the server as `plugin:dx:mcp`, and the tool IDs become `mcp__plugin_dx_mcp__<tool>`.

## 2. Problem and Motivation
Claude Code shows the server as `plugin:<plugin>:<server>`. Now this is `plugin:dx:dx`, which looks like an error. `plugin:dx:mcp` tells that the call goes to the MCP server of the dx plugin, and it matches the command `claude-dx mcp`. The change is cheap before the first release; after it, the change breaks the permission rules of users.

## 3. Goals
- G-1: The session shows `plugin:dx:mcp`, and the dx tools run with no permission prompt.

## 4. Out of Scope
- The server name in the MCP `initialize` result (`dx`): Claude Code does not show it.
- Closed tasks that give the old name: they record what was verified at that time.

## 5. Context
### 5.1 Current State
`internal/plugin/assets/.mcp.json:3` (server key `dx`), `internal/launcher/launcher.go:52` (rule `mcp__plugin_dx_dx`), `internal/plugin/assets/agents/scout.md:6` (7 tool IDs), ADR-005 Follow-up.

### 5.2 Assumptions
- A-1: The rule name follows the server key, as TASK-002 A-2 verified: `mcp__plugin_<plugin>_<server>`. Verified: session of the maintainer (AC-1).

### 5.3 Constraints
- ADR-005.

## 6. Requirements
### 6.1 Functional
- FR-1: The server key in `.mcp.json` is `mcp`.
- FR-2: The launcher allows `mcp__plugin_dx_mcp`.
- FR-3: The `dx:scout` tools are `mcp__plugin_dx_mcp__<tool>`.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture
No change in the code structure. Only names change.

### 7.2 Public API
Tool IDs: `mcp__plugin_dx_mcp__<tool>`. Breaking change for permission rules that users wrote with `mcp__plugin_dx_dx`. No release has these names yet.

### 7.3 Data and File Formats
`.mcp.json`: `{"mcpServers": {"mcp": {...}}}`.

### 7.4 User-facing Behavior
Tool calls show `plugin:dx:mcp - <tool> (MCP)`.

### 7.5 Alternatives
- Keep `dx`: rejected by the maintainer. The name looks like an error.
- `tools` or `docs`: rejected. `docs` is wrong after the git tools of v0.2; `mcp` matches `claude-dx mcp`.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `.mcp.json`, `launcher.go`, `scout.md`, ADR-005 | FR-1 to FR-3. | `go test ./...`; session of the maintainer |

## 9. Acceptance Criteria
- AC-1: Given a session started with `claude-dx`, when the model calls a dx tool, then the UI shows `plugin:dx:mcp` and Claude Code asks no permission.
- AC-2: Given a fact search brief, when `dx:scout` runs, then it can call the read-only dx tools.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Permission prompts for dx tools | 0 | 0 | Session of the maintainer |

- Tests: `TestArgs`; one manual session by the maintainer.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A missed old name blocks a tool | Medium | `git grep plugin_dx_dx` finds only closed tasks |

## 12. Removal and Compatibility
Revert the names. No data migrates.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [x] All acceptance criteria pass, with evidence. Session of the maintainer in another project (Claude Code 2.1.308, binary built from `0d1b1e9`). AC-1: `task_list` showed as `plugin:dx:mcp` with no permission prompt. AC-2: `dx:scout` answered a glossary question with `path:line` facts (3 tool uses).
- [x] Project checks (build, tests, lint) pass. Local on macOS: `go test ./...`, golangci-lint. CI runs after the push.
- [x] Docs are updated (README). Not applicable: the README does not give tool IDs.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no diagram.
- [x] ADRs are recorded for architectural decisions. ADR-005 Follow-up gives the new rule names.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new lesson.
