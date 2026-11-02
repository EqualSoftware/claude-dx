---
slug: cli-update
title: Update command of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [cli]
related: [ADR-005, ADR-006, ADR-018]
---

# Update command of claude-dx

## 1. Summary

claude-dx update installs the latest release in place of the running binary. It checks the SHA-256 of the archive first. On Windows it works for a running binary, and the next start deletes the old file.

## 2. Problem

The user must update the binary with no package manager (ADR-005). A bad download must never replace a good binary, and a running binary on Windows must not block the update.

## 3. Goals

- G-1: `claude-dx update` installs the latest release after a checksum check, on all platforms.

## 4. Out of Scope

- The background check for a new release and the file update.json: see the spec mcp-server.
- The notice at start: see the spec cli-launch.
- The cache lock: see the spec session.
- The release workflow and the archive names: see the spec release.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | At start, the launcher deletes each regular file <binary>.old and <binary>.old.<n> if it exists, and ignores errors. It does not delete directories or other names. | Done |
| R-2 | update compares the latest release with the current version (semver) and stops with a message when the release is not newer. | Done |
| R-3 | update downloads the archive for the OS and architecture, verifies its SHA-256 from checksums.txt, and replaces the binary. A bad checksum leaves the binary and exits with code 1. | Done |
| R-4 | On Windows, update replaces a running binary and the next start deletes claude-dx.exe.old. | Done |
| R-5 | update stops with exit code 1 in a dev build. | Done |
| R-6 | The update command holds the shared cache lock during the command. | Done |

## 6. Design

### Modules

- `internal/update/`: `Latest` and `Apply`. The update archive names follow `scripts/install.sh`.
- `cmd/claude-dx/main.go`: `runUpdate` takes the shared cache lock with `session.Use`, then gets the latest release and replaces the binary.
- `internal/launcher/launcher.go`: deletes the file `<binary>.old` at start.

### Data flow

1. `update` takes the shared cache lock (see session).
2. `update` gets the latest release, checks the SHA-256 from `checksums.txt`, and replaces the binary.
3. On Windows the running binary is renamed to `claude-dx.exe.old`. The next start deletes it.

### Decisions

- ADR-005: self-update from GitHub releases.
- ADR-006: the shared cache lock.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestLaunchDeletesOldBinary` (`cmd/claude-dx/main_test.go`).
- R-2: `TestNewer` and `TestUpToDateAndDev` (`internal/update/update_test.go`).
- R-3: `TestUpdateReplacesBinary` and `TestBadChecksum` (`internal/update/update_test.go`).
- R-4: `TestUpdateRunningBinary` (`internal/update/update_test.go`), which runs on Windows CI.
- R-5: `TestUpdateRefusesDev` (`cmd/claude-dx/main_test.go`).
- R-6: no automated test. Check that `runUpdate` calls `session.Use` before the update.

All tests pass on macOS. Linux and Windows run in CI.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.
