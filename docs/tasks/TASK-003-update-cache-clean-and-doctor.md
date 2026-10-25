---
id: TASK-003
title: Update, cache clean and doctor commands
status: In Development
priority: P1
area: [cli]
depends_on: [TASK-002, TASK-004]
---

# TASK-003: Update, cache clean and doctor commands

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | `M` (1–3 days) |
| Blocks | None |
| Related | ADR-001, ADR-004 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |
| 2026-10-25 | Equal | Work started. |
| 2026-10-25 | Claude | Decisions of the maintainer: Q-2 to Q-5. Facts: version comparison uses `golang.org/x/mod/semver` (already in the module graph); download URLs come from `browser_download_url` of the release assets; archive names are the names of `scripts/install.sh`; `update.json` is written to a temporary file and renamed; requests send `User-Agent: claude-dx/<version>`. `CheckInBackground` is `Check`; the MCP server runs it in a goroutine. `claude-dx cache` with no `clean` prints the usage and exits with code 1. |

## 1. Summary
Add `claude-dx update`, `claude-dx cache clean` and `claude-dx doctor`, and the background update check in the MCP server.

## 2. Problem and Motivation
The user must update the binary with no package manager (ADR-004), remove all data of claude-dx (ADR-001) and find setup problems quickly.

## 3. Goals
- G-1: `claude-dx update` installs the latest release after a checksum check, on all platforms.
- G-2: `claude-dx cache clean` removes `<cache>` and reports files that it cannot remove.
- G-3: `claude-dx doctor` reports each check as `OK` or `FAIL` with a fix.
- G-4: The MCP server checks for a new release a maximum of one time in 24 hours.

## 4. Out of Scope
- Signature verification: ADR-004.
- Git check in `doctor`: TASK-009 adds it.
- LSP checks: TASK-015.

## 5. Context
### 5.1 Current State
TASK-002 gives the parser and the update notice that reads `<cache>/update.json`. TASK-004 gives the MCP server.

### 5.2 Assumptions
- A-1: The GitHub API `GET https://api.github.com/repos/EqualSoftware/claude-dx/releases/latest` with no token is sufficient. The limit is 60 requests in 1 hour for each IP. One check in 24 hours is far below it.

### 5.3 Constraints
- ADR-004.

## 6. Requirements
### 6.1 Functional
- FR-1: `update` gets the latest release, compares it with the current version (semver), and stops with `claude-dx is up to date (vX.Y.Z).` if it is not newer.
- FR-2: `update` downloads the archive for `runtime.GOOS` and `runtime.GOARCH` and `checksums.txt`, verifies the SHA-256 and extracts the binary to a temporary file in the same directory as the current binary.
- FR-3: `update` replaces the binary: on macOS and Linux with an atomic rename; on Windows it renames the current binary to `claude-dx.exe.old` and renames the new file to `claude-dx.exe`.
- FR-4: `update` writes `<cache>/update.json` with the new state.
- FR-5: `cache clean` removes `<cache>`. It reports each path that it cannot remove (example: a file in use on Windows) and exits with code 1 in this case.
- FR-6: `doctor` checks: `claude` is in `PATH`; the `claude` version is equal to or more than the minimum (Q-1); `<cache>` is writable; the plugin of the current version is extracted. The exit code is 1 if a check fails (Q-4).
- FR-7: The MCP server starts a goroutine at start. If `<cache>/update.json` is older than 24 hours or does not exist, it gets the latest release and writes the file. Network errors are ignored and the file is not changed.
- FR-8: `dev` builds do not check for updates. `update` in a `dev` build stops with exit code 1 (Q-3).
- FR-9: At start, the launcher reads `<cache>/update.json`. If it has a newer version, the launcher prints one line to stderr before it starts `claude`: `claude-dx vX.Y.Z is available. Run: claude-dx update` (from TASK-002 FR-8).

### 6.2 Non-Functional
- NFR-1: The update check never blocks MCP tool calls.
- NFR-2: HTTP timeout is 10 seconds for the API call, and for the connection and the response headers of a download. The body of a download has no time limit (Q-2).

## 7. Design
### 7.1 Architecture
`internal/update/` has `Latest(ctx)`, `Apply(ctx, release)` and `CheckInBackground(ctx)`. `internal/cache/` has `Dir()` and `Clean()`. `internal/doctor/` has a list of checks; each check returns a status, a message and a fix.

### 7.2 Public API
```text
claude-dx update
claude-dx cache clean
claude-dx doctor
```

### 7.3 Data and File Formats
```json
{"checked_at": "2026-10-25T10:00:00Z", "latest": "v0.2.0"}
```

### 7.4 User-facing Behavior
```text
$ claude-dx doctor
OK    claude 2.1.308 in PATH
OK    cache ~/Library/Caches/claude-dx is writable
FAIL  plugin v0.1.0 is not extracted. Fix: run claude-dx once.
```

### 7.5 Alternatives
- See ADR-004.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/update/` | FR-1 to FR-4, FR-7, FR-8. Tests with an `httptest` server that serves releases and checksums, also a bad checksum. | `go test ./internal/update/` |
| 2 | `internal/cache/`, `internal/doctor/`, `internal/cli/` | FR-5, FR-6. Connect the commands. | `go test ./internal/...` |
| 3 | `internal/mcpserver/` | Start `CheckInBackground` at server start. | `go test ./internal/mcpserver/` |

## 9. Acceptance Criteria
- AC-1: Given a newer release, when `update` runs on Windows, then the new binary runs and the next start deletes `claude-dx.exe.old`.
- AC-2: Given a bad checksum, when `update` runs, then the binary does not change and the exit code is 1.
- AC-3: Given `update.json` from 2 days ago, when the MCP server starts, then `update.json` gets a new `checked_at`.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Network calls of the launcher | None | 0 | Code review: the launcher only reads `update.json` |

- Tests: unit tests with `httptest`. AC-1: a test serves a release with a real `claude-dx` binary from `httptest`, runs the update while the old binary runs, then runs the new binary and the launcher. CI runs it on macOS, Linux and Windows (Q-5).
- Platforms not verified: Linux arm64.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| No write permission to the binary directory | Medium | Stop with a message that names the directory |
| `cache clean` during a session | Low | The session keeps running; the next start extracts the plugin again |

## 12. Removal and Compatibility
Remove `internal/update/` and the `update` command. Users then update with the install scripts.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Minimum Claude Code version. Recommended: the version on which TASK-002 verifies A-1 to A-4 (now 2.1.308). | Maintainer | Accepted: the version on which TASK-002 verifies A-1 to A-4 (now 2.1.308). |
| Q-2 | NFR-2 timeout for a download of some MB. Recommended: 10 s for the API call and until the response headers; no limit for the body. | Maintainer | Accepted. |
| Q-3 | `update` in a `dev` build. Recommended: refuse, because a dev build has no version to compare. | Maintainer | Accepted. |
| Q-4 | Exit code of `doctor` with a failed check. Recommended: 1. | Maintainer | Accepted. |
| Q-5 | AC-1 evidence with no Windows computer and no public test release. Recommended: an automatic test on the CI runners. | Maintainer | Accepted. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
