---
title: "Workflow release integration hardening with GoReleaser and mise"
module: "Release CI/CD"
problem_type: "integration_issue"
component: ".github/workflows, cinzel/*.hcl, .goreleaser.yaml"
severity: "high"
root_cause: "release workflow logic was fragmented and inconsistent across triggers, tooling setup, permissions, and publishing behavior"
symptoms:
  - "release path used mixed tooling and manual glue"
  - "trigger semantics were weaker than required for production releases"
  - "workflow security posture lacked consistent least-privilege and checkout hardening"
  - "pipeline setup/cache strategy was inconsistent across jobs"
tags:
  - "github-actions"
  - "goreleaser"
  - "mise"
  - "release"
  - "security"
  - "permissions"
  - "integration"
created_date: "2026-03-11"
updated_date: "2026-03-11"
---

## Problem

The release pipeline had become hard to reason about and easy to drift. Release behavior, packaging assumptions, and Homebrew distribution were spread across workflow logic, and the setup/security model was inconsistent.

Key pain points included trigger mismatch, mixed setup patterns, and no clear dry-run contract for validating release workflow changes safely.

## Root Cause

- Workflow behavior was not fully consolidated around one release tool contract.
- CI setup and caching were not uniformly modeled per workflow/job.
- Security controls (permissions, checkout credential persistence, action pinning) required explicit hardening.
- Release flow and docs/plan checkpoints needed synchronization.

## Solution

Refactored release integration to a single coherent path:

1. Switched release packaging/publishing to GoReleaser:
   - added `.goreleaser.yaml`
   - replaced custom release step logic with `goreleaser/goreleaser-action`

2. Kept workflow source-of-truth in HCL only:
   - updated `cinzel/workflows.hcl`, `cinzel/jobs.hcl`, `cinzel/steps.hcl`
   - regenerated `.github/workflows/*.yaml` from cinzel sources

3. Hardened workflow execution model:
   - release trigger moved to `release.published`
   - added `workflow_dispatch` dry-run path (snapshot + skip publish/announce/validate)
   - added release concurrency control keyed by release context

4. Standardized setup path on mise:
   - added `jdx/mise-action` setup in relevant jobs
   - enabled `install=true` and `cache=true`
   - removed redundant separate Go setup path

5. Improved security and policy compliance:
   - explicit least-privilege permissions at workflow/job level
   - `persist-credentials: false` on checkout steps
   - pinned external actions to commit SHAs and documented corresponding semantic versions in HCL comments
   - fixed `ghalint` findings (timeouts and checkout policy)

6. Synced plan progress:
   - advanced relevant acceptance/reliability checkboxes in
     `docs/plans/2026-03-09-feat-release-package-distribution-plan.md`

## Verification

Executed after each refactor stage:

- `mise run cinzel github parse --directory ./cinzel --output-directory ./.github/workflows`
- `mise run test-ci`
- `mise x actionlint -- actionlint`
- `ghalint run`

All checks passed.

## Prevention

- Keep `cinzel/*.hcl` as the only editable workflow source; treat generated YAML as outputs.
- Require SHA-pinned action refs and explicit least-privilege permissions.
- Keep release dry-run (`workflow_dispatch`) as a first-class contract for safe validation.
- Require setup/cache consistency (`mise-action` with cache enabled) across CI/release jobs.
- Add guard checks for:
  - unpinned actions
  - floating tool versions in release workflows
  - missing `timeout-minutes`
  - checkout steps without `persist-credentials: false`

## Related

- `docs/plans/2026-03-09-feat-release-package-distribution-plan.md`
- `docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md`
- `docs/solutions/logic-errors/license-and-provider-docs-consistency.md`
- `docs/solutions/logic-errors/style-guidance-rule-interpretation-fix.md`
