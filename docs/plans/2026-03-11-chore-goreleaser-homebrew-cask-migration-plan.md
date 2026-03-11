---
title: "chore: Migrate GoReleaser brews to homebrew_casks"
type: chore
status: completed
date: 2026-03-11
origin: user:/ce:plan
---

# chore: Migrate GoReleaser brews to homebrew_casks

## Overview

Migrate release packaging from deprecated GoReleaser `brews` to `homebrew_casks` without changing the current release trigger model, token model, or tap PR automation path.

## Current State

- `.goreleaser.yaml` still uses `brews` with PR automation to `yldio/homebrew-cinzel`.
- Release workflow already delegates publishing to GoReleaser and provides `HOMEBREW_TAP_GITHUB_TOKEN`.
- Workflow observability reports generated Ruby artifacts as "formula output".
- Workflow source of truth is `cinzel/*.hcl`; `.github/workflows/*.yaml` is generated output.

## Goals

- Remove deprecated GoReleaser Homebrew config usage.
- Preserve release behavior (`release.published`) and existing token/automation contracts.
- Preserve tap automation (branch/PR creation in `yldio/homebrew-cinzel`).
- Keep rollback simple if cask rollout introduces install regressions.

## Non-Goals

- Redesigning release orchestration.
- Introducing new package managers.
- Refactoring unrelated release workflow steps.

## Migration Plan

### Phase 1: Configuration Migration

1. Update `.goreleaser.yaml`:
- Replace top-level `brews` block with `homebrew_casks`.
- Keep existing tap repository owner/name and token interpolation.
- Keep metadata parity (`homepage`, `description`, `license`).
- Add `binaries: [cinzel]` explicitly for cask packaging.
- Set `directory: Casks` only if the tap repository uses that layout; otherwise rely on GoReleaser defaults.
- Add cask conflicts metadata for legacy formula name to avoid dual-install ambiguity.

2. Preserve action invocation contract:
- Keep current `goreleaser/goreleaser-action` step shape and release args.
- Keep `HOMEBREW_TAP_GITHUB_TOKEN` secret name unchanged.

### Phase 2: Tap Compatibility and Deprecation Bridge

1. In `yldio/homebrew-cinzel` tap:
- Ensure cask file path/layout matches what GoReleaser will update.
- Keep (or add) legacy formula as disabled/deprecated with replacement pointing to cask.
- Confirm tap still passes `brew audit` for cask output.

2. Installation continuity:
- Document transition from formula-first to cask-first install path.
- Validate user messaging in README/release docs so operators know expected command/output changes.

### Phase 3: Workflow Observability Updates

1. Update workflow source in `cinzel/steps.hcl`:
- Rename release summary section from "Formula output" to "Homebrew artifact output" (or "Cask output").
- Keep artifact discovery logic over `dist/*.rb` (works for generated cask Ruby files).

2. Regenerate workflow YAML:
- Regenerate `.github/workflows/release.yaml` from `cinzel/*.hcl` using cinzel workflow generation flow.
- Do not hand-edit generated YAML.

### Phase 4: Validation and Rollout

1. Pre-merge validation:
- Run GoReleaser config validation.
- Run release configuration checks and workflow linting.
- Confirm generated Homebrew Ruby artifact is cask-oriented and contains expected metadata.

2. Post-merge release validation:
- Cut a test release tag in a safe release window.
- Verify GoReleaser opens/updates PR in `yldio/homebrew-cinzel`.
- Verify tap PR contents are cask updates, not formula mutations.
- Verify install smoke test from tap on macOS.

## Risks and Mitigations

- Linux install expectations may change with cask migration.
  - Mitigation: explicitly document supported install paths and keep fallback formula deprecation bridge during transition.
- Tap repository layout mismatch (`Formula/` vs `Casks/`) can break PR updates.
  - Mitigation: inspect tap layout before merge and set `directory` explicitly when needed.
- Existing automation/observability language may become misleading.
  - Mitigation: update summary labels and docs in same PR as config migration.

## Acceptance Criteria

- [x] `.goreleaser.yaml` uses `homebrew_casks` and no longer uses `brews`.
- [x] Release workflow still runs via existing triggers and continues using `HOMEBREW_TAP_GITHUB_TOKEN`.
- [ ] A release run updates/open PR against `yldio/homebrew-cinzel` through GoReleaser automation.
- [x] Release summary reports generated Homebrew Ruby artifacts with cask-accurate wording.
- [x] Operator docs describe the formula-to-cask transition and rollback path.

## Rollback Plan

1. Revert `.goreleaser.yaml` Homebrew section to prior `brews` config.
2. Regenerate/revert workflow summary text if changed.
3. Close superseded tap PR and restore previous known-good formula PR path.
