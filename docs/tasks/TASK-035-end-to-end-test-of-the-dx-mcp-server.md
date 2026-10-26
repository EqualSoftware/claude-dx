---
id: TASK-035
title: End-to-end test of the dx MCP server
status: Draft
priority: P2
type: test
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [mcp]
depends_on: []
related: [ADR-002, ADR-005, TASK-004]
---

# TASK-035: End-to-end test of the dx MCP server

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

Add a test that starts the dx server from `mcpserver.New`, connects an MCP client, lists the tools and calls a flow of document tools in a temporary git repository.

## 2. Problem and Motivation

`internal/mcpserver` has no tests (`go test ./...` gives `[no test files]`). The tests of each package call the stores, not the registered tools. In the v0.1.0 QA session, a manual JSON-RPC script on stdio found the bugs of TASK-021, TASK-022 and TASK-023, but the script is not in the repository. A tool that is not registered, or a wrong input schema, is not found by the current tests.

## 3. Goals

- G-1: `go test ./...` finds a missing tool registration or a wrong tool input schema.

## 4. Out of Scope

- A test of the stdio process. The SDK in-memory transport covers the protocol.
- Tests of tool rules that the package tests cover.

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
- FR-1: The test makes a temporary git repository and a temporary cache directory, and starts the server with `mcpserver.New`.
- FR-2: The client connects with `mcp.NewInMemoryTransports` (go-sdk v1.8.0, `mcp/transport.go:183`), so the test starts no process.
- FR-3: The test checks the list of tool names against a fixed list, so a new tool needs a change of the test.
- FR-4: The test calls one flow: `task_create`, `task_list`, `adr_create`, `adr_list`, `glossary_upsert`, `glossary_list`, `learning_create`, `learning_list`, and checks each result text.

### 6.2 Non-Functional
- NFR-1: The test runs in less than 2 seconds and passes on macOS, Linux and Windows.

## 7. Design

### 7.1 Architecture
`internal/mcpserver/server_test.go`.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- Keep a JSON-RPC script in `scripts/`: rejected, because `go test` does not run it.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/mcpserver/server_test.go` | FR-1 to FR-4. | `go test ./internal/mcpserver/` on the CI of three systems |

## 9. Acceptance Criteria

- AC-1: When `go test ./internal/mcpserver/` runs, then it passes and calls each tool of FR-4 through the MCP client.
- AC-2: Given a tool registration is removed, when the test runs, then it fails.

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

Delete `internal/mcpserver/server_test.go`.

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
