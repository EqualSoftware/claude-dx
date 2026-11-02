---
slug: docs-spike
title: Local spikes and spike tools
status: Approved
milestone: v0.4.0
priority: P2
area: [docs, mcp]
related: [ADR-002, ADR-006, ADR-007, ADR-018]
---

# Local spikes and spike tools

## 1. Summary

A spike is a local record of one exploration. It has one question, the findings and the result. It is stored in the project data directory, not in git. A spike has no tasks and needs no spec. The spike tools create, read, change, close and delete spikes. See ADR-018.

## 2. Problem

Before a spec exists, an exploration has no place to keep its findings. The findings stay in the conversation or in model memory, and they are lost. A Spike status of the spec, or a task with no spec, would change the spec and task models. See ADR-018.

## 3. Goals

- G-1: An exploration keeps its question, its findings and its result in one local file. All worktrees of one repository share it.
- G-2: A spike does not change specs or tasks.
- G-3: Many processes can change spikes at the same time, with no lost write and no corrupt file.

## 4. Out of Scope

- A conversion of a spike to a spec or to tasks. /dx:spec makes the spec from the spike and the code. See `plugin`.
- The skill and the rule that tell when to use a spike. See `plugin`.
- The spikes in the output of the resume command. See `resume-command`.
- A tool that changes the Question or the title of a spike, and a tool that opens a Closed spike again.
- The shared rules for document files, such as title length, frontmatter format and atomic writes. See `docs-common`.
- The docs lock rules. See `docs-common`. The move of the project data directory. See `project-data`.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Each spike is a file spikes/<slug>.md in the project data directory, git does not track it, and all worktrees of a repository see the same spikes. | Planned |
| R-2 | A spike has the frontmatter slug, title, status and created, and the sections Question, Notes and Result, in this order. | Planned |
| R-3 | spike_create makes an Open spike with the date of today in created, and refuses a spike with no Question, and a slug that is not kebab-case, is longer than 40 characters or is the slug of a spike that exists. | Planned |
| R-4 | spike_note adds one finding to Notes as the list item "YYYY-MM-DD: <text>" with the date of today, and refuses an empty text. | Planned |
| R-5 | spike_close writes the Result and sets the status to Closed, and refuses an empty Result. | Planned |
| R-6 | A Closed spike is final: spike_note and spike_close refuse it, and only spike_get, spike_list and spike_delete accept it. | Planned |
| R-7 | spike_list gives one line for each spike with its slug, status, created date and title, Open spikes first and then by slug, with an optional status filter, and spike_get gives the full spike or one section. | Planned |
| R-8 | spike_delete deletes one or more spikes of any status, and refuses the full call before any change when a slug does not exist. | Planned |
| R-9 | Each spike write holds the exclusive docs lock of the project data directory, and each spike read holds the shared lock. | Planned |
| R-10 | The spike tools that the plugin texts name are registered tools of the dx server. | Planned |

## 6. Design

### 6.1 Modules

- `internal/docs/spike`: the spike store and the spike tools. It does not import `internal/docs/spec` or `internal/docs/task`.
- `internal/project`: the constant `Spikes`, next to `DocsLock` and `Tasks`, and the project data directory.
- `internal/mcpserver`: registers the spike tools. It sets the spike folder of the spike store.

### 6.2 File layout

A spike is `<data dir>/spikes/<slug>.md`. The data directory is `<git-common-dir>/dx/` in a git repository, else `<cache>/projects/<key>/`. See ADR-007 and ADR-018.

### 6.3 Frontmatter and sections

The frontmatter has `slug`, `title`, `status` and `created`. The sections are Question, Notes and Result, in this order. The Question is the text of `spike_create`. The Notes are a list of findings. The Result of an Open spike is empty.

```markdown
---
slug: lock-on-nfs
title: File locks on NFS
status: Open
created: 2026-11-02
---

## Question

Do the docs lock calls work on an NFS mount?

## Notes

- 2026-11-02: flock on NFSv4 works on Linux.

## Result
```

Decision: each note starts with its date, because git does not track spikes and the file is the only timeline.

### 6.4 Status flow

- `Open` can change to `Closed` with `spike_close`. `Closed` needs a Result.
- `Closed` is final. Decision: no reopen tool, because a new exploration is a new spike.

### 6.5 Writes and locks

Each spike write holds the exclusive docs lock of the project data directory. Each read holds the shared lock. The lock rules are in `docs-common`. See ADR-006.

### 6.6 Tool list

`spike_create`, `spike_list`, `spike_get`, `spike_note`, `spike_close` and `spike_delete`. The tools use compact text results. See ADR-002.

## 7. Acceptance Criteria

- R-1: `TestWorktreesShareSpikes` (`internal/mcpserver/server_test.go`).
- R-2, R-3: `TestCreate` (`internal/docs/spike/spike_test.go`).
- R-4: `TestNote` (`internal/docs/spike/spike_test.go`).
- R-5, R-6: `TestClose` (`internal/docs/spike/spike_test.go`).
- R-7: `TestListAndGet` (`internal/docs/spike/spike_test.go`).
- R-8: `TestDelete` (`internal/docs/spike/spike_test.go`).
- R-9: `TestHoldDocsLock` (`internal/docs/spike/lock_test.go`).
- R-10: `TestPluginTextsNameRegisteredTools` and `TestToolNames` (`internal/mcpserver/server_test.go`).
- The tests pass on macOS, Linux and Windows.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.
