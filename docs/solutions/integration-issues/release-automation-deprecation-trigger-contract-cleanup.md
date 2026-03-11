---
title: "Release automation deprecation migration and trigger contract cleanup"
module: "Release CI/CD"
problem_type: "integration_issue"
component: ".goreleaser.yaml, cinzel/*.hcl, .github/workflows/release.yaml, mise.toml"
severity: "high"
root_cause: "config_error"
symptoms:
  - "GoReleaser configuration still used deprecated brews instead of homebrew_casks"
  - "Release workflow had diverging manual and published trigger behaviors"
  - "Local VERSION overrides could keep a leading v and drift from release version normalization"
resolution_type: "workflow_improvement"
tags:
  - "goreleaser"
  - "homebrew-casks"
  - "github-actions"
  - "release"
  - "workflow-dispatch"
  - "versioning"
created_date: "2026-03-11"
updated_date: "2026-03-11"
---

## Problem

Release automation had drifted from current contracts. The repository still used deprecated GoReleaser Homebrew configuration and maintained a manual dry-run workflow branch that no longer matched the intended production release path.

In parallel, local build inputs accepted `VERSION=vX.Y.Z` without full normalization, which could produce inconsistent version strings versus tag-based release artifacts.

## Root Cause

- Release/distribution config evolved incrementally and left deprecated `brews` in place.
- Release workflow retained dual behavior (`release.published` and manual dry-run behavior), reducing determinism.
- `VERSION` normalization stripped `v` for tag-derived fallback but not for user-provided env overrides.

## What Didn't Work

**Attempted Solution 1:** Add cask conflict metadata using `homebrew_casks.conflicts.formula`.
- **Why it failed:** `goreleaser check` reported this property as deprecated/no-op, so it was removed to keep the config clean and future-safe.

## Solution

Applied one cohesive migration across release config, workflow source, generated workflow output, docs, and local build tasks.

1. Migrate GoReleaser Homebrew integration to casks:
   - `brews` -> `homebrew_casks`
   - add explicit `binaries: [cinzel]`
   - keep existing tap repo/token metadata

2. Remove release dry-run contract from published-release workflow sources:
   - delete `workflow_dispatch` trigger in `cinzel/workflows.hcl`
   - simplify release concurrency/job conditions to published release context
   - set GoReleaser action args to a single path: `release --clean`

3. Update release observability wording:
   - change formula-centric summary labels to Homebrew artifact wording
   - keep artifact discovery over `dist/*.rb`

4. Normalize local version overrides:
   - add `VERSION="${VERSION#v}"` in local build version paths

5. Regenerate workflows from HCL source of truth.

6. Update operator/user docs:
   - Homebrew docs now describe cask integration and no dry-run section
   - README install command uses `brew install --cask cinzel`

## Verification

Executed and validated during the fix:

- `mise run cinzel github parse --directory ./cinzel --output-directory ./.github/workflows`
- `mise run test-ci`
- `go run github.com/goreleaser/goreleaser/v2@v2.14.3 check`
- `mise x actionlint@latest -- actionlint`
- `mise x ghalint@latest -- ghalint run`

Observed outcomes:

- GoReleaser config validates successfully after removing deprecated conflict field.
- Published-release workflow no longer contains snapshot skip arguments.
- CI tests and workflow lint checks pass (only non-blocking local `mise` warning about missing `tmux` surfaced).

## Current State Note

Subsequent changes introduced a dedicated manual release workflow (`workflow_dispatch`) while keeping published-release packaging in a separate workflow file. This solution remains correct for the deprecated `brews` migration and published-path cleanup context, and current release topology now uses both `release.yaml` (manual) and `release-published.yaml` (published packaging).

## Prevention

- Keep `cinzel/*.hcl` as the only workflow source of truth; regenerate YAML and review generated diff.
- Add CI guardrails to fail on deprecated GoReleaser keys (especially `brews`).
- Add CI check to block release dry-run trigger/args from reappearing in release workflow.
- Keep GoReleaser action SHA and tool version pinned together and validated in the same change.
- Treat version normalization as a contract in every local release/build entrypoint.

## Related

- `docs/solutions/integration-issues/workflow-release-integration-fix.md`
- `docs/release/homebrew.md`
- `docs/plans/2026-03-09-feat-release-package-distribution-plan.md`
- `docs/plans/2026-03-11-chore-goreleaser-homebrew-cask-migration-plan.md`
- `docs/plans/2026-03-11-feat-release-distribution-follow-up-plan.md`
- `docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md`
