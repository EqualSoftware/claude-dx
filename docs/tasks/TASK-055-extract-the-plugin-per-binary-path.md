---
id: TASK-055
title: Extract the plugin per binary path
status: Draft
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [plugin, cli]
depends_on: []
related: [ADR-001, ADR-014]
---

# TASK-055: Extract the plugin per binary path

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

## 1. Summary

A release build extracts the plugin to a directory with a key of the binary path, so a moved or second binary does not use a plugin that starts an old binary path.

## 2. Problem and Motivation

Found by the code review of 2026-10-26. `DirName` (`internal/plugin/plugin.go:50`) gives only the version for a release build. The extracted `.mcp.json` and hooks have the absolute binary path (`{{BINARY}}`). `Ensure` uses an existing directory again. If the user moves the binary, or has two binaries of one version (example: `go install` and Homebrew), Claude Code starts the MCP server and the hooks from the first path. The dx tools then fail or run another binary.

## 3. Goals

- G-1: Each binary path of one version has its own plugin directory.

## 4. Out of Scope

- The `dev` build: it has a key of the files and the binary path already.

## 5. Context
### 5.1 Current State

- ADR-001 sets `<cache>/plugin/<version>/`.
- `removeUnused` (`internal/session/session.go`) removes the plugin directories that no session uses, so old directories go away.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-001: plugin extraction. It needs a change.
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: A release `DirName` is `<version>-<key>`, with `<key>` the first 8 bytes of the SHA-256 of the binary path, in hex. The key uses only the path because the version fixes the embedded files.
- FR-2: ADR-001 gives the new directory name.

### 6.2 Non-Functional

- NFR-1: One hash of a short string at start.

## 7. Design
### 7.1 Architecture

`DirName` hashes the binary path for all builds and adds the hash of the files only for `dev`.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

The plugin directory name changes from `<version>` to `<version>-<key>`.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Compare the binary path in the extracted `.mcp.json` with the current path and extract again: rejected, two binaries that run at the same time need two directories.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/plugin/plugin.go`, `internal/plugin/plugin_test.go` | FR-1, with tests. | `go test ./internal/plugin/` |
| 2 | ADR-001 (with `adr_update`) | FR-2. | `adr_get ADR-001` |

## 9. Acceptance Criteria

- AC-1: Given one release version and two binary paths, when `Ensure` runs for each path, then it gives two directories, and each `.mcp.json` has its own path.
- AC-2: Given one release version and one path, when `Ensure` runs two times, then it gives one directory.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Plugins that start another binary path | Possible | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Old `<version>` directories stay in the cache | Low | `removeUnused` removes them at the next start. |

## 12. Removal and Compatibility

Revert the change and ADR-001.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
