# Homebrew release automation

This document describes how `cinzel` release distribution is automated for Homebrew and what operators should do when a release run fails.

## Scope

- Source release trigger: GitHub `release.published`
- Homebrew target: `yldio/homebrew-cinzel` tap (PR-based updates)
- Supported Homebrew targets: macOS and Linux
- Windows distribution: release artifacts now, winget PR flow in a later phase

## Asset contract

The release is expected to publish exactly one archive per required target:

- `cinzel_<version>_darwin_amd64.tar.gz`
- `cinzel_<version>_darwin_arm64.tar.gz`
- `cinzel_<version>_linux_amd64.tar.gz`
- `cinzel_<version>_linux_arm64.tar.gz`
- `cinzel_<version>_windows_amd64.tar.gz`
- `cinzel_<version>_windows_arm64.tar.gz`

Checksums are generated from published release artifacts by GoReleaser into `dist/checksums.txt`.

## Credential model

Preferred model:

- GitHub App installation token scoped to `yldio/homebrew-cinzel`

Fallback model:

- Fine-grained PAT scoped only to `yldio/homebrew-cinzel`

Current workflow secret contract:

- `HOMEBREW_TAP_GITHUB_TOKEN`: token used by GoReleaser Homebrew cask integration

## Permission matrix

Source repository workflow permissions:

- Workflow: `contents: read`
- Release job: `contents: write`

Tap repository token permissions:

- `contents: write`
- `pull_requests: write`

## Release flow

1. Release workflow starts on `release.published`.
2. GoReleaser builds archives and computes checksums.
3. GoReleaser renders the Homebrew cask update.
4. GoReleaser pushes/update branch and creates or updates a PR in `yldio/homebrew-cinzel`.
5. Workflow summary publishes checksum and generated Homebrew Ruby artifact context.

Recommended pre-merge checks:

- `mise run cinzel github parse --directory ./cinzel --output-directory ./.github/workflows`
- `mise run test-ci`
- `mise x actionlint -- actionlint`
- `ghalint run`

## Failure modes and recovery

Tap PR creation failure:

- Symptom: GoReleaser step fails during Homebrew cask publish/PR update
- Effect: GitHub release artifacts stay published; tap cask is not updated
- Recovery: fix token/repo access and rerun the release workflow for the same tag

No-op Homebrew update:

- Symptom: rerun has no Homebrew Ruby artifact change
- Effect: no new PR churn should be created
- Recovery: none required

Checksum or artifact mismatch:

- Symptom: missing or mismatched archive/checksum output
- Effect: release workflow fails before Homebrew update is considered healthy
- Recovery: fix artifact generation contract and rerun

Emergency rollback:

1. Disable Homebrew automation path by guarding/removing release publish step in workflow source (`cinzel/*.hcl`) and regenerate workflow YAML.
2. Revert or close the tap PR for the bad version.
3. Ship a corrected patch release.

## Windows distribution note

Windows users should install from GitHub release artifacts until winget automation is enabled.

Planned package-manager path:

- Generate deterministic `winget` manifests from release assets/checksums
- Open/update PRs against `microsoft/winget-pkgs`
- Treat winget merge latency as asynchronous to GitHub release success
