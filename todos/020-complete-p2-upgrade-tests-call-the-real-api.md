---
status: complete
priority: p2
issue_id: "020"
tags: [code-review, testing]
dependencies: ["019"]
---

# Two upgrade tests called the live GitHub API

## Problem Statement

`TestUpgradeFileDryRun` and `TestUpgradeDirectoryNoHCL` built their resolver
with `NewGitHubResolver("")`, which is the real one. `go test` therefore
reached api.github.com on every run, CI included, and `NewGitHubResolver`
reads `GITHUB_TOKEN` from the environment, so where a token was set the
calls went out authenticated.

`TestUpgradeFileDryRun` also asserted nothing. It checked only that the file
was unchanged under `dryRun`, which is equally true of a run that resolved
nothing and of an `UpgradeFile` that returns immediately.

## Findings

Found while checking a `t.Skip` in the same file during 019.

The comments in the dry-run test stated the opposite of what happened:
"This will fail because there's no HTTP mock" and "API calls fail". The
calls succeeded. Running it against the real API resolved `actions/checkout`
to `v7.0.1` at `3d3c42e5aac5` from within the test suite.

The vacuity and the network call were each confirmed by mutation:

- `UpgradeFile` gutted to `return nil, nil`: the old test passed.
- All HTTP forced through a closed port: the old test passed in 0.00s,
  against 0.775s with the network available. A live round trip was in that
  difference, and the test could not tell the two apart.

`TestUpgradeFileIntegration` was a bare `t.Skip` deferring to "e2e tests".
There is no e2e or integration suite in the repository. The multi-action
behaviour it pointed at is covered by `version_rewrite_test.go`, added in
019, so the skip was removed rather than left citing coverage that does not
exist.

## Fix

Both tests take `stubUpgrader`, which was already declared in the file. The
dry-run test now asserts what a dry run is for: that the upgrade was found
and reported, and only the write declined.

## Verification

`TestUpgradeFileDryRun` was proven to fail under two mutations, neither of
which the old test caught:

- `UpgradeFile` returning `nil, nil` before doing anything
- the `!dryRun` guard dropped, so a dry run writes

With every HTTP call forced through a closed port the package passes, and no
test constructs `NewGitHubResolver` or `NewCachedResolver` any more.
