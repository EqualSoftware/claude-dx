# ADR-016: Scope of the filesystem tools

- Status: Accepted
- Date: 2026-11-02
- Owners: Equal

## Context
ADR-002 lists the planned filesystem tools: fs_move, fs_copy, fs_delete, fs_mkdir, fs_tree, fs_read_many, and fs_move that updates imports with LSP. It also says that a PreToolUse hook blocks mv, rm, cp and mkdir in Bash. The refinement of the spec filesystem-tools found three problems. The native Edit tool refuses a file that the native Read tool did not read in the session, so fs_read_many does not save tool calls. ADR-013 says that claude-dx does not manage language servers. A hook that blocks all mv, rm, cp and mkdir also blocks the cleanup of temporary directories outside the project, and the fs_ tools refuse those paths. ADR-002 is pushed and does not change.

## Decision
This ADR replaces the part "Filesystem tools" of ADR-002. The other parts of ADR-002 stay valid.
- The tools are fs_move, fs_copy, fs_delete, fs_mkdir and fs_tree. fs_read_many is not built.
- fs_move does not update imports (ADR-013).
- Each tool refuses a path outside the project root with Project.Resolve.
- fs_move and fs_copy take overwrite (default false). They refuse the project root and .git as source or destination. fs_delete needs recursive for a directory that is not empty, and refuses the project root and .git.
- Read, Edit, Write, Glob and Grep stay native.
- The PreToolUse hook on Bash denies mv, rm, cp and mkdir only when a path argument is inside the project root. A path with a variable or a glob counts as inside. The hook uses the parser of the git block (ADR-003).
- The spec filesystem-tools gives the full behavior.

## Alternatives
- Keep fs_read_many: Rejected. The native Edit tool needs a native Read first, and the model can send many Read calls in one message.
- Block all mv, rm, cp and mkdir in Bash: Rejected. The agent cannot clean temporary directories outside the project, because the fs_ tools refuse those paths.
- Supersede ADR-002 with a full copy: Rejected. It copies all the text about the MCP server to change one part.

## Consequences
- Positive: The tool list is small and does not conflict with the native tools. The hook protects the project files and does not block temporary files.
- Negative: ADR-002 still shows the old list of filesystem tools. A reader must also read this ADR. The hook cannot see file operations inside scripts or make targets.
- Follow-up: Implement the spec filesystem-tools.

## Evidence
internal/hook/hook.go: the parser of the git block (RunsGit). Spec filesystem-tools, questions Q-1 to Q-9. ADR-013: claude-dx does not manage language servers.
