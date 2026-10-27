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

## Go language server

Use the LSP tool for Go navigation and refactoring (definitions, references, symbols, rename) before Grep. The project enables the `gopls-lsp` plugin in `.claude/settings.json`. It runs `gopls` from `PATH`: install it with `go install golang.org/x/tools/gopls@latest` and put `$(go env GOPATH)/bin` in `PATH`.

## Docs and process

- Tasks: `docs/tasks/`. Decisions: `docs/decisions/` (ADRs). Learnings: `docs/learnings/`. Terms: `docs/GLOSSARY.md`.
- The `README.md` of each document folder is a generated index for persons. Do not read it. Use `task_list`, `adr_list` and `learning_list`.
- Work from an approved task. Record each architecture decision as an ADR.
- Write docs, code comments, identifiers and commit messages in ASD-STE100 Simplified Technical English. Talk to the user in the language of the user.
- Commits: Conventional Commits. The type sets the next version (ADR-012). Rebase only, no merge.

## dx-agents profile

- Language: ASD-STE100
- Acceptance check: `go build ./... && go vet ./... && GOOS=windows go vet ./... && GOOS=linux go vet ./... && go test -race ./... && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run && go tool govulncheck ./...`
  - Checks: Build, Vet (macOS, Windows, Linux), Test (race), Lint, Vulnerabilities
  - The tests on Linux and Windows run only in CI after a push.
- Independent review areas: locks and concurrency (`internal/filelock`, `internal/session`, the docs lock in `internal/docs/docfile`), process start and signals (`internal/launcher`), file formats and data moves (docs frontmatter, `project.json`, `internal/project`), the release workflow (`.github/workflows`)
- Commit scopes: task, spec, adr, learning, glossary, docs, plugin, mcp, cli, launcher, session, project, release
- Main branch: main
