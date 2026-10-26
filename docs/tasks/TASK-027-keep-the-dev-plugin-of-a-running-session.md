---
id: TASK-027
title: Keep the dev plugin of a running session
status: Draft
priority: P3
area: [cli, plugin]
depends_on: []
---

# TASK-027: Keep the dev plugin of a running session

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-001, TASK-002 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

For a `dev` build, `plugin.Ensure` removes and extracts `<cache>/plugin/dev` at each start (`internal/plugin/plugin.go:25`, `:46`). A second session removes the plugin files of a session that runs.

## 2. Problem and Motivation

The maintainer uses `dev` builds of claude-dx on this project. When a second session starts, the first session loses `rules.md`, the skills and the agents until the extraction ends, and then reads files of the new build. On Windows, the remove can fail when a file is open. Release builds are not affected.

## 3. Goals

- G-1: The start of a `dev` session does not change the plugin files of another session.

## 4. Out of Scope

- Release builds.
- Removal of old dev directories by `cache clean`: it removes all data.

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
- FR-1: A `dev` build extracts to a directory with a key of the embedded files, example `dev-<hash>`, and uses it again when it exists, because the same files give the same key.

### 6.2 Non-Functional
- NFR-1: The key comes from the embedded files and the binary path at start. No network call.

## 7. Design

### 7.1 Architecture
`plugin.Ensure` (`internal/plugin/plugin.go:22`).

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
New cache directories `<cache>/plugin/dev-<hash>`.

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

- AC-1: Given a `dev` session that runs, when a second `dev` session with other embedded files starts, then the files of the first session do not change.

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

Revert to the `dev` directory. `cache clean` removes the old directories.

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
