---
slug: release
title: Release pipeline and install scripts
status: Approved
milestone: v0.1.0
priority: P1
area: [release, ci]
related: [ADR-005]
---

# Release pipeline and install scripts

## 1. Summary

The release pipeline builds, tests and publishes claude-dx. Each push runs the checks on macOS, Linux and Windows. Each release publishes the binaries with checksums and build provenance, and the install scripts download the release.

## 2. Problem

The user installs claude-dx with no package manager (ADR-005). The binary must build on each platform from one runner, and each file must be verifiable. A release must follow the version rules of ADR-005.

## 3. Goals

- G-1: Each push passes the checks on macOS, Linux and Windows.
- G-2: Each release has six archives and checksums.txt.
- G-3: The install scripts check the SHA-256 and install nothing when the check fails.
- G-4: A person can verify each release file with gh attestation verify.

## 4. Out of Scope

- Package managers (Homebrew, winget and others).
- Signature verification with a key of the project (ADR-005).
- The update of the binary. See the spec cli-update.

## 5. Requirements

| ID | Requirement | Status |
| :--- | :--- | :--- |
| R-1 | Each push and pull request runs build, vet, tests and golangci-lint on macOS, Linux and Windows. The race tests run on macOS and Linux. | Done |
| R-2 | Each push to main runs release-please, which keeps a release pull request open. | Done |
| R-3 | GoReleaser runs when release-please makes a release, or when a person pushes a tag v*. | Done |
| R-4 | GoReleaser builds six targets (darwin, linux and windows, each on amd64 and arm64) with CGO_ENABLED=0. The archives are named claude-dx_<version>_<os>_<arch>, and checksums.txt is made. | Done |
| R-5 | The release build sets the version with the ldflag main.version. claude-dx --version prints claude-dx <version>, and a local build prints claude-dx dev. | Done |
| R-6 | scripts/install.sh detects the OS and the architecture, downloads the latest release or the version of the first argument, verifies the SHA-256 from checksums.txt, and installs to CLAUDE_DX_INSTALL_DIR or ~/.local/bin. | Done |
| R-7 | scripts/install.ps1 does the same on Windows, installs to CLAUDE_DX_INSTALL_DIR or %LocalAppData%\Programs\claude-dx, and adds that directory to the user PATH. | Done |
| R-8 | The release job adds a build provenance attestation for each archive and for checksums.txt. README.md and SECURITY.md give the command gh attestation verify. | Done |
| R-9 | The repository is public, with a ruleset on main that refuses deletion and non-fast-forward pushes, and with private vulnerability reporting, secret scanning and CodeQL on. | Done |

## 6. Design

### Files

- `.github/workflows/ci.yml`: the matrix of macos-latest, ubuntu-latest and windows-latest. Steps: build, vet, test, golangci-lint (race tests on macOS and Linux only).
- `.github/workflows/release.yml`: the `release-please` job on each push to main, and the `goreleaser` job. The `goreleaser` job runs for a tag or a release made by release-please, and it runs `actions/attest` for `dist/*.tar.gz`, `dist/*.zip` and `dist/checksums.txt`.
- `.goreleaser.yaml`: six targets with `CGO_ENABLED=0`, the archive names and `checksums.txt`.
- `scripts/install.sh` and `scripts/install.ps1`: download, SHA-256 check and install.

### Data flow

1. A push to main makes or updates the release pull request (release-please).
2. A release or a tag starts GoReleaser. It builds the archives and checksums.txt.
3. The attest step signs the build provenance of each file.
4. The install script downloads the archive for the platform and checks it against checksums.txt.

### Decisions

- ADR-005: release with SemVer and release-please, and install and update from GitHub releases.

## 7. Acceptance Criteria

Done requirements:
- R-1: the CI run of `.github/workflows/ci.yml` on macOS, Linux and Windows. No Go test covers the workflow.
- R-2 and R-3: the release workflow run after a push to main or a tag. Checked by the workflow log.
- R-4: a release run gives six archives and `checksums.txt`. Checked by the GoReleaser result.
- R-5: `TestRunVersion` (`cmd/claude-dx/main_test.go`) for `claude-dx --version`. The ldflag is checked by the release run.
- R-6 and R-7: manual check. Run `scripts/install.sh` and `scripts/install.ps1` with a changed archive. The check must fail and install nothing.
- R-8: manual check. Run `gh attestation verify` on one archive and on `checksums.txt`.
- R-9: manual check of the GitHub settings. A forced push to main is refused, and the Community Standards page shows all items complete.

Planned requirements: none.

## 8. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| — | None. | — | — |

The status cannot be `Approved` while a question is `Open`.

## 9. Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-27 | Equal | Initial draft specification |
| 2026-10-27 | Equal | Added R-1. |
| 2026-10-27 | Equal | Added R-2. |
| 2026-10-27 | Equal | Added R-3. |
| 2026-10-27 | Equal | Added R-4. |
| 2026-10-27 | Equal | Added R-5. |
| 2026-10-27 | Equal | Added R-6. |
| 2026-10-27 | Equal | Added R-7. |
| 2026-10-27 | Equal | Added R-8. |
| 2026-10-27 | Equal | Added R-9. |
| 2026-10-27 | Equal | R-1: status Done. |
| 2026-10-27 | Equal | R-2: status Done. |
| 2026-10-27 | Equal | R-3: status Done. |
| 2026-10-27 | Equal | R-4: status Done. |
| 2026-10-27 | Equal | R-5: status Done. |
| 2026-10-27 | Equal | R-6: status Done. |
| 2026-10-27 | Equal | R-7: status Done. |
| 2026-10-27 | Equal | R-8: status Done. |
| 2026-10-27 | Equal | R-9: status Done. |
| 2026-10-27 | Equal | Approved by the user. |
| 2026-10-27 | Equal | Changed fields: related. |
| 2026-10-27 | Equal | Status: In Review. The scope changed in section 4. |
| 2026-10-27 | Equal | Approved by the user after the reference change (ADR-012). |