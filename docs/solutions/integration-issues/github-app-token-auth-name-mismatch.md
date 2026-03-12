---
title: "Release automation failed due to GitHub App token contract mismatch"
module: "Release CI/CD"
problem_type: "integration_issue"
component: "cinzel/steps.hcl, .github/workflows/release.yaml, .github/workflows/release-published.yaml, .goreleaser.yaml"
severity: "high"
root_cause: "different release actions and GoReleaser expected the GitHub App token under different input and environment names"
resolution_type: "workflow_improvement"
symptoms:
  - "Manual release tagging failed with token/auth errors despite a generated GitHub App token"
  - "Published release packaging did not trigger or failed to publish Homebrew updates"
  - "GoReleaser failed templating homebrew token because HOMEBREW_TAP_GITHUB_TOKEN was missing"
tags:
  - "github-actions"
  - "github-app"
  - "release"
  - "goreleaser"
  - "homebrew"
  - "auth"
created_date: "2026-03-12"
updated_date: "2026-03-12"
---

## Problem

Release automation broke after migrating from PAT-based auth to a GitHub App installation token. The token existed, but different actions and GoReleaser expected it under different names, so some steps silently ignored it while others failed explicitly.

## Root Cause

The release flow treated GitHub App auth as one generic token contract, but the actual integrations differed:

- `mathieudutour/github-tag-action` requires `github_token`
- `ncipollo/release-action` requires `token`
- `orhun/git-cliff-action` expects `github_token`
- GoReleaser Homebrew publishing still resolves from `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`

Using the wrong name caused actions to ignore the installation token and fall back to defaults or fail with auth errors.

## What Didn't Work

**Attempted state:** pass the same input name to every action after moving to GitHub App auth.

- **Why it failed:** action interfaces are not uniform; a valid token value is useless if the action reads a different input/env name.

## Solution

Create one GitHub App installation token per workflow run, then map it to the exact contract each tool expects.

1. Validate release app secrets exist:
   - `RELEASE_APP_ID`
   - `RELEASE_PRIVATE_KEY`

2. Mint installation token with `actions/create-github-app-token`.

3. Pass that token to each step using the correct name:
   - `github-tag-action` -> `github_token`
   - `git-cliff-action` -> `github_token`
   - `release-action` -> `token`

4. Export the same token for GoReleaser packaging as:
   - `GITHUB_TOKEN`
   - `HOMEBREW_TAP_GITHUB_TOKEN`

5. Keep `.goreleaser.yaml` unchanged so `homebrew_casks.repository.token` continues resolving from `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`.

## Verification

Commands run while fixing and regenerating workflow output:

- `mise run cinzel github parse --directory ./cinzel --output-directory ./.github/workflows`
- `mise x actionlint@latest -- actionlint`

Observed result:

- Manual release workflow now uses the app token with action-specific input names.
- Published release workflow now exports the app token under `HOMEBREW_TAP_GITHUB_TOKEN` for GoReleaser Homebrew publishing.
- Generated workflow YAML is back in sync with `cinzel/*.hcl` source.

## Prevention

- Treat auth mapping as an explicit release contract, not a generic token assumption.
- Keep a per-action auth table in release docs when multiple third-party actions are chained together.
- Always review each action’s real input names during auth migrations or version bumps.
- Prefer one token source per workflow run, then map it deliberately to each downstream interface.
- Preserve `.goreleaser.yaml` env contracts unless there is a strong reason to change them in the same PR.

## Related

- `docs/solutions/integration-issues/workflow-release-integration-fix.md`
- `docs/solutions/integration-issues/release-automation-deprecation-trigger-contract-cleanup.md`
- `docs/solutions/integration-issues/git-cliff-action-offline-token-fix.md`
- `docs/plans/2026-03-11-ci-git-cliff-manual-release-automation-plan.md`
- `docs/release/homebrew.md`
