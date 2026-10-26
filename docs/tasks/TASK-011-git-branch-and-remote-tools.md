---
id: TASK-011
title: Git tools for branches and remotes
status: Approved
priority: P1
type: feat
size: M
owner: Equal
platforms: [macOS, Linux, Windows]
area: [git]
depends_on: [TASK-010]
related: [ADR-008]
---

# TASK-011: Git tools for branches and remotes

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |

## 1. Summary
Add `git_branch`, `git_sync`, `git_rebase`, `git_conflicts` and `git_push` with the safety rules of ADR-008: no merge, `--force-with-lease` only, no force push to the main branch unless the user authorizes it with elicitation.

## 2. Problem and Motivation
The project uses rebase, not merge. Force push is dangerous. The model must not decide alone to rewrite remote history.

## 3. Goals
- G-1: `git_sync` updates the current branch with fetch and rebase in one call.
- G-2: `git_push` never runs plain `--force`, and asks the user before a force push to the main branch.

## 4. Out of Scope
- Merge: ADR-008 forbids it.
- Pull requests (`gh`): not part of dx.

## 5. Context
### 5.1 Current State
TASK-010 gives the elicitation helper and the remote check.

### 5.2 Assumptions
- A-1: `origin/HEAD` names the main branch. If it is not set, the main branch is `main`, then `master`. Verify: `git symbolic-ref refs/remotes/origin/HEAD`.

### 5.3 Constraints
- ADR-008.

## 6. Requirements
### 6.1 Functional
- FR-1: `git_branch(action, name?, new_name?, start?)`, actions `list`, `create`, `switch`, `rename`, `delete`. `delete` uses `-d`; a branch that is not merged needs elicitation.
- FR-2: `git_sync()` runs `fetch --prune` and `rebase --autostash <upstream>`. On a conflict it returns the `git_conflicts` result.
- FR-3: `git_rebase(action, onto?)`, actions `start` (needs `onto`), `continue`, `abort`, `skip`. `continue` uses `GIT_EDITOR=true`.
- FR-4: `git_conflicts()` returns each conflicted path with its conflict hunks and line numbers.
- FR-5: `git_push(set_upstream?)`. If the remote branch is not an ancestor of the local branch (history changed), it uses `--force-with-lease --force-if-includes`. If that branch is the main branch, it asks with elicitation first.

### 6.2 Non-Functional
- NFR-1: Network commands have a timeout of 120 seconds and `GIT_TERMINAL_PROMPT=0`, so a credential prompt never blocks the server.

## 7. Design
### 7.1 Architecture
`internal/git/branch.go`, `internal/git/remote.go`. All use the runner (TASK-009) and the elicitation helper (TASK-010).

### 7.2 Public API
```text
git_branch(action, name?, new_name?, start?) -> text
git_sync() -> text
git_rebase(action, onto?) -> text
git_conflicts() -> text
git_push(set_upstream?: bool) -> text
```

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
The elicitation dialog for a force push names the branch, the remote and the number of remote commits that the push replaces.

### 7.5 Alternatives
- `git pull --rebase`: rejected. `git_sync` gives the same result with autostash and a compact conflict result.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/git/branch.go` | FR-1, FR-3, FR-4. | `go test ./internal/git/` |
| 2 | `internal/git/remote.go` | FR-2, FR-5, NFR-1. Tests with a bare remote. | `go test ./internal/git/` |

## 9. Acceptance Criteria
- AC-1: Given a branch behind its upstream with local changes, when `git_sync` runs, then the branch is rebased and the local changes stay.
- AC-2: Given an amended commit on `main` and a user that says no, when `git_push` runs, then nothing is pushed.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Plain `--force` calls in code | None | 0 | `grep -rn '"--force"' internal/git/` |

- Tests: temp repositories with a bare remote.
- Platforms not verified: none.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| Credential helpers on Windows open a window | Medium | Document it; the timeout stops a blocked command |

## 12. Removal and Compatibility
Delete `branch.go` and `remote.go`.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README tool list).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
