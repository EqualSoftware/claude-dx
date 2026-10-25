# ADR-004: Install and update only from GitHub Releases

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
The project is published at `github.com/is-equal/claude-dx`. The maintainer does not want to publish to Homebrew, package managers or other places. The targets are macOS, Linux and Windows.

## Decision
- GoReleaser builds `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64` and `windows/arm64`. It publishes the archives and `checksums.txt` to GitHub Releases only.
- First install: `install.sh` (macOS, Linux) and `install.ps1` (Windows) in the repository download the release asset and verify its SHA-256. A manual download is also possible.
- The MCP server checks the latest release a maximum of one time in 24 hours, in the background. It writes the result to `<cache>/update.json`.
- At start, the launcher reads `<cache>/update.json`. If a new version exists, it shows one line: `claude-dx vX.Y.Z is available. Run: claude-dx update`. The launcher does no network call.
- `claude-dx update` downloads the asset for the OS and architecture, verifies the SHA-256 from `checksums.txt` and replaces the binary atomically. It never updates without the command.
- Windows locks a running `.exe`. `update` renames the current binary to `claude-dx.exe.old` and writes the new binary. At each start, on all platforms, the launcher deletes `<binary>.old` if it exists.
- No signature (minisign, cosign) in v1. Checksum only.

## Alternatives
- Homebrew tap, Scoop, `go install`: rejected by the maintainer.
- Check for updates in the launcher: rejected. It delays the start of Claude Code, and `syscall.Exec` ends the launcher process (ADR-003).

## Consequences
- Positive: one distribution channel. No start delay.
- Negative: no signature. A compromised GitHub account can publish a bad binary with a matching checksum.
- Follow-up: add a signature in a later version.

## Evidence
- Design interview (2026-10-25): Q-11, Q-12, Q-17, Q-18, Q-19.
