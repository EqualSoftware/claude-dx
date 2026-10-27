---
slug: docs-glossary
title: Glossary file and glossary tools
status: Approved
milestone: v0.1.0
priority: P1
area: [docs, mcp]
related: [ADR-002]
---

# Glossary file and glossary tools

## 1. Summary

The glossary is the file `docs/GLOSSARY.md`. It is one Markdown table in alphabetical order, with the columns Term, Definition, Area, Do not use and Source. The glossary tools look up, list, add, change and delete its terms. The glossary skill uses the tools.

## 2. Problem

The glossary is mechanical work, and the model made errors in it. Rows could lose their data, have empty fields, or let a word in Do not use become a term.

## 3. Goals

- G-1: `glossary_lookup` finds a term from its term or from its Do not use word.
- G-2: `glossary_list` gives all terms, or the terms of one area, in one call.
- G-3: Each row has a term, a definition, an area and a source, and a change keeps the other data of the row.

## 4. Out of Scope

- A change of the spelling of a term: use `glossary_delete`, then `glossary_upsert`.
- A text search in definitions: `glossary_lookup` and Grep cover it.
- The rows of the existing file: the checks apply at the next upsert of a row.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | glossary_lookup finds a term in the Term column or the Do not use column, with no case difference, and returns its row. | Done |
| R-2 | glossary_lookup returns a message that names the approved term when the term is a word in the Do not use column, and No match. when it finds no term. | Done |
| R-3 | glossary_list returns one row for each term in file order, and keeps only the rows of one area when an area is given, with no case difference. | Done |
| R-4 | glossary_upsert adds a term or replaces its row in alphabetical order, and makes docs/GLOSSARY.md with the table header when the file does not exist. | Done |
| R-5 | glossary_upsert refuses a row with an empty term, definition, area or source, and refuses a Do not use word that is empty or has a comma. | Done |
| R-6 | A changed row keeps the spelling of the term, and keeps the old Do not use words when do_not_use is omitted. | Done |
| R-7 | glossary_delete removes the row of one term. | Done |
| R-8 | The glossary skill uses glossary_list for many terms, and does not read docs/GLOSSARY.md for that question. | Done |

## 6. Design

### Modules

- `internal/docs/glossary/glossary.go`: lookup, list, upsert, delete, and the row checks (`check`) and the word list (`words`).
- `internal/docs/glossary/tools.go`: the MCP tool registrations.
- `internal/docs/glossary/template.md`: the header of a new file.

### Data flow

`glossary_upsert` reads the table, checks the row, puts it in alphabetical order and writes the file. A changed row keeps its old spelling and, when `do_not_use` is omitted, its old Do not use words.

### Formats

- Header: `| Term | Definition | Area | Do not use | Source |`.
- An empty Do not use cell is `—`. A `|` in a cell is escaped.
- A Do not use word is one list item. A word has no comma.

### Decisions

- ADR-002: document tools and paths.

## 7. Acceptance Criteria

- R-1, R-2: `TestToolsFlow` (`internal/docs/glossary`).
- R-3: `TestListArea` (`internal/docs/glossary`).
- R-4: `TestUpsertSortsAndEscapes` and `TestRepositoryRoundTrip` (`internal/docs/glossary`).
- R-5: `TestUpsertRequiredFields` and `TestUpsertRefusesBadWords` (`internal/docs/glossary`).
- R-6: `TestUpsertKeepsOmittedFields` and `TestUpsertConflict` (`internal/docs/glossary`).
- R-7: `TestDeleteMissing` (`internal/docs/glossary`).
- R-8: Not automated. The skill text is checked by review.

The tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |