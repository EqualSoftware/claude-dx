# ADR-011: Download LSP servers on demand and control them from the dx MCP server

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The project must give LSP servers for TypeScript, JavaScript, C, C++, Go, HTML, CSS, JSON, Bash, CMake, Lua, Python, Make and Ninja with no install by the user. Some servers are large (clangd). Many servers need Node.js. gopls needs the Go toolchain. Make and Ninja have no mature LSP server.

## Decision
- The binary does not embed LSP servers.
- The server downloads pinned versions of LSP servers to `<cache>/lsp/`, and verifies their checksums. For servers that need Node.js, it downloads a pinned portable Node.js runtime.
- The dx MCP server starts the LSP servers and keeps them warm for the session. It gives its own LSP tools (examples: references, rename, symbols, read one symbol).
- gopls uses the Go toolchain of the system.
- The full design (server list, start policy, tool list, Make and Ninja support) is decided when v0.4 starts.

## Alternatives
- Embed the servers in the binary: rejected. Hundreds of MB.
- Use the servers that the system has: rejected. Versions change and conflict.
- Native LSP support of Claude Code plugins: rejected as the main path. It has few operations.

## Consequences
- Positive: no install by the user. Same versions on all machines.
- Negative: first use needs a download. Large cache.
- Follow-up: interview for v0.4 before TASK-015 starts.

## Evidence
- Design interview (2026-10-25): Q-7 and Q-8.
