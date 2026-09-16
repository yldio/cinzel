---
status: complete
priority: p1
issue_id: "034"
tags: [code-review, correctness, github, gitlab, unparse]
dependencies: []
---

# Two input files with the same basename write to one output path

## Problem Statement

Unparse builds the output path as `outputDir/<name>.hcl` with no claim on the
name. A recursive run over `a/ci.yaml` and `b/ci.yaml` writes `out/ci.hcl`
twice; the first result is lost with no warning.

## Findings

`provider/github/github.go:176-204`. The parse path already guards this with
`claimFilename` in `provider/github/io_helpers.go`.

GitLab's unparse loop (`provider/gitlab/gitlab.go`) had the same defect and is
fixed here too.

## Recommended Action

Thread the same `claimFilename` set through the unparse loop.

## Technical Details

Diverged from the recommendation. `claimFilename` **errors** on a collision,
which is right for parse: the filename there is an authored attribute, so the
author can change it. On unparse the name comes from the input's position on
disk. Erroring would make a recursive run over a repo that has a `ci.yaml` in
two directories fail with nothing the user can do short of renaming their CI
files. The acceptance criterion also asks for distinct files, not an error.

- `internal/fsutil.UniqueOutputName` records a name and returns a `_2`, `_3`
  variant when it is taken. Case is folded for the comparison, as in
  `claimFilename`, because a name differing only in case is one file on macOS
  and Windows. The returned name keeps the caller's casing
- It sits in `fsutil` rather than `naming` because both providers need it and
  it is about output paths, not identifiers. `naming.UniqueIdentifierInSet`
  does not fold case and takes a pre-built set
- A file that is skipped claims nothing: the claim happens after the `nil`
  check, so a directory full of non-workflow YAML does not push names along

Out of scope, found while verifying: reparsing the two deduped files together
fails with `error in step 'echo': already defined`. Steps are keyed globally
across an input directory, so two workflows holding the same inline `run` do
not coexist. That is the same class as todo 030 and is left to it.

## Acceptance Criteria

- [x] Colliding basenames produce distinct output files
- [x] Single-file runs keep their current names
- [x] GitLab's unparse loop is guarded the same way
- [x] A skipped file does not consume a name

## Work Log

- 2026-09-16: Fixed for both providers. Chose dedup over the suggested error,
  reasoning above.
