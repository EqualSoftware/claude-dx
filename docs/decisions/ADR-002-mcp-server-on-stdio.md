# ADR-002: One stdio MCP server for each session

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The MCP server must live for all of a Claude Code session, because it keeps LSP servers warm (ADR-011). Two models are possible: Claude Code starts the server, or the launcher starts a daemon that Claude Code connects to.

## Decision
- Claude Code starts `claude-dx mcp` as a stdio MCP server from the plugin declaration (ADR-001).
- One server process serves the main session and its subagents.
- The server uses the official Go SDK `github.com/modelcontextprotocol/go-sdk`.
- The server does the background tasks of the session. Example: the update check (ADR-004).

## Alternatives
- A daemon with HTTP transport that the launcher starts: rejected for now. It can share LSP servers between sessions, but it adds a lifecycle, a port and stale process management. We can add it later.
- `github.com/mark3labs/mcp-go`: rejected. The official SDK supports stdio and elicitation and has long-term maintenance.

## Consequences
- Positive: simple lifecycle. Claude Code stops the server when the session stops.
- Negative: two sessions in the same project have two servers and two sets of LSP servers.
- Follow-up: ADR-007 handles number allocation between parallel servers.

## Evidence
- Design interview (2026-10-25): Q-3 and Q-39.
