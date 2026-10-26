---
id: TASK-018
title: Task list with no extra calls
status: Done
priority: P1
milestone: v0.1.0
type: fix
size: S
owner: Equal
platforms: [Not applicable]
area: [plugin]
depends_on: [TASK-006]
related: [TASK-006, ADR-005, ADR-009]
---

# TASK-018: Task list with no extra calls

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial specification from a session of the maintainer in another project |
| 2026-10-25 | Equal | Approved in the chat (option 1: change only the skill). Work started. The maintainer does the real session test. |
| 2026-10-25 | Equal | Done. The session of the maintainer used 1 tool call. CI runs after the push. |

## 1. Summary
For a request to list tasks, `/dx:task` must give the result of `task_list` and stop. The Report section applies only to create, refine and status changes.

## 2. Problem and Motivation
In a real session (Claude Code 2.1.308), the request `/dx:task liste as tasks abertas` caused `task_list`, then `task_get` (Open Questions) and `Bash(ls docs/tasks/)`. The Report section of the skill asks for the task path and the open questions, and `task_list` does not give the path. The extra calls use tokens, and Bash can cause a permission prompt.

## 3. Goals
- G-1: A list request uses only `task_list`.

## 4. Out of Scope
- A file name in the `task_list` result: it changes the FR-1 format of TASK-006 and adds tokens to each list (option 2, rejected in the chat).

## 5. Context
### 5.1 Current State
`internal/plugin/assets/skills/task/SKILL.md`: the Report section says "Give the task path, its status, its priority and the open questions." for all requests. `task_create` gives the path (`Created docs/tasks/...`); `task_list` and `task_get` do not.

### 5.2 Assumptions
- None.

### 5.3 Constraints
- ADR-005: compact tool results. ADR-009: skills tell the model when to use the tools.

## 6. Requirements
### 6.1 Functional
- FR-1: The skill has a "List tasks" part: run `task_list` with the filters of the request, give its lines, and do not run other tools. Run `task_get` only if the user asks for the content of a task.
- FR-2: The Report section applies to create, refine and status changes. It gives the path from the `task_create` result, and it does not get the path with other tools.

### 6.2 Non-Functional
- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture
Only `internal/plugin/assets/skills/task/SKILL.md` changes.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
`/dx:task liste as tasks abertas` shows the open tasks after one `task_list` call.

### 7.5 Alternatives
- File name in `task_list` (option 2): rejected in the chat. See Out of Scope.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `internal/plugin/assets/skills/task/SKILL.md` | FR-1, FR-2. | Session of the maintainer: `/dx:task liste as tasks abertas` |

## 9. Acceptance Criteria
- AC-1: Given a project with tasks, when the user runs `/dx:task` to list the open tasks, then the transcript shows `task_list` and no other tool call.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Tool calls for a list request | 3 (`task_list`, `task_get`, `Bash`) | 1 | Transcript of the session of the maintainer |

- Tests: one manual session by the maintainer.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The model still reads the files | Low | Then add the file name to `task_list` (option 2) |

## 12. Removal and Compatibility
Revert the skill text.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [x] All acceptance criteria pass, with evidence. AC-1: session of the maintainer in another project (Claude Code 2.1.308, binary built from `40e462f`): the request "liste as tarefas" gave one `task_list` call and no other tool call. Metric: 1 call; baseline 3.
- [x] Project checks (build, tests, lint) pass. Local on macOS: `go build ./...`, `go test ./...`. Only a skill file changed. CI runs after the push.
- [x] Docs are updated (README). Not applicable: no user-facing command changed.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no diagram.
- [x] ADRs are recorded for architectural decisions. No new decision.
- [x] `/dx:eureka` is done. The tool does not exist yet. No new lesson: the cause is in section 2.
