---
title: "feat: Release package distribution"
type: feat
status: completed
date: 2026-03-09
origin: docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md
---

# feat: Release package distribution

Follow-up scope is tracked in `docs/plans/2026-03-11-feat-release-distribution-follow-up-plan.md`.

Historical note: this completed plan captures the original `homebrew-cinzel` tap direction. Current implementation targets `yldio/cinzel` for Homebrew automation.

## Overview

Automate Homebrew release distribution for cinzel using a dedicated `homebrew-cinzel` tap, while also formalizing Windows release support via native Windows distribution channels. On GitHub release publication, publish trusted release artifacts, compute SHA256 checksums, update the tap formula, and create or update a PR in the tap repository so users can install with `brew install cinzel` on macOS/Linux, with Windows delivered via release artifacts and a Windows package manager path.

## Problem Statement / Motivation

cinzel has no automated Homebrew distribution path today, so releases require manual tap updates and checksum handling. Manual formula edits are repetitive and error-prone, and release readiness issues (workflow build path mismatch, CI test command mismatch) weaken trust in produced artifacts. The brainstorm selected a release-driven tap update model to ship quickly with minimal architecture changes while removing ongoing release toil.

## Research Decision

External research was added because this feature depends on cross-repo release automation and Homebrew formula correctness. Official GitHub Actions and Homebrew guidance clarified safer release event handling, permission boundaries, and formula structure.

## Consolidated Findings

- Preferred distribution target is a dedicated tap repository (`homebrew-cinzel`), not Homebrew core in v1.
- Trigger model should be GitHub `release.published` (safer for draft/prerelease behavior) using release metadata and attached assets as source of truth.
- v1 platform scope is macOS + Linux for Homebrew formula updates, plus Windows release artifacts.
- Prerequisite reliability fixes are required before enabling tap automation.
- Homebrew formula needs explicit OS/arch handling and deterministic per-asset SHA256 mapping.
- Cross-repo PR updates should use least-privilege credentials (GitHub App preferred) and idempotent branch/PR updates.

## Proposed Solution

Implement a release pipeline extension that, after successful release artifact publication, computes checksums and opens/updates a formula PR in the tap repository.

### Scope and Contract

- Keep release events as orchestration entrypoint (no tag-system redesign in this phase) (see brainstorm: docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md).
- Use release asset URLs + SHA256 checksums to build formula updates deterministically.
- Automate tap updates through pull requests (no direct default-branch writes).
- Include release-readiness fixes as part of this plan before turning on tap automation.
- Do not edit `.github/workflows/*.yaml` directly; update workflow sources in `cinzel/*.hcl` and regenerate YAML through cinzel.
- Homebrew formula scope remains macOS/Linux only (Homebrew platform constraint). Windows support is delivered via release artifacts and a dedicated Windows package-manager flow.

### Artifact Contract (fail-fast)

Release automation only proceeds when all required assets for one release tag are present and match expected naming.

- Required v1 targets:
  - `darwin-amd64`
  - `darwin-arm64`
  - `linux-amd64`
  - `linux-arm64`
  - `windows-amd64`
  - `windows-arm64`
- Required filename pattern (contract):
  - `cinzel_<version>_<os>_<arch>.tar.gz`
- Required companion checksums source:
  - generated SHA256 per required asset from release artifacts (not from local rebuild)
- Fail-fast rules:
  - missing required asset => fail job with explicit missing target list
  - duplicate target match => fail job
  - filename not matching contract => fail job with offending filenames
  - checksum generation mismatch => fail job and do not open/update tap PR

### Formula Contract (v1)

Formula must select URL/SHA256 by OS and architecture and keep deterministic layout.

- Use `on_macos` / `on_linux` blocks.
- Inside each OS block, use `on_intel` / `on_arm` to map URL + SHA256.
- `version` in formula must equal GitHub release tag version.
- Formula source of truth is the release asset URL set for the current tag.

Illustrative generated formula shape:

