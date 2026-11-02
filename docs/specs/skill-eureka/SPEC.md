---
slug: skill-eureka
title: Eureka skill
status: Approved
milestone: v0.4.0
priority: P2
area: [skills, plugin]
related: [ADR-010]
---

# Eureka skill

## 1. Summary

The `/dx:eureka` skill records the lessons of a session. It keeps only the lessons that are verified, reusable and not yet documented. The user selects the lessons. The skill records each selected lesson as an ADR, a learning or a glossary term. It does not commit: the user commits the records with /dx:commit.

## 2. Problem

The README roadmap plans skills "to record the lessons of a session" for v0.4.0. A session can end with lessons that no document records. The project then loses them. The learning skill (`dx:learning`) and the ADR skill (`dx:adr`) exist, but no skill selects the lessons of a session.

## 3. Goals

- G-1: At the end of a session, the lessons that are verified, reusable and not yet documented go into the correct dx document.
- G-2: The user selects each lesson before the skill records it. The skill records nothing that the user did not select.

## 4. Out of Scope

- Notes in a spike. The findings of an exploration stay with the spike skill. See `docs-spike` and `plugin`.
- A change to specs or tasks. The report names a lesson that needs new work, and the skill changes no spec and no task.
- Commits. The skill writes the records and does not commit them. The user commits them with /dx:commit. See `skill-commit`.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The /dx:eureka skill keeps only verified, reusable and undocumented lessons of the session, gives each one a type (ADR, learning or glossary term), and reports "No new lessons" and changes no file when there is none. | Done |
| R-2 | The skill shows the list of candidate lessons with their type and asks the user to select them, and records only the selected lessons, each one with the skill of its type (dx:adr, dx:learning or dx:glossary). | Done |
| R-3 | The skill does not commit, does not change specs, tasks or spikes, and its report lists each record and each lesson that needs new work. | Done |
| R-4 | The plugin ships the skill eureka, and the tool names in its text are registered tools of the dx server. | Done |

## 6. Design

### Module

- `internal/plugin/assets/skills/eureka/SKILL.md`: the skill text.

### Flow

1. Find the candidate lessons of the session: the conversation and the commits of the session.
2. Remove each lesson that is not verified, not reusable, or already in a dx document. Use adr_list, learning_list and glossary_lookup to find a document that has it.
3. When no lesson stays, report "No new lessons" and stop.
4. Show the list with the type of each lesson and ask the user to select.
5. Record each selected lesson with the skill of its type.
6. Report the records and the lessons that need new work.

### Related skills

- `dx:adr`: records an ADR.
- `dx:learning`: records a learning.
- `dx:glossary`: records a glossary term.
- `dx:commit`: the user commits the records with it.

Decision: the user selects the lessons before the skill records them, because a wrong record costs more than one question.
Decision: the skill does not commit, because /dx:commit groups and writes the commits.

## 7. Acceptance Criteria

- R-1, R-2, R-3: `TestEurekaText` (`internal/plugin/plugin_test.go`) checks that the skill text has the filter, the "No new lessons" report, the user selection, the three skills of the types, and no commit step.
- R-4: `TestPluginTextsNameRegisteredTools` (`internal/mcpserver/server_test.go`) and the skill list test of `internal/plugin`.
- The tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Define the goals and the out-of-scope items of the Eureka skill. git-tools gives only requirement R-20 for this skill. | Equal | Goals: G-1 and G-2. The user selects the lessons before the skill records them. The skill does not commit; the user uses /dx:commit. Out of scope: notes in a spike, changes to specs and tasks. |

The status cannot be `Approved` while a question is `Open`.
