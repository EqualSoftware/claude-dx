---
id: TASK-028
title: Do not end a file name slug with a stop word  # no brackets; quote it if it has ": " or " #"
status: Draft  # Draft / In Review / Approved / In Development / Blocked / Done / Cancelled
priority: P3  # P0 / P1 / P2 / P3
area: [docs-tools]  # the areas of the project, for example [api], [docs, ci]
depends_on: [TASK-005, TASK-006]  # for example [TASK-004]
---

# TASK-028: Do not end a file name slug with a stop word

| Field | Value |
| :--- | :--- |
| Type | fix |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows |
| Size | S |
| Blocks | None |
| Related | ADR-006, TASK-005, TASK-006 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |

## 1. Summary

`docfile.Slug` cuts a long title at 60 characters, at a word boundary. The cut can leave a stop word at the end of the file name. Remove the stop words at the end of the slug after the cut.

## 2. Problem and Motivation

`internal/docs/docfile/docfile.go:118` cuts the slug at the last `-` before 61 characters. In the v0.1.0 QA session, this made these file names:
- `TASK-021-prevent-adr-supersede-cycles-and-status-changes-of.md`
- `TASK-023-refuse-task-dependencies-that-do-not-exist-point-to-the-same.md`
- `TASK-026-remove-the-dx-eureka-item-from-the-task-template-until-the.md`
A name that ends with `of`, `the` or `to` looks broken and is hard to read in a file list.

## 3. Goals

- G-1: A cut slug does not end with a stop word.

## 4. Out of Scope

- A change of the names of files that exist, because links to them would break.
- Removal of stop words inside the slug.
- A limit on the title length.

## 5. Context
### 5.1 Current State
[How the system works today. Use `path:line` references.]

### 5.2 Assumptions
- A-1: [Fact that the task accepts as true. How to verify it.]

### 5.3 Constraints
- [Technical, platform, performance, memory, dependency, compatibility or time limits.]
- [Accepted ADRs that apply: ADR-NNN.]

## 6. Requirements

### 6.1 Functional
- FR-1: After the cut, `Slug` removes the stop words at the end of the slug, one by one, because one cut can leave two (example: `point-to-the`).
- FR-2: The stop words are a short fixed list of English function words: a, an, and, as, at, by, for, from, in, into, of, on, or, the, to, until, with.
- FR-3: A slug that is not cut does not change, because a short title is complete.
- FR-4: If all words are stop words, `Slug` keeps the cut result.

### 6.2 Non-Functional
- NFR-1: No new allocation for a slug that is not cut.

## 7. Design

### 7.1 Architecture
One change in `docfile.Slug` (`internal/docs/docfile/docfile.go:118`). ADR and task files use it.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
New long file names can be shorter than before.

### 7.4 User-facing Behavior
Not applicable.

### 7.5 Alternatives
- A shorter title rule in the skills: rejected, because the tool is the place for the same result each time (ADR-005).

## 8. Implementation Plan
[Few, large work batches in order.]

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | [paths] | [exact steps, commands, symbols] | [command] |

## 9. Acceptance Criteria

- AC-1: Given the title of TASK-021, when `Slug` runs, then the result is `prevent-adr-supersede-cycles-and-status-changes`.
- AC-2: Given the title of TASK-023, when `Slug` runs, then the result ends with `point`.
- AC-3: Given a title of less than 60 characters that ends with `to`, when `Slug` runs, then the result ends with `to`.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| [Response time (ms)] | [value] | [value] | [tool, setup, steps] |

- Tests: [unit tests, manual steps].
- Platforms not verified: [list, or None].

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| [Risk] | Low / Medium / High | [Action] |

## 12. Removal and Compatibility

Delete the stop word loop and its test. Existing files keep their names.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | [Question] | [Name] | [Answer, or Open] |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (area docs, `CLAUDE.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