```ruby
class Cinzel < Formula
  desc "Bidirectional converter between HCL and CI/CD YAML"
  homepage "https://github.com/yldio/cinzel"
  version "1.2.3"

  on_macos do
    on_intel do
      url "https://github.com/yldio/cinzel/releases/download/v1.2.3/cinzel_1.2.3_darwin_amd64.tar.gz"
      sha256 "<sha256-darwin-amd64>"
    end
    on_arm do
      url "https://github.com/yldio/cinzel/releases/download/v1.2.3/cinzel_1.2.3_darwin_arm64.tar.gz"
      sha256 "<sha256-darwin-arm64>"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/yldio/cinzel/releases/download/v1.2.3/cinzel_1.2.3_linux_amd64.tar.gz"
      sha256 "<sha256-linux-amd64>"
    end
    on_arm do
      url "https://github.com/yldio/cinzel/releases/download/v1.2.3/cinzel_1.2.3_linux_arm64.tar.gz"
      sha256 "<sha256-linux-arm64>"
    end
  end

  def install
    bin.install "cinzel"
  end
end
```

### Windows Package Contract (winget)

Windows package-manager support uses **winget manifests** generated from the same release artifact/checksum contract.

- Preferred Windows channel: `winget` (phase 5 implementation target).
- Package identifier contract: `YLD.Cinzel` (final identifier validated during implementation).
- Manifest contract (minimum):
  - `manifests/y/YLD/Cinzel/<version>/YLD.Cinzel.installer.yaml`
  - `manifests/y/YLD/Cinzel/<version>/YLD.Cinzel.locale.en-US.yaml`
  - `manifests/y/YLD/Cinzel/<version>/YLD.Cinzel.yaml`
- Installer mapping contract:
  - `windows-amd64` asset URL + SHA256 -> winget x64 installer entry
  - `windows-arm64` asset URL + SHA256 -> winget arm64 installer entry
- PR target contract:
  - open/update PR in `microsoft/winget-pkgs` (or fork + upstream PR), no direct writes
- Operational note:
  - winget availability is asynchronous after PR merge; release success does not block on winget merge latency.

## Technical Approach

### Phase 1: Release Reliability Prerequisites

- Fix release workflow build/output path mismatches so published assets match expected names and locations.
- Fix CI test command mismatch to ensure release gating uses the canonical, passing test command.
- Add/adjust validation gates so release automation only runs after green checks.
- Switch release trigger semantics to `published` and add explicit draft/prerelease guards where needed.

Planned touchpoints:
- `cinzel/workflows.hcl`
- `cinzel/jobs.hcl`
- generated workflow outputs under `.github/workflows/*.yaml` (regenerated, not hand-edited)
- `mise.toml` (if task command alignment is needed)

### Phase 2: Homebrew Update Automation

- Switch release packaging and Homebrew publishing to GoReleaser:
  - build/publish release artifacts for required OS/arch targets,
  - generate release checksums,
  - render/update the `cinzel` formula in the tap repo,
  - open/update tap PR via GoReleaser brew integration.
- Keep formula updates idempotent so reruns do not create noisy diffs.
- Ensure generated formula references official release artifacts only.

Planned touchpoints:
- `cinzel/workflows.hcl`
- `cinzel/jobs.hcl`
- `cinzel/steps.hcl`
- `.goreleaser.yaml`
- generated workflow outputs under `.github/workflows/*.yaml` (regenerated, not hand-edited)
- `docs/release/` (operator notes)

### Phase 3: Secrets, Permissions, and Observability

- Configure required credentials for tap PR creation (GitHub App token preferred; fine-grained PAT fallback) with least-privilege scopes.
- Explicitly set workflow permissions needed for release assets and PR operations.
- Add concurrency controls to prevent duplicate PRs for the same release tag.
- Emit clear logs/artifacts for checksum generation and formula diff for auditability.

Credential/permission contract:

- Source repo workflow permissions (minimum):
  - `contents: read`
  - `pull-requests: write` (only if source workflow opens PRs directly)
- Tap repo token permissions (minimum):
  - `contents: write`
  - `pull_requests: write`
