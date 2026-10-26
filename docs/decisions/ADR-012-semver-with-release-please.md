# ADR-012: SemVer with release-please

- Status: Accepted
- Date: 2026-10-25
- Owners: Maintainer

## Context
Releases need a version number. The maintainer wants Semantic Versioning (SemVer). All commits use Conventional Commits (ADR-008), so the commits give the type of each change. A tag that a GitHub Action makes with `GITHUB_TOKEN` does not start other workflows.

## Decision
- Versions obey SemVer 2.0.0: `MAJOR.MINOR.PATCH`, with an optional pre-release part (`-rc.N`, `-beta.N`). Tags have the prefix `v`.
- release-please (`googleapis/release-please-action`) runs on each push to `main`. It keeps one release pull request open with the next version and `CHANGELOG.md`.
- Bump rules: `feat` increases MINOR, `fix` and `perf` increase PATCH, `!` or `BREAKING CHANGE` increases MAJOR. Before 1.0.0, a breaking change increases MINOR (`bump-minor-pre-major`).
- The first release is 0.1.0 (`initial-version`), the v0.1 phase. Without this option, release-please starts at 1.0.0.
- The maintainer decides when to release: a merge of the release pull request makes the tag and the GitHub release.
- GoReleaser runs in the same workflow after release-please makes a release, and adds the files to that release (`mode: keep-existing`).
- A pre-release tag that a person pushes (example: `v0.1.0-rc.0`) runs GoReleaser only. GoReleaser marks it as a pre-release (`prerelease: auto`).
- Config: `release-please-config.json`, `.release-please-manifest.json`, `.github/workflows/release.yml`.

## Alternatives
- Manual tags: rejected. A person can select the wrong number, and there is no changelog file.
- A command that calculates the version (svu): rejected. No changelog file and no review step before a release.
- semantic-release: rejected. It releases at each push with no review step, and it needs Node.js configuration.

## Consequences
- Positive: the number comes from the commits. `CHANGELOG.md` is current. The release is one merge.
- Negative: the repository must allow GitHub Actions to create pull requests. Commit types must be correct, because they set the version.
- Follow-up: none.

## Evidence
- Design interview (2026-10-25): request to use SemVer; release-please selected.
- GitHub docs: events made with `GITHUB_TOKEN` do not create new workflow runs.
