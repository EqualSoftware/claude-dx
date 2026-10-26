---
id: TASK-038
title: Replace block lists in the frontmatter
status: Done
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [macOS, Linux, Windows]
area: [docs-tools]
depends_on: []
related: [TASK-033, go-yaml-flow-lists]
---

# TASK-038: Replace block lists in the frontmatter

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |
| 2026-10-26 | Equal | Approved by the maintainer. |
| 2026-10-26 | Equal | Status: In Development. |
| 2026-10-26 | Equal | Status: Done. |

## 1. Summary

`markdown.SetField` replaces the full value of a key, also when the value is a YAML block list or a multi-line scalar.

## 2. Problem and Motivation

`SetField` (`internal/markdown/frontmatter.go:43`) replaces only the line of the key. A probe test changed `area:` followed by `  - a` and `  - b` to `area: [c]` and kept the old item lines. The YAML is not valid (`did not find expected key`), so `task_list` shows the task only as a warning. A person who edits a task by hand can write a block list.

## 3. Goals

- G-1: After `SetField`, the frontmatter is valid YAML and the key has only the new value, for all YAML forms of the old value.

## 4. Out of Scope

- Other frontmatter changes (comments, key order).
- A YAML parser rewrite of the frontmatter: the other lines stay as they are.

## 5. Context
### 5.1 Current State

`SetField` (`internal/markdown/frontmatter.go:43`) uses the pattern `^key:.*$`. `format` (`:54`) writes the new value in flow style (learning `go-yaml-flow-lists`).

### 5.2 Assumptions

- A-1: The continuation lines of a top-level key start with a space or with `- `. Verify: YAML 1.2 block sequences and block scalars.

### 5.3 Constraints

- Tests pass on macOS, Linux and Windows.

## 6. Requirements
### 6.1 Functional

- FR-1: `SetField` replaces the line of the key and all next lines that start with a space or with `-`, up to the next line that does not. Empty lines are part of the match only when a continuation line comes after them (Q-2).
- FR-2: The new value is in flow style, as now.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

A change in `SetField` only (`internal/markdown/frontmatter.go:43`): the match ends after the continuation lines.

### 7.2 Public API

No change.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

Not applicable.

### 7.5 Alternatives

- Refuse a block list with an error: rejected by the maintainer, because the tool can fix the form itself.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/markdown/frontmatter.go`, `internal/markdown/markdown_test.go` | FR-1, FR-2. Tests: indented block list, block list with no indent, block list as the last key, multi-line scalar with an empty line, flow list (no change of behavior). | `go test ./internal/markdown/` |

## 9. Acceptance Criteria

- AC-1: Given `area:` with two indented `- ` lines, when `SetField(area, [c])` runs, then the result is `area: [c]`, the next key stays, and `Decode` passes.
- AC-2: Given a block list with no indent, then AC-1 is also true.
- AC-3: Given a flow list, then the result is the same as before this task.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| `SetField` results that are not valid YAML | Block lists | 0 | AC tests |

- Tests: the cases of work batch 1.
- Platforms not verified: none. CI runs on macOS, Linux and Windows.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The match removes a line of the next key | Medium | A top-level key starts at column 0 and not with `-`. AC-1 checks the next key. |

## 12. Removal and Compatibility

Revert the change in `SetField`.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Does `SetField` fix a block list, or refuse it? Recommended: replace the key and its continuation lines. | Equal | Replace (grill Q-3, Q-16). |
| Q-2 | A block scalar can have an empty line between its indented lines. FR-1 stops at the empty line and keeps the next indented lines, so the YAML is not valid. Does the match also include empty lines? Recommended: include empty lines only when a continuation line comes after them. | Equal | Include empty lines only when a continuation line comes after them. Empty lines before the next key stay. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [x] All acceptance criteria pass, with evidence.
  - Evidence: TestSetFieldReplacesBlockValue passes (go test ./internal/markdown/ -run SetField -v): AC-1 "indented list", AC-2 "list with no indent", AC-3 "flow list"; each case checks the next key and Decode. Also "list as last key" and "scalar with empty line" (Q-2). Commit d2c11b8.
- [x] Project checks (build, tests, lint) pass.
  - Evidence: go build, go vet, go test ./... pass; golangci-lint v2.15.0: 0 issues; govulncheck: No vulnerabilities found. CI runs on macOS, Linux and Windows after the push.
- [x] Docs are updated (no docs change expected).
  - Evidence: No docs change: public API and behavior of the tools stay the same. The SetField doc comment is updated (internal/markdown/frontmatter.go:41).
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
  - Evidence: Not applicable: a change inside one function, no design diagram (section 7.1).
- [x] ADRs are recorded for architectural decisions.
  - Evidence: No architecture decision: a bug fix in SetField with no change of structure, interface or file format.