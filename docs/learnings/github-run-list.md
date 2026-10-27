---
topic: github-run-list
description: How to find the GitHub Actions runs of one commit with gh. Use it when you check CI for a commit.
---

- `gh run list` has no `--commit` flag. Use `gh run list --json headSha,status,conclusion,name,databaseId` and filter on `headSha`.
