---
id: TASK-017
title: Prepare the repository for release
status: In Development
priority: P2
type: chore
size: S
owner: Equal
platforms: [Not applicable]
area: [repo]
depends_on: []
related: [ADR-008, ADR-012]
---

# TASK-017: Prepare the repository for release

## Document History
| Date | Author | Description / Change Notes |
| :--- | :--- | :--- |
| 2026-10-25 | Claude | Initial specification from the open source review. Decisions by the maintainer. |
| 2026-10-25 | Equal | Approved. |
| 2026-10-25 | Equal | Repository made public. FR-2, FR-3, FR-5, FR-6 applied. FR-4 needs the web UI (no API). |
| 2026-10-25 | Equal | Reported content (FR-4) exists only for organization repositories. Repository moved to the EqualSoftware organization; module path is now `github.com/EqualSoftware/claude-dx`. Secret scanning and CodeQL were off after the transfer and are on again. Code owner: `@EqualSoftware/developers`. |
| 2026-10-25 | Equal | FR-4: reported content set to "all users" in the repository settings (confirmed in the web UI). The community profile API still shows `content_reports_enabled: false` and 87 %; check again before Done. |

## 1. Summary
Make the repository public and turn on the GitHub features that need a public repository: the main branch ruleset, private vulnerability reporting, reported content, secret scanning and CodeQL.

## 2. Problem and Motivation
The repository is private. On the free plan, these features work only for public repositories. The community files (`CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`) refer to some of them.

## 3. Goals
- G-1: The repository is public with all features of section 6 on.
- G-2: The GitHub Community Standards page shows all items complete.

## 4. Out of Scope
- The first release (v0.1.0): the maintainer merges the release-please pull request when v0.1 is complete.

## 5. Context
### 5.1 Current State
The repository has rebase-only merges and branch deletion on merge. A ruleset request returned HTTP 403: "Upgrade to GitHub Pro or make this repository public to enable this feature."

### 5.2 Assumptions
- A-1: `github.com/is-equal/claude-equal-workflow` is public before this task, so the reference links in tasks and ADRs work. Verify: open a link with no login.

### 5.3 Constraints
- ADR-008: rebase only, no force push to the main branch.

## 6. Requirements
### 6.1 Functional
- FR-1: Make the repository public.
- FR-2: Ruleset "Protect main" on the default branch: rules `deletion`, `non_fast_forward`, `required_linear_history`. No required pull request and no required status check, so the maintainer can push to `main`.
- FR-3: Turn on private vulnerability reporting (`SECURITY.md` refers to it).
- FR-4: Turn on reported content for all users (`CODE_OF_CONDUCT.md` refers to it).
- FR-5: Turn on secret scanning with push protection and Dependabot alerts.
- FR-6: Turn on CodeQL default setup for Go and GitHub Actions.

### 6.2 Non-Functional
- Not applicable.

## 7. Design
### 7.1 Architecture
Not applicable.

### 7.2 Public API
Not applicable.

### 7.3 Data and File Formats
Ruleset request body:
```json
{"name": "Protect main", "target": "branch", "enforcement": "active",
 "conditions": {"ref_name": {"include": ["~DEFAULT_BRANCH"], "exclude": []}},
 "rules": [{"type": "deletion"}, {"type": "non_fast_forward"}, {"type": "required_linear_history"}]}
```

### 7.4 User-facing Behavior
Visitors see the Security tab with "Report a vulnerability" and the "Report content" option.

### 7.5 Alternatives
- Required pull requests for the maintainer: rejected by the maintainer. Each change would need a branch and a pull request.

## 8. Implementation Plan
| Work batch | Files | Steps | Check |
| :--- | :--- | :--- | :--- |
| 1 | None (repository settings) | FR-1 to FR-6 with `gh repo edit` and `gh api`. | `gh api repos/EqualSoftware/claude-dx/rulesets`; Community Standards page |

## 9. Acceptance Criteria
- AC-1: Given the public repository, when a person pushes with force to `main`, then GitHub refuses the push.
- AC-2: Given the Community Standards page, when it loads, then all items are complete.

## 10. Validation and Metrics
| Metric | Baseline | Target | How to measure |
| :--- | :--- | :--- | :--- |
| Community Standards items complete | Not available (private) | All | Insights → Community Standards |

- Tests: manual checks.
- Platforms not verified: not applicable.

## 11. Risks and Mitigations
| Risk | Impact | Mitigation |
| :--- | :--- | :--- |
| A secret is in the git history | High | Run secret scanning; check its results before the announcement |

## 12. Removal and Compatibility
Not applicable.

## 13. Open Questions
| ID | Question | Owner | Answer |
| :--- | :--- | :--- | :--- |
| Q-1 | None. | — | — |

## 14. Definition of Done
- [ ] All acceptance criteria pass, with evidence.
- [ ] Project checks (build, tests, lint) pass.
- [ ] Docs are updated.
- [ ] The final Design diagrams that describe the code are in the docs of the area, with a link from this task.
- [ ] ADRs are recorded for architectural decisions.
- [ ] `/dx:eureka` is done.
