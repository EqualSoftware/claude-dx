---
id: TASK-027
title: Keep the dev plugin of a running session
status: Done
priority: P3
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [cli, plugin]
depends_on: []
related: [ADR-001, TASK-002]
---

# TASK-027: Keep the dev plugin of a running session

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-25 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-25 | Equal | Approved by the maintainer. |
| 2026-10-25 | Equal | Status: In Development. |
| 2026-10-25 | Equal | Status: Done. |

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

`plugin.Ensure` (`internal/plugin/plugin.go:25`) extracts again at each start of a `dev` build, and `install` (`:46`) removes the `dev` directory first. `doctor` checks `<cache>/plugin/<version>` (`internal/doctor/doctor.go:78`).

### 5.2 Assumptions

- None.

### 5.3 Constraints

- ADR-001 (plugin in the cache). Release builds do not change.
- Tests pass on macOS, Linux and Windows. On Windows, a file that is open cannot be removed.

## 6. Requirements

### 6.1 Functional

- FR-1: A `dev` build extracts to a directory with a key of the embedded files and the binary path, example `dev-<hash>`, and uses it again when it exists, because the same files give the same key.
- FR-2: `claude-dx doctor` checks the directory of FR-1 for a `dev` build, because it checks `plugin/<version>` now (`internal/doctor/doctor.go:78`).

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

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/plugin/plugin.go`, `internal/plugin/plugin_test.go`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go` | FR-1, FR-2. A test that extracts two different builds and checks that the first directory does not change. | `go test ./internal/plugin/ ./internal/doctor/` |

## 9. Acceptance Criteria

- AC-1: Given a `dev` session that runs, when a second `dev` session with other embedded files starts, then the files of the first session do not change.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Plugin files of a running dev session that a new start removes | All | 0 | AC-1 test |

- Tests: unit tests in `internal/plugin/plugin_test.go`.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Old `dev-<hash>` directories use disk space | Low | `claude-dx cache clean` removes them. |

## 12. Removal and Compatibility

Revert to the `dev` directory. `cache clean` removes the old directories.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Which milestone for TASK-027? It changes only dev builds, not releases. Recommended: v0.2.0, because v0.1.0 is the first release for users. | Equal | v0.1.0, because the maintainer tests with dev builds before the first release. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence. AC-1: `TestEnsureDevKeepsRunningSession` (`internal/plugin/plugin_test.go`): a second dev build goes to a new `dev-<key>` directory and the marker file of the first directory stays. Manual: with the binary of `e1366e3`, `claude-dx -- --version` made `plugin/dev-6845b11f93707d5c` and the `plugin/dev` directory of this running session did not change; `claude-dx doctor` gives `OK plugin dev-6845b11f93707d5c is extracted`.
- [x] Project checks (build, tests, lint) pass. Local on macOS at `e1366e3`: `go build`, `go vet`, `go test ./...` pass; golangci-lint v2.15.0: 0 issues. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (CONTRIBUTING.md, if it tells how to test a dev build). Not applicable: `CONTRIBUTING.md` and the ADRs do not tell how a dev build extracts the plugin.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no new call flow. Section 7.3 gives the new directory name.
- [x] ADRs are recorded for architectural decisions. Not applicable: no architecture decision. ADR-001 stays correct.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new reusable lesson.