---
status: complete
priority: p3
issue_id: "045"
tags: [code-review, assist, validation]
dependencies: []
---

# A directory named ..hidden is read as traversal

## Problem Statement

`validateRelativePath` refuses any cleaned path starting with the two
characters `..`, so `..hidden` and `...x` are rejected although neither
escapes the working directory.

## Findings

`internal/command/assist.go:585`.

## Recommended Action

Compare the first path element against `..` exactly, or check for a `../`
prefix and the bare `..` case.

## Acceptance Criteria

- [x] `..hidden` is accepted
- [x] `../x`, `..` and absolute paths are still refused

## Technical Details

`validateRelativePath` (`internal/command/assist.go`) now refuses a cleaned path
that is exactly `..` or begins with `..` followed by a separator, rather than
one whose first two characters are dots.

The check stays on the cleaned path rather than splitting it. `filepath.Clean`
resolves every `..` it can and leaves the rest at the front, so `a/../../b`
cleans to `../b` and is still caught by the prefix test. The absolute-path
branch is unchanged.

`filepath.Separator` is used rather than a literal slash, so a Windows
`..\x` is refused the same way.

## Work Log

### 2026-09-16

Verified through the CLI: `--output-directory ..hidden` used to be refused with
`path must not escape the project directory` and now passes validation, reaching
the cost-confirmation prompt. `--output-directory ../x` is still refused with
that message.

`TestValidateRelativePath` gained five cases: `..hidden`, `...x`, a `..hidden`
element in the middle of a path, bare `..`, and `../x`. The three accepting
cases were confirmed to fail with the fix removed (the middle one passes either
way, since Clean leaves it where it is — it is there to pin that behaviour).

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
