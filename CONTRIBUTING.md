# Contributing to claude-dx

Thank you for your help. This file tells you how to set up the project, how to make a change and how to send it.

By contributing, you agree to obey the [Code of Conduct](CODE_OF_CONDUCT.md).

## Report a problem or an idea

- Bug: open an issue with the **Bug report** form.
- Idea: open an issue with the **Feature request** form.
- A git operation that no dx tool covers: run `claude-dx report`, or open an issue with the **Unmapped git operation** form.
- Security problem: do not open an issue. Read [SECURITY.md](SECURITY.md).

## Set up

You need:

- Go. The version is in [go.mod](go.mod).
- git 2.30 or later.
- [Claude Code](https://docs.claude.com/claude-code), to test the launcher in a real session.
- Optional: `gopls` (`go install golang.org/x/tools/gopls@latest`), with `$(go env GOPATH)/bin` in your `PATH`. The project enables the `gopls-lsp` Claude Code plugin in `.claude/settings.json`.

Commands:

```sh
go build ./...                      # Build.
go test ./...                       # Test. Add -race on macOS and Linux.
go vet ./...                        # Vet.
go tool govulncheck ./...           # Check for known vulnerabilities.
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run   # Lint.
```

CI runs the same checks on macOS, Linux and Windows. All checks must pass.

## Plan a change

The project records its design in [docs/](docs/):

- [Tasks](docs/tasks/INDEX.md): one spec for each unit of work.
- [Decisions](docs/decisions/): one ADR for each architecture decision.
- [Glossary](docs/GLOSSARY.md): the project terms.

For a small fix, open a pull request directly. For a new feature or a design change, open an issue first, so we can agree on the task before you write code.

## Write code

Write simple, performant code that is easy to delete.

- Make the smallest complete change. Do not refactor unrelated code.
- Use the most direct solution. Do not add abstractions or options that the change does not need.
- Do not add unnecessary allocations, I/O, network calls or processes.
- Keep functions short. The linter limits the cyclomatic complexity to 10.
- Use early returns. Handle errors and special cases first. Avoid deep nesting.
- Keep each feature in few places, so it is easy to delete.
- Add tests for each change. Tests must pass on macOS, Linux and Windows.
- Write docs, code comments and identifiers in [ASD-STE100](https://www.asd-ste100.org/) Simplified Technical English: short sentences, active voice, one instruction in each sentence.

## Commit

Use [Conventional Commits](https://www.conventionalcommits.org/). The commit type sets the next version (release-please, [ADR-012](docs/decisions/ADR-012-semver-with-release-please.md)), so use the correct type.

```text
<type>(<scope>): <summary>
```

| Type | Use for | Version change |
| :--- | :--- | :--- |
| `feat` | A new feature | MINOR |
| `fix` | A bug fix | PATCH |
| `perf` | A performance change | PATCH |
| `refactor`, `test`, `docs`, `build`, `ci`, `style`, `chore` | Other changes | None |
| `revert` | A revert of an earlier commit | PATCH |

- The summary is imperative and lowercase, with no period. The header has a maximum of 72 characters.
- Add `!` after the type, or a `BREAKING CHANGE:` footer, for an incompatible change.
- Make one commit for each logical change.

## Send a pull request

1. Fork the repository and make a branch from `main`.
2. Make your commits. Rebase on `main` to update your branch. Do not merge `main` into your branch.
3. Open a pull request and complete the checklist.
4. A maintainer reviews it. The project accepts only rebase merges, so the history stays linear.

## License

Your contribution is published under the [MIT License](LICENSE).
