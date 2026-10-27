---
topic: subagent-tool-names
description: How subagent models use MCP tool names from their prompt. Use it when you write agent or skill rules that name tools.
---

- A `haiku` subagent calls an MCP tool by the name that its prompt gives. If the prompt gives only the short name (example `glossary_list`), the call fails with `No such tool available`, and the model tries again with the full name `mcp__plugin_dx_mcp__glossary_list`. Evidence: a Haiku search agent of the dx plugin (removed by ADR-022) in the v0.1.0 QA session of 2026-10-25, Claude Code 2.1.308, 2 failed calls of 7 (TASK-024).
- The `tools` list in the agent frontmatter does not stop this, because the model takes the name from the rules text.
