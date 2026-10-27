# ADR-005: Release and update

- Status: Accepted
- Date: 2026-10-27
- Owners: Maintainer

## Context
The project is published at `github.com/EqualSoftware/claude-dx`. The maintainer does not want to publish to Homebrew, package managers or other places. The targets are macOS, Linux and Windows. Releases need a version number, and the maintainer wants Semantic Versioning (SemVer). All commits use Conventional Commits, so the commits give the type of each change. A tag that a GitHub Action makes with `GITHUB_TOKEN` does not start other workflows.

## Decision
Version and release:
- Versions obey SemVer 2.0.0: `MAJOR.MINOR.PATCH`, with an optional pre-release part (`-rc.N`, `-beta.N`). Tags have the prefix `v`.
- release-please (`googleapis/release-please-action`) runs on each push to `main`. It keeps one release pull request open with the next version and `CHANGELOG.md`.
- Bump rules: `feat` increases MINOR, `fix` and `perf` increase PATCH, `!` or `BREAKING CHANGE` increases MAJOR. Before 1.0.0, a breaking change increases MINOR (`bump-minor-pre-major`).
- The first release is 0.1.0 (`initial-version`). Without this option, release-please starts at 1.0.0.
- The maintainer decides when to release: a merge of the release pull request makes the tag and the GitHub release.
- GoReleaser runs in the same workflow after release-please makes a release, and adds the files to that release (`mode: keep-existing`).
- A pre-release tag that a person pushes (example: `v0.1.0-rc.0`) runs GoReleaser only. GoReleaser marks it as a pre-release (`prerelease: auto`).

Distribution:
- GoReleaser builds `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64` and `windows/arm64`. It publishes the archives and `checksums.txt` to GitHub Releases only.
- First install: `install.sh` (macOS, Linux) and `install.ps1` (Windows) in the repository download the release asset and verify its SHA-256. A manual download is also possible.
- The MCP server checks the latest release a maximum of one time in 24 hours, in the background. It writes the result to `<cache>/update.json`.
- At start, the launcher reads `<cache>/update.json`. If a new version exists, it shows one line: `claude-dx vX.Y.Z is available. Run: claude-dx update`. The launcher does no network call.
- `claude-dx update` downloads the asset for the OS and architecture, verifies the SHA-256 from `checksums.txt` and replaces the binary: it writes the new binary to a temporary file in the same directory and renames it over the old one. It never updates without the command. A `dev` build does not update.
- Windows locks a running `.exe`. `update` renames the current binary to `claude-dx.exe.old` and writes the new binary. At each start, on all platforms, the launcher deletes `<binary>.old` if it exists.
- `claude-dx update` checks the checksum only. It does not check a signature.

## Alternatives
- Homebrew tap, Scoop, `go install`: Rejected by the maintainer.
- Check for updates in the launcher: Rejected. It delays the start of Claude Code.
- Manual tags: Rejected. A person can select the wrong number, and there is no changelog file.
- A command that calculates the version (svu): Rejected. No changelog file and no review step before a release.
- semantic-release: Rejected. It releases at each push with no review step, and it needs Node.js configuration.

## Consequences
- Positive: one distribution channel. No start delay. The version number comes from the commits. `CHANGELOG.md` is current. The release is one merge.
- Negative: the repository must allow GitHub Actions to create pull requests. Commit types must be correct, because they set the version. `update` has no signature check: a compromised GitHub account can publish a bad binary with a matching checksum.
- Follow-up: add a signature check to `update` later.

## Evidence
- release-please-config.json:5-15 (`initial-version`, `bump-minor-pre-major`, changelog sections), .release-please-manifest.json.
- .github/workflows/release.yml:26-37 (release-please job), :39-72 (GoReleaser job and build provenance attestation for each release file).
- .goreleaser.yaml:10-11 (targets), :27-29 (`checksums.txt`), :48-50 (`prerelease: auto`, `mode: keep-existing`).
- internal/update/update.go:20-27 (constants), :107-119 (check once in 24 hours), :122-128 (start notice).
- internal/update/apply.go:23-45 (download and replace), :65-80 (SHA-256 check), :122-155 (temporary file and rename, `.old` file on Windows).
- internal/launcher/launcher.go:30-31 (delete `<binary>.old`), :37-40 (read the notice only).
- scripts/install.sh:75-80 and scripts/install.ps1:45 (checksum check at first install).
- GitHub docs: events made with `GITHUB_TOKEN` do not create new workflow runs.
