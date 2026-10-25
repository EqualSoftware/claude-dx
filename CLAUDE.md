# claude-dx

Go CLI that starts Claude Code with the dx plugin: skills, an MCP server and hooks. Read `README.md` for the goal and `docs/` for the design.

## Code rules

Write simple, performant code that is easy to delete.

- **Simple.** Use the most direct solution. Do not add abstractions, layers or options that the task does not need.
- **Performant.** Do not add unnecessary allocations, I/O, network calls or processes.
- **Easy to delete.** Keep each feature in few places, with few dependencies. A person must be able to remove it with no change to unrelated code.
- **Short functions.** Keep cyclomatic complexity low (the linter limit is 10).
- **Early returns.** Handle errors and special cases first and return. Avoid deep nesting and long `if`/`switch` chains.
- Make the smallest complete change. Do not refactor unrelated code.
- Add tests for each change. Tests must pass on macOS, Linux and Windows.

## Commands

```sh
go build ./... && go vet ./... && go test ./...
go tool govulncheck ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
```

## Docs and process

- Tasks: `docs/tasks/` (index: `docs/tasks/INDEX.md`). Decisions: `docs/decisions/` (ADRs). Terms: `docs/GLOSSARY.md`.
- Work from an approved task. Record each architecture decision as an ADR.
- Write docs, code comments, identifiers and commit messages in ASD-STE100 Simplified Technical English. Talk to the user in the language of the user.
- Commits: Conventional Commits. The type sets the next version (ADR-012). Rebase only, no merge.
