---
id: TASK-001
title: Project skeleton, CI and release pipeline
status: Done
priority: P0
area: [build, ci]
depends_on: []
---

# TASK-001: Project skeleton, CI and release pipeline

| Field | Value |
| :--- | :--- |
| Type | `build` |
| Owner | Maintainer |
| Reviewers | Maintainer |
| Platforms | macOS, Linux, Windows (amd64, arm64) |
| Size | `M` (1–3 days) |
| Blocks | TASK-002, TASK-004 |
| Related | ADR-003, ADR-004, ADR-012 |

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial draft specification from the design interview |
| 2026-10-25 | Equal | Open questions answered. Approved. |
| 2026-10-25 | Equal | Work started. |
| 2026-10-25 | Equal | Done. All acceptance criteria pass. |

## 1. Summary
Make the Go project layout, the CI workflow and the release pipeline. After this task, each push runs build, lint and tests on macOS, Linux and Windows, and each `v*` tag publishes the binaries to GitHub Releases.

## 2. Problem and Motivation
The repository has only `go.mod` and docs. All later tasks need a layout, a CI that runs on the three platforms and a way to publish binaries (ADR-004).

## 3. Goals
- G-1: `go build ./...`, `go vet ./...`, `go test ./...` and `golangci-lint run` pass in CI on macOS, Linux and Windows.
- G-2: A `v*` tag publishes 6 archives and `checksums.txt` to GitHub Releases.
- G-3: `install.sh` and `install.ps1` install the latest release and verify the SHA-256.

## 4. Out of Scope
- Launcher behavior: TASK-002.
- Binary signatures: ADR-004 rejects them for v1.
- Other distribution channels: ADR-004.

## 5. Context
### 5.1 Current State
`go.mod` declares `module github.com/is-equal/claude-dx` and `go 1.27.3`. There is no Go code.

### 5.2 Assumptions
- A-1: The repository `github.com/is-equal/claude-dx` exists before the first release. Verify: `gh repo view is-equal/claude-dx`.
- A-2: golangci-lint supports Go 1.27. Verify: CI run.

### 5.3 Constraints
- No dependencies other than the Go standard library in the launcher (ADR-003).
- ADR-004: GitHub Releases only.

## 6. Requirements
### 6.1 Functional
- FR-1: The layout is the layout in section 7.1.
- FR-2: `.github/workflows/ci.yml` runs on push and pull request. Matrix: `macos-latest`, `ubuntu-latest`, `windows-latest`. Steps: build, vet, test (with `-race` on macOS and Linux), golangci-lint.
- FR-3: `.github/workflows/release.yml` runs release-please on each push to `main`, and GoReleaser when release-please makes a release or when a person pushes a `v*` tag (ADR-012).
- FR-4: GoReleaser builds `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`, with `CGO_ENABLED=0`. It sets the version with `-ldflags "-X main.version=<tag>"`. Archive name: `claude-dx_<version>_<os>_<arch>.tar.gz` (`.zip` on Windows).
- FR-5: `scripts/install.sh` detects the OS and the architecture, downloads the latest release (or the version in the first argument), verifies the SHA-256 from `checksums.txt` and installs the binary into the install directory: `$CLAUDE_DX_INSTALL_DIR`, else `~/.local/bin` (Q-1).
- FR-6: `scripts/install.ps1` does the same on Windows and installs into `$env:CLAUDE_DX_INSTALL_DIR`, else `%LocalAppData%\Programs\claude-dx`, and adds it to the user `PATH` (Q-1).
- FR-7: `claude-dx --version` prints `claude-dx <version>`. A local build prints `claude-dx dev`.

### 6.2 Non-Functional
- NFR-1: The binary has no CGO, so cross-compilation works from one runner.
- NFR-2: CI time is less than 10 minutes for each platform.

## 7. Design
### 7.1 Architecture
```text
cmd/claude-dx/main.go          entry point, version variable
internal/cli/                  argument parser and dispatch (TASK-002)
internal/launcher/             find claude, build flags, start claude (TASK-002)
internal/plugin/               embedded plugin files and extraction (TASK-002)
internal/plugin/assets/        plugin.json, skills, agents, hooks, templates
internal/cache/                cache paths, cache clean (TASK-003)
internal/update/               release check, download, verify, replace (TASK-003)
internal/doctor/               environment checks (TASK-003)
internal/mcpserver/            MCP server and tool registration (TASK-004)
internal/project/              project root, project key, path guard (TASK-004)
internal/filelock/             OS file lock (TASK-004)
internal/numbering/            ADR and task number allocation (TASK-004)
internal/markdown/             frontmatter and section helpers (TASK-004)
internal/docs/{adr,task,learning,glossary}/   document tools (TASK-005 to TASK-007)
internal/git/                  git tools (TASK-009, TASK-010)
internal/hook/                 PreToolUse hook (TASK-011)
scripts/install.sh, scripts/install.ps1
.goreleaser.yaml, .golangci.yml, .github/workflows/{ci,release}.yml
README.md, LICENSE
```
Each package under `internal/` is small and has one responsibility, so a feature is easy to delete.

