---
slug: cli-doctor
title: Doctor command of claude-dx
status: Approved
milestone: v0.1.0
priority: P1
area: [cli]
related: [ADR-001, ADR-006, ADR-012]
---

# Doctor command of claude-dx

## 1. Summary

claude-dx doctor checks the setup of claude-dx. It prints OK or FAIL for each check, with a fix for each FAIL, and exits with code 1 if one check fails.

## 2. Problem

When a session does not start, the user needs to know why. The cause is often a missing claude, an old claude, a cache that is not writable, or a plugin that is not extracted.

## 3. Goals

- G-1: `claude-dx doctor` reports each check as OK or FAIL with a fix.

## 4. Out of Scope

- The start of claude and the plugin extraction: see the specs cli-launch and plugin.
- The cache lock: see the spec session.
- A repair: doctor reports a fix and does not change the setup.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | doctor checks that claude is in PATH, that its version is at least the minimum, that the cache is writable, and that the plugin is extracted. It prints OK or FAIL with a fix and exits with code 1 on a FAIL. | Done |
| R-2 | The doctor command holds the shared cache lock during the command. | Done |

## 6. Design

### Modules

- `internal/doctor/doctor.go`: the list of checks. Each check gives a status, a message and a fix.
- `cmd/claude-dx/main.go`: `runDoctor` takes the shared cache lock with `session.Use`, then runs the checks. If the lock fails, the cache is not writable, and the checks report it.

### Data flow

1. `doctor` tries to take the shared cache lock (see session).
2. `doctor` runs the checks and prints one line for each.
3. The exit code is 1 if one check is FAIL.

### Decisions

- ADR-001: the launcher and the CLI.
- ADR-006: the shared cache lock.

## 7. Acceptance Criteria

Done requirements (tests that exist):
- R-1: `TestDoctor` (`cmd/claude-dx/main_test.go`) and `TestClaudeVersion` (`internal/doctor/doctor_test.go`).
- R-2: no automated test. Check that `runDoctor` calls `session.Use` before the checks.

All tests pass on macOS. Linux and Windows run in CI.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification. Split from cli. |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | Status: In Review. |
| 2026-10-27 | Equal | Approved by the user. |