---
title: "feat: Release distribution follow-up"
type: feat
status: planned
date: 2026-03-11
origin: docs/plans/2026-03-09-feat-release-package-distribution-plan.md
---

# feat: Release distribution follow-up

## Overview

Follow-up items intentionally split from the completed v1 release package distribution plan to keep delivery honest and incremental.

## Scope

1. Winget automation
- Generate winget manifests from release metadata/checksums.
- Create/update PRs against `winget-pkgs` (or fork + upstream PR flow).
- Ensure deterministic/idempotent reruns for same version.

2. No-op and diff observability
- Add explicit no-op detection/reporting for Homebrew formula updates.
- Publish formula diff context in release summaries.

3. Failure-mode hardening
- Make winget PR failure diagnostics explicit while keeping Homebrew/release paths isolated.

## Acceptance Criteria

- [ ] Winget manifests are generated from release assets/checksums and submitted via PR automation.
- [ ] Winget flow is idempotent for same version.
- [ ] Release workflow reports explicit no-op states for Homebrew updates.
- [ ] Release summary includes formula diff context when changes exist.
- [ ] Winget PR failures produce clear diagnostics without breaking already-published release assets.

## Notes

- Keep `cinzel/*.hcl` as workflow source of truth; regenerate `.github/workflows/*.yaml`.
- Maintain SHA-pinned actions and least-privilege permissions.
