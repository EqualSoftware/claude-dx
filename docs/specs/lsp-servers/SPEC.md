---
slug: lsp-servers
title: LSP manager and language servers
status: Draft
milestone: v0.4.0
priority: P2
area: [lsp, mcp]
related: [ADR-004]
---

# LSP manager and language servers

## 1. Summary

The LSP manager gives the dx MCP server language servers for code work: definitions, references, symbols and rename. It downloads pinned servers into the cache and checks their checksums. The binary does not embed them (ADR-004). This spec is planned for v0.4.0 and has no code yet.

## 2. Problem

The agent needs the language server results (references, symbols, rename) to work on code with few tool calls and no wrong text match. Each language server is a separate download with its own runtime, so the manager must pin and check each one.

## 3. Goals

- G-1: The LSP tools give the references, rename, symbols and reads of one symbol for each supported language.
- G-2: Each server is a pinned version with a checked checksum. The binary does not change when a server changes.
- G-3: A server stays running for the session, so each call does not start it again.

## 4. Out of Scope

- The full design (server list, start policy, tool list, Make and Ninja support) is decided when v0.4.0 starts (ADR-004). Make and Ninja are in the planned language list (R-6), but the decision is open.
- The cleanup of the LSP servers in the cache. The spec of the interactive clean command (cli-clean) adds its item when this spec is approved.
- The import updates of fs_move: see filesystem-tools.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | The claude-dx binary does not embed LSP servers. The LSP manager downloads pinned versions to <cache>/lsp/ and verifies their checksums. | Planned |
| R-2 | For a server that needs Node.js, the manager downloads a pinned portable Node.js runtime. | Planned |
| R-3 | The dx MCP server starts the LSP servers and keeps them warm for the session. | Planned |
| R-4 | The LSP tools give the references, rename, symbols and the read of one symbol. | Planned |
| R-5 | gopls uses the Go toolchain of the system. | Planned |
| R-6 | The manager supports Go, TypeScript, JavaScript, Python, C, C++, HTML, CSS, JSON, Bash, CMake, Lua, Make and Ninja. | Planned |

## 6. Design

### Planned modules

- `internal/lsp/`: the manager. It downloads the pinned server and its runtime to `<cache>/lsp/`, checks the SHA-256, and starts the server.
- `internal/lsp/tools`: the MCP tools (references, rename, symbols, read one symbol).

### Data flow (planned)

1. The manager starts when the first LSP tool is called in a session. It downloads and checks the server if it is not in the cache.
2. The server stays running for the session. The tools send requests to it.
3. gopls runs with the Go toolchain of the system.

### Formats

- `<cache>/lsp/`: one directory for each server and version, with the checked files.

### Decisions

- ADR-004: LSP servers are downloaded, pinned and checked. They are not embedded in the binary.

## 7. Acceptance Criteria

Done requirements: none.

Planned requirements (tests to write):
- R-1 and R-2: a test with a fake download server that serves a file with a bad checksum (refused, nothing is kept) and with a good checksum (kept in `<cache>/lsp/`).
- R-3: a test that a second LSP call in one session does not start the server again.
- R-4: a test of references, rename, symbols and the read of one symbol on a small project for each language of R-6.
- R-5: a test that gopls is found in the PATH of the system.
- R-6: the tests of R-4 for each language in the list (Go, TypeScript, JavaScript, Python, C, C++, HTML, CSS, JSON, Bash, CMake, Lua, Make and Ninja).

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
| 2026-10-27 | Equal | R-6: text changed. |
| 2026-10-27 | Equal | Changed fields: related. |