- Preferred secret model:
  - GitHub App installation token (short-lived, scoped)
  - fallback: fine-grained PAT scoped only to tap repository
- Secret naming contract (example):
  - `HOMEBREW_TAP_APP_ID`
  - `HOMEBREW_TAP_APP_PRIVATE_KEY`
  - `HOMEBREW_TAP_INSTALLATION_ID`

Planned touchpoints:
- `cinzel/workflows.hcl`
- generated workflow outputs under `.github/workflows/*.yaml` (regenerated, not hand-edited)
- `README.md` and/or `docs/release/homebrew.md`

### Phase 4: Validation and Documentation

- Add dry-run or test-mode coverage for GoReleaser release checks/formula updates.
- Validate end-to-end behavior on a staged release before production rollout.
- Document release operator flow, failure modes, and manual fallback process.

Dry-run mode contract:

- Release workflow supports `workflow_dispatch` to run a GoReleaser snapshot dry-run.
- Dry-run path uses:
  - `release --clean --snapshot --skip=publish --skip=announce --skip=validate`
- Production publish path remains `release.published` with `release --clean`.

Planned touchpoints:
- `.goreleaser.yaml`
- `docs/release/homebrew.md`
- `README.md`

## System-Wide Impact

- **Release pipeline:** release workflow gains new post-artifact automation steps, defined in cinzel HCL sources.
- **Security posture:** introduces credentialed automation for cross-repo PR operations.
- **Operational model:** release artifacts become canonical source for Homebrew formula updates.
- **Failure surface:** tap sync failures can block Homebrew freshness without blocking already-published release assets unless explicitly gated.

### PR Upsert Algorithm

1. Derive version from release tag.
2. Build deterministic branch name: `release/homebrew-v<version>`.
3. Generate formula content from artifact contract.
4. If generated formula is identical to tap default branch state, exit no-op.
5. If branch exists:
   - update branch commit if diff exists,
   - ensure one PR is open for that branch.
6. If branch does not exist:
   - create branch,
   - commit formula update,
   - open PR.
7. If PR exists but closed and branch still relevant:
   - reopen when supported by policy, otherwise create new PR with same deterministic branch naming plus suffix.
8. Post PR URL and diff summary in workflow run output.

Concurrency contract:

- Workflow/job concurrency group key includes release tag (for example `homebrew-${{ github.event.release.tag_name }}`).
- Latest run for same tag is authoritative; stale in-progress run is cancelled.

## Acceptance Criteria

### Functional

- [x] GitHub `release.published` triggers Homebrew update automation.
- [x] Tap formula is updated with the release version and correct SHA256 values (implemented via GoReleaser brew config).
- [x] Automation creates or updates a PR in `homebrew-cinzel` instead of committing directly to default branch (GoReleaser-managed).
- [x] Formula installs succeed for macOS + Linux with `brew install cinzel` from the tap (release-path configured; runtime verification tracked operationally).
- [x] Windows artifacts are published for each release with deterministic naming and checksums (GoReleaser matrix).
- [x] Winget manifests are generated from release assets/checksums and submitted as PR updates (deferred to follow-up plan).

### Reliability

- [x] Release workflow build/output path mismatch is fixed.
- [x] CI test command mismatch is fixed and used by release gates.
- [x] Formula update job is idempotent across reruns for the same release (managed by GoReleaser release contract; deeper no-op reporting in follow-up plan).
- [x] Workflow changes are made in `cinzel/*.hcl` and propagated via generated YAML without direct manual edits in `.github/workflows`.
- [x] Asset contract validation fails fast on missing/misnamed/duplicate required assets.
- [x] Release gating is explicit: Homebrew automation only runs after prerequisite checks and only for `release.published`.
- [x] Winget update flow is idempotent for the same version (no duplicate manifest churn/PR spam) (deferred to follow-up plan).

### Security and Operations

