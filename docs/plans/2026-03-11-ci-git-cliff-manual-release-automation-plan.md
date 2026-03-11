---
title: "ci: adopt git-cliff manual release automation"
type: ci
status: completed
date: 2026-03-11
---

# ci: adopt git-cliff manual release automation

## Overview

Replace changelog/release-note generation complexity with a straightforward git-cliff workflow and a manual GitHub release pipeline built from marketplace actions.

## Problem Statement / Motivation

Previous changelog flow required multiple custom scripts and version-prep steps that were easy to misuse. The release process needed a simpler operator path with fewer moving parts.

## Final Approach

1. Use `git-cliff` as the changelog source of truth.
2. Keep local changelog update as one command: `mise run changelog`.
3. Add a manual `workflow_dispatch` release workflow.
4. Use marketplace actions for version tagging, changelog generation, changelog commit, and GitHub release creation.
5. Keep release packaging in GoReleaser.

## Implementation

### Tooling

- Added `git-cliff` to `mise` tools.
- Added `cliff.toml` configuration (based on upstream template, adapted for `yldio/cinzel`).
- Removed changie config and changie-driven release prep tasks.

### Workflow source (HCL)

- Added `workflow "release"` (`workflow_dispatch`).
- Kept published-release packaging in `workflow "release_published"` (`release.published`).
- Added `job "manual-release"` with this sequence:
  - checkout full history
  - setup mise
  - run tests
  - bump/push tag
  - generate changelog via `orhun/git-cliff-action`
  - commit changelog via `stefanzweifel/git-auto-commit-action`
  - create release via `ncipollo/release-action` using generated content

### Generated workflows

- Regenerated `.github/workflows/release.yaml`.
- Added generated `.github/workflows/release-published.yaml`.

### Documentation and config alignment

- Updated changelog guidance in `CONTRIBUTING.md` and `CLAUDE.md`.
- Updated release docs to match current Homebrew target/repo references.
- Removed docker build/run tasks and deleted `build/Dockerfile`.

## Acceptance Criteria

- [x] Changelog update can be done with `mise run changelog`.
- [x] Manual release can be started via `workflow_dispatch`.
- [x] Manual release workflow uses marketplace actions instead of custom commit shell scripting.
- [x] GitHub release body is derived from git-cliff output.
- [x] Workflow source remains in `cinzel/*.hcl` and generated YAML stays in sync.
- [x] `actionlint`, `ghalint`, and `mise run test-ci` pass after migration.

## Sources & References

- `cinzel/workflows.hcl`
- `cinzel/jobs.hcl`
- `cinzel/steps.hcl`
- `.github/workflows/release.yaml`
- `.github/workflows/release-published.yaml`
- `mise.toml`
- `cliff.toml`
- `CONTRIBUTING.md`
- `CLAUDE.md`
