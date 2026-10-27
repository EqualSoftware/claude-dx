# ADR-013: No LSP management

- Status: Accepted
- Date: 2026-10-27
- Owners: Equal
- Supersedes: ADR-004

## Context
ADR-004 planned a manager. The manager downloads, pins and checks language servers. An MCP server keeps them warm for the session, and it gives its own LSP tools. Most language servers need Node.js, Java or Python. Many servers have no prebuilt binaries or checksums. The MCP server cannot see the changes that the Edit tool makes, so its view of the code can be stale. Claude Code already has an LSP tool. It gives reads and diagnostics after edits when a Claude Code LSP plugin is installed and its binary is on PATH.

## Decision
claude-dx does not download, start or wrap language servers. claude-dx has no LSP tools. When the session has a language server (a Claude Code LSP plugin with its binary on PATH), the agents use it first. The shared rules of dx-agents already say this. ADR-004 is superseded.

## Alternatives
- Build the LSP manager as ADR-004 planned: Rejected. It adds downloads, runtimes, checksums and a warm server process. Most servers need other runtimes, and the MCP server cannot see Edit changes.
- Use only the Claude Code LSP plugins, with no change to claude-dx: Accepted. It needs no code in claude-dx. The agents use the language server of the session when it exists.

## Consequences
- Positive: claude-dx has less code, fewer downloads and no runtime per language. The cache has no lsp directory. The clean command has no LSP item.
- Negative: claude-dx gives no language server when the session has none. An agent without an LSP plugin uses the text tools (Grep, Glob, Read).
- Follow-up: Remove the LSP mentions from the specs, the glossary and the docs. The spec lsp-servers is deleted. The LSP mentions in ADR-002 and ADR-011 do not apply any more. They stay unchanged (ADR-009).

## Evidence
docs/decisions/ADR-004-lsp-servers-on-demand.md:8-13 (the planned manager and its runtime needs). docs/specs/lsp-servers/SPEC.md:15 (planned status, no code). docs/specs/cli-clean/SPEC.md:30 (LSP item in the clean command). The user decided to remove the LSP feature.
