---
slug: skill-eureka
title: Eureka skill
status: Draft
milestone: v0.4.0
priority: P2
area: [skills, plugin]
related: [ADR-010]
---

# Eureka skill

## 1. Summary

The `/dx:eureka` skill records the lessons of a session. It keeps only the lessons that are verified, reusable and not yet documented. It records each one as an ADR, a learning or a glossary term. It uses the git tools of git-tools when it commits.

## 2. Problem

The README roadmap plans skills "to record the lessons of a session" for v0.4.0. A session can end with lessons that no document records. The project then loses them. The learning skill (`dx:learning`) and the ADR skill (`dx:adr`) exist, but no skill selects the lessons of a session.

## 3. Goals

Goals for the skill are not set in git-tools. See Open Questions.

## 4. Out of Scope

Out-of-scope items for the skill are not set in git-tools. See Open Questions.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The /dx:eureka skill keeps only verified, reusable and undocumented lessons, records each one as an ADR, a learning or a glossary term, and reports "No new lessons" and changes no file when there is none. | Planned |

## 6. Design

### Planned module

- `internal/plugin/assets/skills/eureka/`: the skill. It uses the git tools of git-tools to commit the records.

### Related skills

- `dx:adr`: records an ADR.
- `dx:learning`: records a learning.
- `dx:glossary`: records a glossary term.

## 7. Acceptance Criteria
- [R-N: the test or check that proves the requirement.]

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Define the goals and the out-of-scope items of the Eureka skill. git-tools gives only requirement R-20 for this skill. | Equal | Open |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |