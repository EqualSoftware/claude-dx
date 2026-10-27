# claude-dx

Go CLI that starts Claude Code with the dx plugin: skills, an MCP server and hooks. Read `README.md` for the goal and `docs/` for the design. The `dx` and `dx-agents` plugins give the shared rules.

## Project rules

- The linter limit of cyclomatic complexity is 10.
- Add tests for each change. Tests must pass on macOS, Linux and Windows.
- Tasks: `docs/tasks/`. Work from an approved task. Record each architecture decision as an ADR.
- The commit type sets the next version (ADR-012). Rebase only: integrate a branch with a rebase and a fast-forward, no merge commit.

## Go language server

The project enables the `gopls-lsp` plugin in `.claude/settings.json`. It runs `gopls` from `PATH`: install it with `go install golang.org/x/tools/gopls@latest` and put `$(go env GOPATH)/bin` in `PATH`.

## dx-agents profile

- Language: ASD-STE100
- Acceptance check: `go build ./... && go vet ./... && GOOS=windows go vet ./... && GOOS=linux go vet ./... && go test -race ./... && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run && go tool govulncheck ./...`
  - Checks: Build, Vet (macOS, Windows, Linux), Test (race), Lint, Vulnerabilities
  - The tests on Linux and Windows run only in CI after a push.
- Independent review areas: locks and concurrency (`internal/filelock`, `internal/session`, the docs lock in `internal/docs/docfile`), process start and signals (`internal/launcher`), file formats and data moves (docs frontmatter, `project.json`, `internal/project`), the release workflow (`.github/workflows`)
- Commit scopes: task, spec, adr, learning, glossary, docs, plugin, mcp, cli, launcher, session, project, release
- Main branch: main
