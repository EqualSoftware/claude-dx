# ADR-004: LSP servers on demand

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
The project must give LSP servers for TypeScript, JavaScript, C, C++, Go, HTML, CSS, JSON, Bash, CMake, Lua, Python, Make and Ninja with no install by the user. Some servers are large (clangd). Many servers need Node.js. gopls needs the Go toolchain. Make and Ninja have no mature LSP server.

## Decision
- The binary does not embed LSP servers.
- The server downloads pinned versions of LSP servers to `<cache>/lsp/`, and verifies their checksums. For servers that need Node.js, it downloads a pinned portable Node.js runtime.
- The dx MCP server starts the LSP servers and keeps them warm for the session. It gives its own LSP tools (examples: references, rename, symbols, read one symbol).
- gopls uses the Go toolchain of the system.
- The spec lsp-servers decides the server list, the start policy, the tool list, and the Make and Ninja support.

## Alternatives
- Embed the servers in the binary: Rejected. Hundreds of MB.
- Use the servers that the system has: Rejected. Versions change and conflict.
- Native LSP support of Claude Code plugins: Rejected as the main path. It has few operations.

## Consequences
- Positive: no install by the user. Same versions on all machines.
- Negative: first use needs a download. Large cache.
- Follow-up: write the design in the spec lsp-servers before the work starts.

## Evidence
- Not built yet. The spec is docs/specs/lsp-servers/SPEC.md, and its status is Draft.
- internal/mcpserver/server.go:56-76: one server process lives for the session, so it can keep LSP servers warm.
- internal/clean/scan.go:81-105: `clean` shows old plugin versions only. It has no LSP server item yet.
