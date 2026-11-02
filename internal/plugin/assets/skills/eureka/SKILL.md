---
name: eureka
description: Record the lessons of a session as an ADR, a learning or a glossary term, after the user selects them. Use at the end of a session, or when the user asks to keep what the session taught.
---

# Record the lessons of a session

A lesson is a fact or a decision that the session proved. This skill keeps only the lessons that are verified, reusable and not yet documented. The user selects each lesson before you record it. You do not commit.

## Find the lessons

- Read the conversation and the commits of the session.
- Make a list of candidate lessons.
- Remove a lesson that is not verified. A guess is not a lesson.
- Remove a lesson that later work cannot reuse.
- Remove a lesson that a dx document already has. Run `adr_list`, `learning_list` and `glossary_lookup` to find it.
- Give each lesson that stays one type:
  - ADR: a decision that is expensive to reverse or that affects more than one spec.
  - Learning: a verified fact about tools, platforms, libraries or the domain.
  - Glossary term: a project term that is new, ambiguous or has a different meaning here.

## When no lesson stays

Report "No new lessons". Do not change a file. Stop.

## Ask the user

- Show the list. Give each lesson a number, its type and one line.
- Ask the user to select the lessons to record.
- Record nothing that the user did not select.

## Record

Record each selected lesson with the skill of its type:

- ADR: `dx:adr`.
- Learning: `dx:learning`.
- Glossary term: `dx:glossary`.

## Limits

- Do not commit. Tell the user to run `/dx:commit` for the records.
- Do not change specs, tasks or spikes. Notes of an exploration stay with the spike skill.
- If a lesson needs new work, do not start it. Put it in the report.

## Report

- Each record: its type and its ID or term.
- Each lesson that needs new work, with one line.
