---
id: TASK-058
title: Keep the file mode in docfile.Write
status: Draft
priority: P3
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [ADR-007, TASK-049]
---

# TASK-058: Keep the file mode in docfile.Write

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-26 | Equal | Initial draft specification |

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

- FR-1: `Write` sets the mode of the existing file at `path` on the temporary file. If `path` does not exist, it sets 0644.

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
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
