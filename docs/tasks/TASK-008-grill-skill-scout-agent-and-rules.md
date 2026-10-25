---
id: TASK-008
title: Grill skill, scout agent and shared rules
status: In Development
priority: P1
area: [plugin]
depends_on: [TASK-002, TASK-006]
---

# TASK-008: Grill skill, scout agent and shared rules

| Field | Value |
| :--- | :--- |
| Type | `feat` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | Not applicable |
| Size | `S` (< 1 day) |
| Blocks | None |
| Related | ADR-001, ADR-009 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |
| 2026-10-25 | Equal | Work started. |
| 2026-10-25 | Claude | FR-3 changed (Q-2): the rules file is `internal/plugin/assets/rules.md`, and the launcher gives it with `--append-system-prompt-file`, because a second `--append-system-prompt` of the user replaces the first. FR-2: `effort: low` (Q-3). |

## 1. Summary
Add the `/dx:grill` skill, the `dx:scout` agent (Haiku) and the shared rules text that the launcher gives with `--append-system-prompt`.

## 2. Problem and Motivation
The other skills use `/dx:grill` and `dx:scout`. The language and tool rules must apply in each session with no file read (ADR-009).

## 3. Goals
- G-1: `/dx:grill` interviews the user in rounds with recommended answers and writes answers through `task_question_answer` when the topic is a task.
- G-2: `dx:scout` runs on Haiku with read-only tools.
- G-3: The rules text is in the system prompt of each `claude-dx` session.

## 4. Out of Scope
- Git rules: TASK-012 adds them to the rules text.

## 5. Context
### 5.1 Current State
Reference: [equal `grill` skill](https://github.com/is-equal/claude-equal-workflow/blob/main/plugins/equal/skills/grill/SKILL.md). TASK-002 gives the placeholder for the rules.

### 5.2 Assumptions
- A-1: A plugin agent file can set `model: haiku` and a `tools` list. Verify: test session.

### 5.3 Constraints
- ADR-009.

## 6. Requirements
### 6.1 Functional
- FR-1: `/dx:grill` is the equal grill skill with these changes: `equal:scout` becomes `dx:scout`; `/equal:task` and `/equal:adr` become `/dx:task` and `/dx:adr`; for a task, answers go through `task_question_answer`.
- FR-2: `agents/scout.md`: name `scout`, `model: haiku`, `effort: low` (Q-3), tools `Read`, `Grep`, `Glob` and the read-only dx tools. Its instructions: answer the closed list of questions in the brief with `path:line` facts only; do not guess; return "Not found" for a question with no fact.
- FR-3: `internal/plugin/assets/rules.md` holds the rules text. The plugin extraction writes it to the plugin directory. The launcher gives it with `--append-system-prompt-file <plugin dir>/rules.md` (Q-2). First text:

```text
# dx rules
- Write dx documents (ADRs, tasks, learnings, glossary) and dx tool arguments in ASD-STE100 Simplified Technical English.
- Talk to the user in the language of the user.
- Use the dx tools to read and change ADRs, tasks, learnings and the glossary. Do not change these files with Edit or Write.
- Do not read docs/tasks/INDEX.md. Use task_list.
- Use the terms of docs/GLOSSARY.md. Use glossary_lookup for a term that is not clear.
- Give a fact search of more than 3 files to the dx:scout agent, with a closed list of questions.
```

### 6.2 Non-Functional
- NFR-1: The rules text has a maximum of 15 lines.

## 7. Design
### 7.1 Architecture
Files: `internal/plugin/assets/skills/grill/SKILL.md`, `internal/plugin/assets/agents/scout.md`, `internal/plugin/assets/rules.md`.

### 7.2 Public API
`/dx:grill`, agent `dx:scout`.

### 7.3 Data and File Formats
Not applicable.

### 7.4 User-facing Behavior
`/dx:grill` asks questions in rounds, numbered `Q-n`, each with `→ Recommended: ...`. The questions use the language of the user (Q-1).

### 7.5 Alternatives
- See ADR-009.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | the three files of 7.1, `internal/launcher/` | FR-1 to FR-3. | Session: `/dx:grill` on a task; scout answers a brief on Haiku |

## 9. Acceptance Criteria
- AC-1: Given a session started with `claude-dx`, when the model is asked for its dx rules, then it gives the rules of FR-3.
- AC-2: Given a fact search brief, when `dx:scout` runs, then the transcript shows the Haiku model.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Rules size | None | ≤ 15 lines | Line count |

- Tests: manual sessions.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The model ignores the rule "do not use Edit on dx files" | Medium | Add a hook check later if this happens |

## 12. Removal and Compatibility
Delete the three files and the `--append-system-prompt-file` flag in the launcher.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Language of the conversation with the user. The project goal says that documents and the communication of MCP and skills use ASD-STE100 English. Recommended: documents, tool arguments, tool results and skill files use ASD-STE100; the chat with the user (example: grill questions) uses the language of the user. Alternative: all chat in ASD-STE100 English. | Maintainer | A: ASD-STE100 for documents, commit messages, tool arguments, tool results and skill files. The chat with the user uses the language of the user. |
| Q-2 | How does the launcher give the rules? Claude Code 2.1.308 uses only the last `--append-system-prompt`, and only the last `--append-system-prompt-file`, but it uses one of each together. Recommended: a file in the plugin directory with `--append-system-prompt-file`, so that the `--append-system-prompt` of the user also applies. | Maintainer | Accepted. |
| Q-3 | Does `dx:scout` use `effort: low`? Recommended: yes, because fact searches are mechanical. | Maintainer | Accepted. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (README).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
