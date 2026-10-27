---
name: grill
description: Interview the user about a plan, spec, ADR or idea until both sides have the same understanding. Use to refine a spec, before an ADR, before a plan, or when the user asks to stress-test a decision.
---

# Grill a decision

Interview the user until you and the user have the same understanding. Do not assume an answer.

Ask the questions in the language of the user. Write the answers in documents and tools in ASD-STE100.

## 1. Decision tree

Make a tree of the decisions. Each decision has child decisions that depend on its answer. The caller gives the levels to cover:

- `/dx:spec`: the levels in its "Refine a spec" section.
- `/dx:adr`: problem and constraints, options, selected option, consequences, follow-up.
- A plan: objective, scope, work batches and their order, files, checks, risks.
- Other topics: goals, scope, constraints, design, acceptance, risks.

## 2. Facts and decisions

- **Facts** come from the code, the docs and the tools. Do not ask the user for a fact. Give a search of more than 3 files, or of other tasks, ADRs and docs, to the `dx:scout` agent. Use one brief with a closed list of questions. Read only 1 to 3 known files yourself. Each tool call of the main session costs the full context.
- **Decisions** belong to the user. Ask the user for each decision. Do not make one silently.
- If a fact search is not finished, wait only with the questions that depend on it. Ask the other questions.

## 3. Rounds

The **frontier** is the set of open decisions whose parent decisions have an answer. Repeat:

1. Ask all questions of the frontier in one round. Do not ask a question that depends on an open question. Ask it in a later round.
2. Number each question (`Q-n`). Give each question a bold title and the context it needs. Give options if they help.
3. Give your recommended answer after each question: `→ Recommended: ...`. Phrase the question so that "yes" accepts the recommendation.
4. Use `AskUserQuestion` for 4 questions or fewer. For more questions, write them as a numbered list separated by horizontal rules.
5. Wait for the answers. Write each answer where the caller says.
6. Recompute the frontier. Add the new questions that the answers cause.

For a spec:
- Add each question with `spec_question_add` before you ask it. Use the `Q-n` ID that the tool returns.
- Write each answer with `spec_question_answer`. Change the sections that the answer affects with `spec_update_section`, `spec_req_add` and `spec_req_set`.

A task has no questions: refine its spec.

## 4. Check for hidden decisions

When the frontier is empty, read the document again, section by section (`spec_get` or `adr_get` with `section`). A line is a hidden decision if it does not come from one of these sources:
- an answer of the user,
- a fact with a `path:line`, a command output or a doc reference.

For example: a return type, a name, an order, a rule that you derived from the answers.

Check these items in each requirement:
- each error path and each fallback,
- each user-facing text (messages, help, docs),
- each existing command or tool that the change affects,
- each conflict with an Accepted ADR (`adr_list`, `adr_get`).

1. Ask: "Can a new agent understand and accept this line without this conversation?" If not, the line is not clear.
2. Make each hidden decision and each line that is not clear into a question of a new round, with a recommendation (section 3).
3. Repeat until no hidden decision is left.

## 5. Stop

Stop when the frontier is empty and section 4 finds no hidden decision: each branch has an answer and nothing is assumed. Then ask the user to confirm that the understanding is the same. Do not continue to the next step (review, ADR status, plan execution) before the user confirms.

If you run as a subagent and cannot ask the user, do not guess. Return the current frontier as questions with recommendations, and stop.

## Report

The decisions with their answers, the questions that are still open, and the user confirmation (yes or no).
