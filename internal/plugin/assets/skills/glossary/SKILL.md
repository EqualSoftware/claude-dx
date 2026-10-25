---
name: glossary
description: Add, change or check terms in docs/GLOSSARY.md with the dx glossary tools. Use when a conversation, task, ADR or brief uses a project term that is new, ambiguous, has synonyms, or has a different meaning in this project.
---

# Maintain the glossary

The glossary is `docs/GLOSSARY.md`. One term has one meaning. One meaning has one term. The dx glossary tools read and write the table. Do not change `docs/GLOSSARY.md` with Edit or Write, and do not read all of it: use `glossary_lookup`.

## When to use

Add or change a term when:
- the user or a doc uses a new project term, a domain term or an abbreviation,
- two words are used for the same thing (for example "scene" and "screen"),
- one word is used for two different things,
- a term in the code changes its name or meaning,
- a task, ADR or brief needs a term that a new agent cannot understand without this conversation.

Do not add general words (for example "file", "function") or terms with the usual technical meaning and no project-specific use.

## Procedure

1. Run `glossary_lookup` for the term and for each synonym. The tool compares the full text with no case difference, and also searches the Do not use column.
2. Get the meaning from evidence: the code (`path:line`), a doc, or the words of the user.
   - If the meaning is not clear, ask the user. Do not guess.
   - Domain and design terms need the confirmation of the user. Do not infer them from sample or test code.
3. If the term exists:
   - Same meaning: do nothing. Use the existing term.
   - Changed meaning: change the row with `glossary_upsert`. Give the search for the old use in `docs/` and `.claude/` to the `dx:scout` agent, and report the places that need an update.
4. If the term is new, run `glossary_upsert`. The tool keeps alphabetical order.
   - `definition`: one or two sentences. Say what it is, not how to use it.
   - `area`: `project`, `process`, `build` or the name of an area of the project.
   - `do_not_use`: the synonyms to avoid, or empty.
   - `source`: the file that defines or uses the term.
5. If `glossary_upsert` refuses a conflict, or two terms have one meaning, ask the user which term to keep. Put the other term in `do_not_use`. Remove a row with `glossary_delete`.
6. If a doc defines the same term, keep the full definition in the glossary. In the doc, link to the glossary.

## Use of the glossary

- In tasks, ADRs, briefs and docs, use the terms with their glossary meaning.
- In a brief, if a term is important for the work batch, write "Terms: see `docs/GLOSSARY.md`" and name the terms.
- Code identifiers keep the code names. The glossary links the term to the identifier.

## Report

List the terms that you added, changed or found in conflict. Or "No glossary changes".
