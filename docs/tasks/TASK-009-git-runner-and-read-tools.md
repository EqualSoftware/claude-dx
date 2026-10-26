---
id: TASK-009
title: Git runner and read-only git tools
status: Approved
priority: P1
area: [git]
depends_on: []
---

# TASK-009: Git runner and read-only git tools

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | `M` (1–3 days) |
| Blocks | TASK-010, TASK-011, TASK-012 |
| Related | ADR-005, ADR-008, TASK-004 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |

## 1. Summary
Add a runner for the system `git` binary, the read-only tools `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, and the git check in `doctor`.

## 2. Problem and Motivation
Git output for persons is long. The model needs short output, and the same behavior on all platforms (ADR-008).

## 3. Goals
- G-1: The five read tools give compact results.
- G-2: `doctor` reports the git version and `FAIL` below the minimum.

## 4. Out of Scope
- Tools that change the repository: TASK-010, TASK-011.
- The Bash hook: TASK-012.

## 5. Context
### 5.1 Current State
TASK-004 gives the server and the project root.

### 5.2 Assumptions
- A-1: Git porcelain v2 output (`status --porcelain=v2 --branch`) is stable for parsing. Verify: git docs.

### 5.3 Constraints
- ADR-008: system git only.

## 6. Requirements
### 6.1 Functional
- FR-1: The runner runs `git` in the project root with `LC_ALL=C`, `GIT_TERMINAL_PROMPT=0`, no pager and no color. It returns stdout, stderr and the exit code. A git error becomes a short tool error with the first lines of stderr.
- FR-2: `git_status()` returns the branch, upstream, ahead and behind counts, and the staged, unstaged, untracked and conflicted paths, in groups.
- FR-3: `git_diff(target?, paths?, stat?, context?)`: `target` is `unstaged` (default), `staged` or a ref range. `stat: true` returns only the file stats. `context` sets the context lines (default 3).
- FR-4: `git_log(count?, range?, path?)` returns one line for each commit: short hash, date, author, subject. Default count 20.
- FR-5: `git_show(ref, path?, stat?)`.
- FR-6: `git_blame(path, start, end)` returns one line for each source line: short hash, author, date, line text.
- FR-7: `doctor` checks that `git` is in `PATH` and that its version is equal to or more than 2.30 (Q-1).

### 6.2 Non-Functional
- NFR-1: `git_status` on a repository with 10 000 files takes less than 300 ms more than `git status`.

## 7. Design
### 7.1 Architecture
`internal/git/runner.go` (runner), `internal/git/read.go` (tools). The runner is the only place that starts `git`.

### 7.2 Public API
```text
git_status() -> text
git_diff(target?, paths?: [string], stat?: bool, context?: int) -> diff
git_log(count?: int, range?: string, path?: string) -> lines
git_show(ref, path?: string, stat?: bool) -> text
git_blame(path, start: int, end: int) -> lines
```

### 7.3 Data and File Formats
Example `git_status` result:
```text
branch main -> origin/main (ahead 1, behind 0)
staged:     M internal/cli/parse.go
unstaged:   M README.md
untracked:  docs/learnings/x.md
```

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- See ADR-008.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/git/runner.go`, `internal/doctor/` | FR-1, FR-7. | `go test ./internal/git/ ./internal/doctor/` |
| 2 | `internal/git/read.go` | FR-2 to FR-6. Tests with temp repositories on the three platforms. | `go test ./internal/git/` |

## 9. Acceptance Criteria
- AC-1: Given a repository with staged and untracked files, when `git_status` runs, then the result has the groups of 7.3.
- AC-2: Given git below the minimum, when `doctor` runs, then git is `FAIL` with the minimum version.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| `git_status` result size | `git status` output | ≤ 50 % | Compare on this repository |

- Tests: temp repositories in CI on the three platforms.
- Platforms not verified: none.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Paths with spaces or non-ASCII characters | Medium | Use `-z` output where git supports it |

## 12. Removal and Compatibility
Delete `internal/git/read.go`. The runner stays for other git tools.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Minimum git version. Recommended: 2.30 (January 2021; it has `--force-if-includes`). | Maintainer | Accepted: 2.30. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
