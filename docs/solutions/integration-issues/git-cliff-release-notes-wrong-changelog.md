---
title: "git-cliff --latest shows wrong version in CI when tag is created remotely"
date: 2026-03-13
category: integration-issues
tags: [git-cliff, goreleaser, github-actions, release-automation, changelog]
severity: medium
components: [".github/workflows/release.yaml", "cliff.toml", "cinzel/steps.hcl"]
symptoms:
  - "GitHub release body contains entire CHANGELOG history instead of current release"
  - "GitHub release notes show previous version's changes after switching to --latest"
  - "git-cliff --latest --offline resolves to wrong tag in CI"
root_cause: "mathieudutour/github-tag-action creates tags via GitHub API (remote), so the local checkout lacks the new tag. git-cliff --latest with --offline sees the previous tag as latest, generating notes for the wrong range."
---

# git-cliff release notes showing wrong version in CI

## Problem symptoms

- GitHub release body contains the **entire CHANGELOG history** (all versions) instead of just the current release.
- After adding `--latest`, release notes show the **previous version's changes** instead of the current one.

## Investigation steps

1. Checked the release workflow — a single git-cliff step generated both `CHANGELOG.md` and the release body via `steps.git_cliff.outputs.content`.
2. Added `--latest` flag — still produced stale output.
3. Traced tag creation: `mathieudutour/github-tag-action` creates tags via the **GitHub API** (remote ref), not via `git tag` locally.
4. Confirmed `git-cliff` runs in `--offline` mode, meaning it only sees **local** git data.

Timeline in the workflow:

```
1. Checkout repo          → local repo has tags up to v0.0.6
2. github-tag-action      → creates v0.0.7 via GitHub API (remote only)
3. git-cliff --latest     → scans local tags, finds v0.0.6 as latest
                           → returns v0.0.6's changes (WRONG)
```

## Root cause

**Root cause 1 (full history):** A single git-cliff step generated the full changelog (for `CHANGELOG.md`) and its `content` output was used as the release body. Without `--latest` or `--unreleased`, git-cliff emits every tagged release.

**Root cause 2 (stale notes with `--latest`):** The `--latest` flag returns the changelog section for the most recent **existing** tag in the local repo. But `mathieudutour/github-tag-action` creates the new tag via the GitHub API (remote only). The local clone never receives this tag, so `--latest` resolves to the **previous** tag.

## Working solution

Split git-cliff into two separate steps:

1. **Full changelog** — writes `CHANGELOG.md`, has its own distinct `id`
2. **Release notes** — uses `--unreleased --tag <new_tag>`, keeps `id: git_cliff` so the release step reference works unchanged

### HCL (source of truth)

```hcl
step "git_cliff_changelog" {
  id   = "git_cliff_changelog"
  name = "Generate full changelog"

  uses {
    action  = "orhun/git-cliff-action"
    version = "c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72"
  }

  with {
    name  = "args"
    value = "--offline --verbose --tag $${{ steps.tag_version.outputs.new_tag }}"
  }

  env {
    name  = "OUTPUT"
    value = "CHANGELOG.md"
  }
}

step "git_cliff_release_notes" {
  id   = "git_cliff"
  name = "Generate release notes"

  uses {
    action  = "orhun/git-cliff-action"
    version = "c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72"
  }

  with {
    name  = "args"
    value = "--offline --verbose --unreleased --tag $${{ steps.tag_version.outputs.new_tag }}"
  }
}
```

### Key details

- The changelog step sets `OUTPUT: CHANGELOG.md` so it writes full history to file.
- The release notes step has **no `OUTPUT` env var** — content goes to stdout, captured by the action's `content` output.
- The release notes step keeps `id: git_cliff` so `${{ steps.git_cliff.outputs.content }}` in the release step picks up the correct output.

## Why `--unreleased` works but `--latest` does not

| Flag | Behavior | When tag is remote-only |
|------|----------|------------------------|
| `--latest` | Returns section for the **most recent existing tag** locally | Finds previous tag (v0.0.6), returns **stale** changes |
| `--unreleased` | Returns all commits **not covered by any local tag** | All commits since previous tag are "unreleased" — **correct** changes |
| `--unreleased --tag X` | Same as above, labels section with tag X | Correct changes, correctly labeled |

`--latest` is a **read** operation ("give me what the latest tag contains"). `--unreleased` is a **gap** operation ("give me everything not yet tagged"). When the new tag exists only on the remote, `--unreleased` captures exactly the right commit range.

## Prevention

- **Separate concerns**: distinct git-cliff calls for changelog vs release notes. Never reuse output from one for the other.
- **Prefer `--unreleased` over `--latest`** when tags are created remotely (via API, not `git tag`).
- **Validate release notes**: check non-empty, correct version reference, reasonable length before publishing.

## Related

- [git-cliff action offline token fix](./git-cliff-action-offline-token-fix.md)
- [GitHub App token auth name mismatch](./github-app-token-auth-name-mismatch.md)
- [Release automation deprecation cleanup](./release-automation-deprecation-trigger-contract-cleanup.md)
- [CI git-cliff manual release plan](../../plans/2026-03-11-ci-git-cliff-manual-release-automation-plan.md)