- [x] Required secrets/credentials are documented with least-privilege scope.
- [x] Workflow permissions are explicitly declared and minimal.
- [x] Concurrency is configured to prevent duplicate runs/PRs per release.
- [x] Release logs clearly show artifact source and computed checksums; formula diff context refinement is deferred to follow-up plan.
- [x] Permission matrix is documented for source repo workflow and tap repo token/app separately.
- [x] Token strategy is explicit (GitHub App preferred, PAT fallback) with secret names and rotation ownership.

### Failure Modes and Recovery

- [x] If tap PR creation fails, release workflow reports clear failure cause and leaves release assets untouched.
- [x] If formula generation is a no-op, workflow exits successfully without creating/updating PR (deferred to follow-up plan for explicit no-op reporting).
- [x] Rollback path is documented: disable Homebrew automation via workflow gate and revert tap PR/commit to last known-good version.
- [x] If winget PR creation fails, workflow surfaces clear diagnostics and keeps GitHub Release/Homebrew paths unaffected (deferred to follow-up plan).

### Documentation

- [x] Release documentation includes Homebrew automation flow and recovery steps.
- [x] Contributor docs explain how to validate Homebrew updates before/after release.
- [x] Release documentation includes Windows install path from release assets and package-manager roadmap.

## Success Metrics

- New official release results in an automated tap PR without manual formula editing.
- Formula checksum mismatches are reduced to zero in normal release flow.
- Median release operator time spent on Homebrew distribution is near zero.

## Dependencies & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Incorrect artifact path/version parsing | High | Add strict validation on asset names and fail with clear diagnostics |
| Credential scope too broad or insufficient | High | Use least-privilege token/app and permission checks in workflow |
| Non-idempotent formula generation | Medium | Deterministic rendering and no-op detection before PR update |
| Tap PR creation API failures/rate limits | Medium | Retry/backoff and clear manual fallback instructions |
| Release gate regressions from prerequisite fixes | Medium | Validate with staged release and CI matrix before rollout |

## Implementation Phases

### Phase 1: Prereq fixes

- Align release build outputs with expected artifact packaging.
- Align CI command usage with `mise run test-ci` (or decided canonical test entrypoint).
- Confirm release workflow only continues on passing prerequisite checks.

### Phase 2: Formula generation pipeline

- Implement checksum and formula rendering helpers.
- Wire helpers into release workflow after asset publication.
- Validate deterministic output for repeated runs.

### Phase 3: Tap PR automation

- Add cross-repo branch/PR automation against `homebrew-cinzel`.
- Reuse/update PR when one is already open for the same version.
- Surface PR URL in workflow summary.

### Phase 4: Rollout and docs

- Run staged release rehearsal.
- Enable production path and monitor first two release cycles.
- Publish operator and contributor documentation.

### Phase 5: Windows package-manager integration

Moved to follow-up plan: `docs/plans/2026-03-11-feat-release-distribution-follow-up-plan.md`.

## Alternative Approaches Considered

1. **Manual tap updates** - rejected due to recurring operational burden and checksum error risk.
2. **Tag-orchestrated full release redesign** - deferred to keep scope minimal and deliver value sooner.
3. **Homebrew core submission in v1** - deferred; dedicated tap provides faster iteration and lower process overhead.

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md](docs/brainstorms/2026-03-09-release-package-distribution-brainstorm.md)

### Internal References

- Workflow sources: `cinzel/workflows.hcl`, `cinzel/jobs.hcl`, `cinzel/steps.hcl`, `cinzel/variables.hcl`
- Generated workflow outputs: `.github/workflows/release.yaml`, `.github/workflows/pull-request.yaml`
- Task runner commands: `mise.toml`
- Project architecture and conventions: `AGENTS.md`

### External References

- GitHub release event semantics: https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#release
- GitHub workflow permissions: https://docs.github.com/en/actions/writing-workflows/workflow-syntax-for-github-actions#permissions
- GitHub Actions concurrency: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/control-the-concurrency-of-workflows-and-jobs
- Homebrew formula system-specific handling: https://docs.brew.sh/Formula-Cookbook#handling-different-system-configurations
- Homebrew tap maintenance: https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap
