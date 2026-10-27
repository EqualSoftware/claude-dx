---
topic: github-repository-checks
description: How to check GitHub repository settings, branch rules and releases. Use it when you verify the community profile or a ruleset of a repository, or delete a release.
---

- The community profile API (`repos/{owner}/{repo}/community/profile`) shows `issue_template: null` when the repository has only issue forms (`.github/ISSUE_TEMPLATE/*.yml`). Do not use this field as proof that issue templates are missing.
- After you set reported content to "all users", the API can still show `content_reports_enabled: false`, and the Community Standards page can still show the item as not complete. The repository settings page is the source of truth.
- To test a rule that blocks force push, push a commit that is older than the remote head (`git push --force origin <older-commit>:main`). If the remote is behind the local branch, `HEAD~1` can be a fast-forward: the push succeeds and does not test the rule. A real refusal gives `GH013 ... Cannot force-push to this branch`.
- `gh release delete <tag> --cleanup-tag` run in a clone deletes the release, the remote tag and also the local tag. A later `git tag -d <tag>` fails with "tag not found". Verify with `gh release list`, `git ls-remote --tags origin` and `git tag -l`.
- `releases/latest` of the GitHub API does not give pre-releases, so install scripts and update checks that use it never install a pre-release.
