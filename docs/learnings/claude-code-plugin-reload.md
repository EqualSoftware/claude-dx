---
topic: claude-code-plugin-reload
description: What /reload-plugins does to the MCP server of a --plugin-dir plugin. Use it to test a new claude-dx binary in a session that runs.
---

- On Claude Code 2.1.308, `/reload-plugins` stops the plugin MCP server and starts it again. Evidence: the PID of `claude-dx mcp` changed from 14605 to 32455 after one `/reload-plugins` in the v0.1.0 QA session of 2026-10-25.
- So after `go install ./cmd/claude-dx`, one `/reload-plugins` runs the new MCP server code. The skills, the agents and `rules.md` do not change, because the launcher extracts them only at start.
