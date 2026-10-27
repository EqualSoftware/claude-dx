---
topic: claude-dx-dev-session
description: How a claude-dx session loads a new build and where the dx tools write. Use it when you test a new claude-dx binary or work in a worktree.
---

- After `go install ./cmd/claude-dx`, exit and start a new `claude-dx` session. `/reload-plugins` does not restart the MCP server, and the plugin runs from a copy that the launcher makes at start (verified 2026-10-26).
- The dx MCP tools write to the main checkout, not to a worktree. Make doc tool changes in the main checkout, and code changes in the worktree.
