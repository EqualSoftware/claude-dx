---
id: TASK-058
title: Keep the file mode in docfile.Write
status: Done
priority: P3
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: [TASK-062]
related: [ADR-007, TASK-049]
---

# TASK-058: Keep the file mode in docfile.Write

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Status: In Review. |
| 2026-10-26 | Equal | Status: Approved. |
| 2026-10-26 | Equal | Status: In Review. The scope changed in section 6.1. |
| 2026-10-26 | Equal | The user approved FR-2 (Q-44) in the conversation. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`docfile.Write` keeps the mode of the file that it replaces.

## 2. Problem and Motivation

Found by the code review of 2026-10-26. `Write` (`internal/docs/docfile/write.go:47`) sets mode 0644 on each file. A repository with docs at 0664 for a group loses the group write permission after each edit of a task, ADR, learning or the glossary. Git does not show this change, because it records only the execute bit.

## 3. Goals

- G-1: An edit does not change the mode of a doc file.

## 4. Out of Scope

- Windows: Go sets only the read-only bit there.

## 5. Context
### 5.1 Current State

- `Write` makes a temporary file with `os.CreateTemp` (mode 0600), sets 0644, then renames it.

### 5.2 Assumptions

Not applicable.

### 5.3 Constraints

- ADR-007: the write rule (temporary file and rename).
- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `Write` sets the mode of the existing file at `path` (`Mode().Perm()`) on the temporary file. If `path` does not exist, it sets 0644.
- FR-2: If `os.Stat(path)` fails with an error that is not "does not exist", `Write` returns the error and makes no temporary file, because a mode from no data could change the mode of the file with no sign.

### 6.2 Non-Functional

- NFR-1: One `os.Stat` for each write.

## 7. Design
### 7.1 Architecture

`Write` calls `os.Stat(path)` and uses its `Mode().Perm()`, or 0644.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Use 0666 and the umask: rejected, `os.CreateTemp` does not use the umask for a later `Chmod`.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/docs/docfile/write.go`, `internal/docs/docfile/write_test.go` | FR-1, with tests. | `go test ./internal/docs/docfile/` |

## 9. Acceptance Criteria

- AC-1 (Unix): Given a file with mode 0664, when `Write` runs, then the file has mode 0664.
- AC-2: Given no file, when `Write` runs, then the file has mode 0644 on Unix.
- AC-3 (Unix): Given a path whose `os.Stat` fails with a permission error, when `Write` runs, then it returns the error and the directory has no temporary file.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Mode changes after an edit | Each edit | 0 | AC-1 test |

- Tests: unit tests for each AC.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| None | — | — |

## 12. Removal and Compatibility

Revert the change.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: AC-1 and AC-2 TestWriteKeepsMode; AC-3 TestWriteStatFails; FR-2 also TestWriteStatLoop (internal/docs/docfile/write_unix_test.go). go test ./... passes on macOS; Linux by CI after push.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet (also GOOS=windows), go test ./... pass; golangci-lint v2.15.0: 0 issues.
- [x] Docs are updated (area docs, `CLAUDE.md`).
  - Evidence: The comment of docfile.Write says that the file keeps its mode. No other doc describes the file mode.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: No diagram: one Stat call in an existing function.
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: the write rule of ADR-007 (temporary file and rename) does not change.