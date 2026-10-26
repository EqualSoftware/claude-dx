---
id: TASK-046
title: Build provenance for release files
status: In Review
priority: P3
milestone: v0.1.0
type: ci
size: S
owner: Equal
platforms: [Not applicable]
area: [ci]
depends_on: [TASK-041]
related: [ADR-004, ADR-012, TASK-001]
---

# TASK-046: Build provenance for release files

## Document History

| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Equal | Initial draft specification |
| 2026-10-26 | Equal | Refined with /dx:grill. The maintainer confirmed the same understanding. |

## 1. Summary

The release workflow adds a GitHub build provenance attestation for each release archive and for `checksums.txt`. The README and `SECURITY.md` tell how to verify it.

## 2. Problem and Motivation

`claude-dx update` (`internal/update/apply.go`) and the install scripts verify the SHA-256 with `checksums.txt`, but that file comes from the same release. A person who changes the release can change both files. An attestation proves that the files come from the release workflow of this repository.

## 3. Goals

- G-1: Each release file has an attestation that `gh attestation verify` accepts.

## 4. Out of Scope

- A verification of the attestation in `claude-dx update` or in the install scripts: it needs `gh` or a new library (grill Q-13).

## 5. Context
### 5.1 Current State

The `goreleaser` job (`.github/workflows/release.yml`) makes the files in `dist/`. The workflow pins each action by commit SHA.

### 5.2 Assumptions

- A-1: `actions/attest-build-provenance` works for a public repository on the free plan. Verify: the GitHub docs of artifact attestations.

### 5.3 Constraints

- ADR-004: releases on GitHub Releases.
- Pin the new action by commit SHA, as the other actions.

## 6. Requirements
### 6.1 Functional

- FR-1: After GoReleaser, the `goreleaser` job runs `actions/attest-build-provenance` with the subjects `dist/*.tar.gz`, `dist/*.zip` and `dist/checksums.txt`.
- FR-2: The job gets the permissions `id-token: write` and `attestations: write`.
- FR-3: `README.md` and `SECURITY.md` give the command `gh attestation verify <file> --repo EqualSoftware/claude-dx`.

### 6.2 Non-Functional

- NFR-1: Not applicable.

## 7. Design
### 7.1 Architecture

One step in `.github/workflows/release.yml`.

### 7.2 Public API

Not applicable.

### 7.3 Data and File Formats

Not applicable.

### 7.4 User-facing Behavior

A person can run `gh attestation verify` on a downloaded file.

### 7.5 Alternatives

- Sign with cosign: rejected, the GitHub attestation needs no key.

## 8. Implementation Plan

| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | `.github/workflows/release.yml` | FR-1, FR-2. | `actionlint`, or the workflow syntax check of GitHub |
| 2 | `README.md`, `SECURITY.md` | FR-3. | Read the docs |

## 9. Acceptance Criteria

- AC-1: Given the first release after this task, when a person runs `gh attestation verify` on an archive and on `checksums.txt`, then the command passes.

## 10. Validation and Metrics

| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Release files with an attestation | 0 | All | `gh attestation verify` on each file |

- Tests: manual check at the first release.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations

| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| The step fails and stops the release | Medium | The step runs after GoReleaser. A failure leaves the release files, and the step can run again. |
| AC-1 can be checked only at a release | Low | Use a pre-release tag (example `v0.1.0-rc.1`) to check it before v0.1.0. |

## 12. Removal and Compatibility

Delete the step and the doc lines.

## 13. Open Questions

| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | Add the attestation in CI and docs only, with no check in `claude-dx update`? | Equal | Yes (grill Q-13). |
| Q-2 | Check AC-1 with a pre-release tag before v0.1.0? | Equal | Yes, `v0.1.0-rc.1` (grill Q-29). |

The status cannot be `Approved` while a question is `Open`.

## 14. Definition of Done

- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated (`README.md`, `SECURITY.md`).
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.