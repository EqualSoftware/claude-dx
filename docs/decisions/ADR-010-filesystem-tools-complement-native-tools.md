# ADR-010: Filesystem tools replace only shell file operations

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The native tools `Read`, `Edit`, `Write`, `Glob` and `Grep` run inside Claude Code and work on all platforms. Claude Code checkpoints (`/rewind`) restore only edits of native tools. `Edit` detects stale files. The permission prompt shows a diff. File operations through Bash (`mv`, `rm`, `cp`, `mkdir`) change between platforms.

## Decision
- The filesystem tools replace only the shell file operations: `fs_move`, `fs_copy`, `fs_delete`, `fs_mkdir`, `fs_tree` (obeys `.gitignore`).
- They add operations that save tool calls: `fs_read_many` (many files or line ranges in one call), and `fs_move` that updates imports with LSP (after ADR-011).
- `Read`, `Edit`, `Write`, `Glob` and `Grep` stay native.
- In v0.3, the `PreToolUse` hook (ADR-008) also blocks `mv`, `rm`, `cp` and `mkdir` in Bash.
- The full tool list is decided when v0.3 starts.

## Alternatives
- Replace all native file tools: rejected. Loss of checkpoints, stale file detection, diff prompts and image and PDF reading.

## Consequences
- Positive: same file operations on all platforms. No loss of native features.
- Negative: two sets of file tools.
- Follow-up: interview for v0.3 before TASK-014 starts.

## Evidence
- Design interview (2026-10-25): Q-5 and Q-13.