### 7.2 Public API
```text
claude-dx --version  -> "claude-dx <version>"
install.sh [version]
install.ps1 [version]
```
`install.ps1` takes a positional version, not `-Version`, because a `param` block is not reliable with `irm | iex`. `CLAUDE_DX_VERSION` also sets the version. Both scripts read `CLAUDE_DX_BASE_URL` to download from a local server in tests. If `claude-dx.exe` exists, `install.ps1` renames it to `claude-dx.exe.old` first (ADR-004).

### 7.3 Data and File Formats
`checksums.txt`: GoReleaser default format, `<sha256>  <archive name>`.

### 7.4 User-facing Behavior
```text
$ curl -fsSL https://raw.githubusercontent.com/is-equal/claude-dx/main/scripts/install.sh | sh
claude-dx v0.1.0 installed to ~/.local/bin/claude-dx
```

### 7.5 Alternatives
- Makefile or Taskfile for local commands: not needed. Plain `go` commands are sufficient.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `cmd/claude-dx/main.go`, `.golangci.yml`, `README.md`, `LICENSE` | Make `main` with `--version`. Add lint config. Add a short README and the MIT license (Q-2). | `go build ./... && go vet ./... && golangci-lint run` |
| 2 | `.github/workflows/ci.yml` | Add the matrix workflow of FR-2. | CI passes on the three platforms |
| 3 | `.goreleaser.yaml`, `.github/workflows/release.yml` | Add FR-3 and FR-4. | `goreleaser release --snapshot --clean` makes 6 archives and `checksums.txt` |
| 4 | `scripts/install.sh`, `scripts/install.ps1` | Add FR-5 and FR-6. | Install a snapshot release on macOS and Windows |

## 9. Acceptance Criteria
- AC-1: Given a push, when CI runs, then all steps pass on macOS, Linux and Windows.
- AC-2: Given the pre-release tag `v0.1.0-rc.0`, when the release workflow runs, then the GitHub pre-release has 6 archives and `checksums.txt`.
- AC-3: Given a changed archive, when `install.sh` runs, then the checksum check fails and nothing is installed.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| CI time for each platform | None | < 10 min | GitHub Actions run summary |

- Tests: CI run, snapshot release, manual install on macOS and Windows.
- Platforms not verified: Linux arm64 install (no runner).

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| PowerShell execution policy blocks `install.ps1` | Medium | Document `irm <url> \| iex` and the manual download |

## 12. Removal and Compatibility
Not applicable. This is the base of the project.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Install directory. Recommended: `~/.local/bin` on macOS and Linux, `%LocalAppData%\Programs\claude-dx` on Windows. Both can change with the environment variable `CLAUDE_DX_INSTALL_DIR`. | Maintainer | Accepted: `~/.local/bin` on macOS and Linux, `%LocalAppData%\Programs\claude-dx` on Windows, override with `CLAUDE_DX_INSTALL_DIR`. |
| Q-2 | License. Recommended: MIT. | Maintainer | MIT. |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done
- [x] All acceptance criteria pass, with evidence. AC-1 and AC-3: CI run 40226407455 (jobs `test` and `install` on macOS, Linux, Windows). AC-2: pre-release `v0.1.0-rc.0` (run 40227361491) has 6 archives and `checksums.txt`; the checksum of the darwin/arm64 archive matches and the binary prints `claude-dx v0.1.0-rc.0`.
- [x] Project checks (build, tests, lint) pass. CI run 40226407455.
- [x] Docs are updated (README). Install, Why and Roadmap sections.
- [x] The final Design diagrams that describe the code are in the docs of the area, with a link from this task. Not applicable: no diagram; the layout is in section 7.1.
- [x] ADRs are recorded for architectural decisions. ADR-012 (SemVer with release-please).
- [x] `/dx:eureka` is done. The tool does not exist yet. Lessons are code comments at their location: `.gitattributes` (CRLF breaks gofmt), `scripts/install.ps1` (exit closes the window with `irm | iex`), `.github/workflows/ci.yml` (macOS background process, Windows step exit code, Go cache key).
